package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/context"
	"github.com/gorilla/csrf"
	"github.com/jmoiron/sqlx"

	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func GetHosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	currentUser := context.Get(r, "currentUser").(*dal.UserRow)

	currentCluster := context.Get(r, "currentCluster").(*dal.ClusterRow)

	db := context.Get(r, "db.Core").(*sqlx.DB)

	query := r.URL.Query().Get("q")

	// -----------------------------------
	// Create channels to receive SQL rows
	// -----------------------------------
	hostsChan := make(chan *dal.HostRowsWithError)
	defer close(hostsChan)

	savedQueriesChan := make(chan *dal.SavedQueryRowsWithError)
	defer close(savedQueriesChan)

	accessTokenChan := make(chan *dal.AccessTokenRowWithError)
	defer close(accessTokenChan)

	metricsMapChan := make(chan *dal.MetricsMapWithError)
	defer close(metricsMapChan)

	// --------------------------
	// Fetch SQL rows in parallel
	// --------------------------
	go func(currentCluster *dal.ClusterRow, query string) {
		hostsWithError := &dal.HostRowsWithError{}
		hostsWithError.Hosts, hostsWithError.Error = dal.NewHost(db).AllByClusterIDAndQuery(nil, currentCluster.ID, query)
		hostsChan <- hostsWithError
	}(currentCluster, query)

	go func(currentCluster *dal.ClusterRow) {
		savedQueriesWithError := &dal.SavedQueryRowsWithError{}
		savedQueriesWithError.SavedQueries, savedQueriesWithError.Error = dal.NewSavedQuery(db).AllByClusterID(nil, currentCluster.ID)
		savedQueriesChan <- savedQueriesWithError
	}(currentCluster)

	go func(currentUser *dal.UserRow) {
		accessTokenWithError := &dal.AccessTokenRowWithError{}
		accessTokenWithError.AccessToken, accessTokenWithError.Error = dal.NewAccessToken(db).GetByUserID(nil, currentUser.ID)
		accessTokenChan <- accessTokenWithError
	}(currentUser)

	go func(currentCluster *dal.ClusterRow) {
		metricsMapWithError := &dal.MetricsMapWithError{}
		metricsMapWithError.MetricsMap, metricsMapWithError.Error = dal.NewMetric(db).AllByClusterIDAsMap(nil, currentCluster.ID)
		metricsMapChan <- metricsMapWithError
	}(currentCluster)

	// -----------------------------------
	// Wait for channels to return results
	// -----------------------------------
	hostsWithError := <-hostsChan
	if hostsWithError.Error != nil && hostsWithError.Error.Error() != "sql: no rows in result set" {
		libhttp.HandleErrorJson(w, hostsWithError.Error)
		return
	}

	savedQueriesWithError := <-savedQueriesChan
	if savedQueriesWithError.Error != nil && savedQueriesWithError.Error.Error() != "sql: no rows in result set" {
		libhttp.HandleErrorJson(w, savedQueriesWithError.Error)
		return
	}

	accessTokenWithError := <-accessTokenChan
	if accessTokenWithError.Error != nil {
		libhttp.HandleErrorJson(w, accessTokenWithError.Error)
		return
	}

	metricsMapWithError := <-metricsMapChan
	if metricsMapWithError.Error != nil {
		libhttp.HandleErrorJson(w, metricsMapWithError.Error)
		return
	}

	data := struct {
		CSRFToken          string
		Addr               string
		CurrentUser        *dal.UserRow
		AccessToken        *dal.AccessTokenRow
		Clusters           []*dal.ClusterRow
		CurrentClusterJson string
		Hosts              []*dal.HostRow
		SavedQueries       []*dal.SavedQueryRow
		MetricsMap         map[string]int64
	}{
		csrf.Token(r),
		context.Get(r, "addr").(string),
		currentUser,
		accessTokenWithError.AccessToken,
		context.Get(r, "clusters").([]*dal.ClusterRow),
		string(context.Get(r, "currentClusterJson").([]byte)),
		hostsWithError.Hosts,
		savedQueriesWithError.SavedQueries,
		metricsMapWithError.MetricsMap,
	}

	tmpl, err := template.ParseFiles("templates/dashboard.html.tmpl", "templates/hosts/list.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tmpl.Execute(w, data)
}

func PostApiHosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := context.Get(r, "db.Core").(*sqlx.DB)

	accessTokenRow := context.Get(r, "accessTokenRow").(*dal.AccessTokenRow)

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	hostRow, err := dal.NewHost(db).CreateOrUpdate(nil, accessTokenRow, dataJson)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	metricsMap, err := dal.NewMetric(db).AllByClusterIDAsMap(nil, hostRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tsMetricDB := context.Get(r, "db.TSMetric").(*sqlx.DB)
	// Asynchronously write time series data to ts_metrics
	go func() {
		err := dal.NewTSMetric(tsMetricDB).CreateByHostRow(nil, hostRow, metricsMap)
		if err != nil {
			logrus.Error(err)
		}
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

	db := context.Get(r, "db.Core").(*sqlx.DB)

	accessTokenRow := context.Get(r, "accessTokenRow").(*dal.AccessTokenRow)

	query := r.URL.Query().Get("q")
	count := r.URL.Query().Get("count")

	hosts, err := dal.NewHost(db).AllByClusterIDAndQuery(nil, accessTokenRow.ClusterID, query)
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
