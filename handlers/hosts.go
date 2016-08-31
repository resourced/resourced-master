package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/csrf"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/messagebus"
)

func GetHosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	currentUser := r.Context().Value("currentUser").(*dal.UserRow)

	currentCluster := r.Context().Value("currentCluster").(*dal.ClusterRow)

	dbs := r.Context().Value("dbs").(*config.DBConfig)

	query := r.URL.Query().Get("q")

	// -----------------------------------
	// Create channels to receive SQL rows
	// -----------------------------------
	hostsChan := make(chan *dal.HostRowsWithError)
	defer close(hostsChan)

	savedQueriesChan := make(chan *dal.SavedQueryRowsWithError)
	defer close(savedQueriesChan)

	metricsMapChan := make(chan *dal.MetricsMapWithError)
	defer close(metricsMapChan)

	// --------------------------
	// Fetch SQL rows in parallel
	// --------------------------
	go func(currentCluster *dal.ClusterRow, query string) {
		hostsWithError := &dal.HostRowsWithError{}
		hostsWithError.Hosts, hostsWithError.Error = dal.NewHost(dbs.GetHost(currentCluster.ID)).AllByClusterIDAndQuery(nil, currentCluster.ID, query)
		hostsChan <- hostsWithError
	}(currentCluster, query)

	go func(currentCluster *dal.ClusterRow) {
		savedQueriesWithError := &dal.SavedQueryRowsWithError{}
		savedQueriesWithError.SavedQueries, savedQueriesWithError.Error = dal.NewSavedQuery(dbs.Core).AllByClusterIDAndType(nil, currentCluster.ID, "hosts")
		savedQueriesChan <- savedQueriesWithError
	}(currentCluster)

	go func(currentCluster *dal.ClusterRow) {
		metricsMapWithError := &dal.MetricsMapWithError{}
		metricsMapWithError.MetricsMap, metricsMapWithError.Error = dal.NewMetric(dbs.Core).AllByClusterIDAsMap(nil, currentCluster.ID)
		metricsMapChan <- metricsMapWithError
	}(currentCluster)

	// -----------------------------------
	// Wait for channels to return results
	// -----------------------------------
	hostsWithError := <-hostsChan
	if hostsWithError.Error != nil && hostsWithError.Error.Error() != "sql: no rows in result set" {
		libhttp.HandleErrorHTML(w, hostsWithError.Error, 500)
		return
	}

	savedQueriesWithError := <-savedQueriesChan
	if savedQueriesWithError.Error != nil && savedQueriesWithError.Error.Error() != "sql: no rows in result set" {
		libhttp.HandleErrorHTML(w, savedQueriesWithError.Error, 500)
		return
	}

	metricsMapWithError := <-metricsMapChan
	if metricsMapWithError.Error != nil {
		libhttp.HandleErrorHTML(w, metricsMapWithError.Error, 500)
		return
	}

	accessToken, err := getAccessToken(w, r, "read")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	data := struct {
		CSRFToken      string
		Addr           string
		CurrentUser    *dal.UserRow
		AccessToken    *dal.AccessTokenRow
		Clusters       []*dal.ClusterRow
		CurrentCluster *dal.ClusterRow
		Hosts          []*dal.HostRow
		SavedQueries   []*dal.SavedQueryRow
		MetricsMap     map[string]int64
	}{
		csrf.Token(r),
		r.Context().Value("addr").(string),
		currentUser,
		accessToken,
		r.Context().Value("clusters").([]*dal.ClusterRow),
		currentCluster,
		hostsWithError.Hosts,
		savedQueriesWithError.SavedQueries,
		metricsMapWithError.MetricsMap,
	}

	var tmpl *template.Template

	currentUserPermission := currentCluster.GetLevelByUserID(currentUser.ID)
	if currentUserPermission == "read" {
		tmpl, err = template.ParseFiles("templates/dashboard.html.tmpl", "templates/hosts/list-readonly.html.tmpl")
	} else {
		tmpl, err = template.ParseFiles("templates/dashboard.html.tmpl", "templates/hosts/list.html.tmpl")
	}
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	tmpl.Execute(w, data)
}

func PostApiHosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbs := r.Context().Value("dbs").(*config.DBConfig)

	accessTokenRow := r.Context().Value("accessToken").(*dal.AccessTokenRow)

	bus := r.Context().Value("bus").(*messagebus.MessageBus)

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	hostRow, err := dal.NewHost(dbs.GetHost(accessTokenRow.ClusterID)).CreateOrUpdate(nil, accessTokenRow, dataJson)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	// Asynchronously write timeseries data
	go func() {
		metricsMap, err := dal.NewMetric(dbs.Core).AllByClusterIDAsMap(nil, hostRow.ClusterID)
		if err != nil {
			logrus.Error(err)
			libhttp.HandleErrorJson(w, err)
			return
		}

		clusterRow, err := dal.NewCluster(dbs.Core).GetByID(nil, hostRow.ClusterID)
		if err != nil {
			logrus.Error(err)
			return
		}

		tsMetricsDeletedFrom := clusterRow.GetDeletedFromUNIXTimestampForInsert("ts_metrics")
		tsMetricsAggr15mDeletedFrom := clusterRow.GetDeletedFromUNIXTimestampForInsert("ts_metrics_aggr_15m")

		// Create ts_metrics row
		err = dal.NewTSMetric(dbs.GetTSMetric(hostRow.ClusterID)).CreateByHostRow(nil, hostRow, metricsMap, tsMetricsDeletedFrom)
		if err != nil {
			logrus.Error(err)
			return
		}

		go func() {
			selectAggrRows, err := dal.NewTSMetric(dbs.GetTSMetric(hostRow.ClusterID)).AggregateEveryXMinutes(nil, hostRow.ClusterID, 15)
			if err != nil {
				logrus.Error(err)
				return
			}

			// Create ts_metrics_aggr_15m rows.
			err = dal.NewTSMetricAggr15m(dbs.GetTSMetricAggr15m(hostRow.ClusterID)).CreateByHostRow(nil, hostRow, metricsMap, selectAggrRows, tsMetricsAggr15mDeletedFrom)
			if err != nil {
				logrus.Error(err)
				return
			}
		}()

		go func() {
			selectAggrRows, err := dal.NewTSMetric(dbs.GetTSMetric(hostRow.ClusterID)).AggregateEveryXMinutesPerHost(nil, hostRow.ClusterID, 15)
			if err != nil {
				logrus.Error(err)
				return
			}

			// Create ts_metrics_aggr_15m rows per host.
			err = dal.NewTSMetricAggr15m(dbs.GetTSMetricAggr15m(hostRow.ClusterID)).CreateByHostRowPerHost(nil, hostRow, metricsMap, selectAggrRows, tsMetricsAggr15mDeletedFrom)
			if err != nil {
				logrus.Error(err)
				return
			}
		}()

		go func() {
			// Publish evey graphed metric to message bus.
			bus.PublishMetricsByHostRow(hostRow, metricsMap)
		}()
	}()

	hostRowJson, err := json.Marshal(hostRow)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(hostRowJson)
}

func GetApiHosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbs := r.Context().Value("dbs").(*config.DBConfig)

	accessTokenRow := r.Context().Value("accessToken").(*dal.AccessTokenRow)

	query := r.URL.Query().Get("q")
	count := r.URL.Query().Get("count")

	hosts, err := dal.NewHost(dbs.GetHost(accessTokenRow.ClusterID)).AllByClusterIDAndQuery(nil, accessTokenRow.ClusterID, query)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	if count == "true" {
		w.Write([]byte(fmt.Sprintf("%v", len(hosts))))

	} else {
		hostRowsJson, err := json.Marshal(hosts)
		if err != nil {
			libhttp.HandleErrorJson(w, err)
			return
		}

		w.Write(hostRowsJson)
	}
}
