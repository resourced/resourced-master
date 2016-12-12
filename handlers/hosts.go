package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/csrf"
	"github.com/pressly/chi"

	"github.com/resourced/resourced-master/contexthelper"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/models/cassandra"
)

func GetHosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	currentUser := r.Context().Value("currentUser").(*cassandra.UserRow)

	currentCluster := r.Context().Value("currentCluster").(*cassandra.ClusterRow)

	query := r.URL.Query().Get("q")

	interval := strings.TrimSpace(r.URL.Query().Get("interval"))
	if interval == "" {
		interval = "1h"
	}

	accessToken, err := getAccessToken(w, r, "read")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	// -----------------------------------
	// Create channels to receive SQL rows
	// -----------------------------------
	hostsChan := make(chan *cassandra.HostRowsWithError)
	defer close(hostsChan)

	savedQueriesChan := make(chan *cassandra.SavedQueryRowsWithError)
	defer close(savedQueriesChan)

	// --------------------------
	// Fetch SQL rows in parallel
	// --------------------------
	go func(currentCluster *cassandra.ClusterRow, query string) {
		hostsWithError := &cassandra.HostRowsWithError{}
		hostsWithError.Hosts, hostsWithError.Error = cassandra.NewHost(r.Context()).AllCompactByClusterIDQueryAndUpdatedInterval(currentCluster.ID, query, interval)
		hostsChan <- hostsWithError
	}(currentCluster, query)

	go func(currentCluster *cassandra.ClusterRow) {
		savedQueriesWithError := &cassandra.SavedQueryRowsWithError{}
		savedQueriesWithError.SavedQueries, savedQueriesWithError.Error = cassandra.NewSavedQuery(r.Context()).AllByClusterIDAndType(currentCluster.ID, "hosts")
		savedQueriesChan <- savedQueriesWithError
	}(currentCluster)

	// -----------------------------------
	// Wait for channels to return results
	// -----------------------------------
	hasError := false

	hostsWithError := <-hostsChan
	if hostsWithError.Error != nil && hostsWithError.Error.Error() != "sql: no rows in result set" {
		libhttp.HandleErrorHTML(w, hostsWithError.Error, 500)
		hasError = true
	}

	savedQueriesWithError := <-savedQueriesChan
	if savedQueriesWithError.Error != nil && savedQueriesWithError.Error.Error() != "sql: no rows in result set" {
		libhttp.HandleErrorHTML(w, savedQueriesWithError.Error, 500)
		hasError = true
	}

	if hasError {
		return
	}

	data := struct {
		CSRFToken      string
		Addr           string
		CurrentUser    *cassandra.UserRow
		AccessToken    *cassandra.AccessTokenRow
		Clusters       []*cassandra.ClusterRow
		CurrentCluster *cassandra.ClusterRow
		Hosts          []*cassandra.HostRow
		SavedQueries   []*cassandra.SavedQueryRow
	}{
		csrf.Token(r),
		r.Context().Value("Addr").(string),
		currentUser,
		accessToken,
		r.Context().Value("clusters").([]*cassandra.ClusterRow),
		currentCluster,
		hostsWithError.Hosts,
		savedQueriesWithError.SavedQueries,
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

func GetHostsID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	currentUser := r.Context().Value("currentUser").(*cassandra.UserRow)

	currentCluster := r.Context().Value("currentCluster").(*cassandra.ClusterRow)

	id := chi.URLParam(r, "id")

	accessToken, err := getAccessToken(w, r, "read")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	host, err := cassandra.NewHost(r.Context()).GetByID(id)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	// -----------------------------------
	// Create channels to receive SQL rows
	// -----------------------------------
	savedQueriesChan := make(chan *cassandra.SavedQueryRowsWithError)
	defer close(savedQueriesChan)

	metricsMapChan := make(chan *cassandra.MetricsMapWithError)
	defer close(metricsMapChan)

	// --------------------------
	// Fetch SQL rows in parallel
	// --------------------------
	go func(currentCluster *cassandra.ClusterRow) {
		savedQueriesWithError := &cassandra.SavedQueryRowsWithError{}
		savedQueriesWithError.SavedQueries, savedQueriesWithError.Error = cassandra.NewSavedQuery(r.Context()).AllByClusterIDAndType(currentCluster.ID, "hosts")
		savedQueriesChan <- savedQueriesWithError
	}(currentCluster)

	go func(currentCluster *cassandra.ClusterRow) {
		metricsMapWithError := &cassandra.MetricsMapWithError{}
		metricsMapWithError.MetricsMap, metricsMapWithError.Error = cassandra.NewMetric(r.Context()).AllByClusterIDAsMap(currentCluster.ID)
		metricsMapChan <- metricsMapWithError
	}(currentCluster)

	// -----------------------------------
	// Wait for channels to return results
	// -----------------------------------
	hasError := false

	savedQueriesWithError := <-savedQueriesChan
	if savedQueriesWithError.Error != nil && savedQueriesWithError.Error.Error() != "sql: no rows in result set" {
		libhttp.HandleErrorHTML(w, savedQueriesWithError.Error, 500)
		hasError = true
	}

	metricsMapWithError := <-metricsMapChan
	if metricsMapWithError.Error != nil {
		libhttp.HandleErrorHTML(w, metricsMapWithError.Error, 500)
		hasError = true
	}

	if hasError {
		return
	}

	data := struct {
		CSRFToken      string
		Addr           string
		CurrentUser    *cassandra.UserRow
		AccessToken    *cassandra.AccessTokenRow
		Clusters       []*cassandra.ClusterRow
		CurrentCluster *cassandra.ClusterRow
		Host           *cassandra.HostRow
		SavedQueries   []*cassandra.SavedQueryRow
		MetricsMap     map[string]int64
	}{
		csrf.Token(r),
		r.Context().Value("Addr").(string),
		currentUser,
		accessToken,
		r.Context().Value("clusters").([]*cassandra.ClusterRow),
		currentCluster,
		host,
		savedQueriesWithError.SavedQueries,
		metricsMapWithError.MetricsMap,
	}

	var tmpl *template.Template

	currentUserPermission := currentCluster.GetLevelByUserID(currentUser.ID)
	if currentUserPermission == "read" {
		tmpl, err = template.ParseFiles("templates/dashboard.html.tmpl", "templates/hosts/each-readonly.html.tmpl")
	} else {
		tmpl, err = template.ParseFiles("templates/dashboard.html.tmpl", "templates/hosts/each.html.tmpl")
	}
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	tmpl.Execute(w, data)
}

func PostHostsIDMasterTags(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	id := chi.URLParam(r, "id")

	masterTagsKVs := strings.Split(r.FormValue("MasterTags"), "\n")

	masterTags := make(map[string]string)

	for _, masterTagsKV := range masterTagsKVs {
		masterTagsKVSlice := strings.Split(masterTagsKV, ":")
		if len(masterTagsKVSlice) >= 2 {
			tagKey := strings.Replace(strings.TrimSpace(masterTagsKVSlice[0]), " ", "-", -1)
			tagValueString := strings.TrimSpace(masterTagsKVSlice[1])
			masterTags[strings.TrimSpace(tagKey)] = tagValueString
		}
	}

	err := cassandra.NewHost(r.Context()).UpdateMasterTagsByID(id, masterTags)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

func PostApiHosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accessTokenRow := r.Context().Value("accessToken").(*cassandra.AccessTokenRow)

	bus, err := contexthelper.GetMessageBus(r.Context())
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	errLogger, err := contexthelper.GetLogger(r.Context(), "ErrLogger")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	hostRow, err := cassandra.NewHost(r.Context()).CreateOrUpdate(accessTokenRow, dataJson)
	if err != nil {
		println("1")
		errLogger.WithFields(logrus.Fields{
			"Error": err.Error(),
		}).Error("Failed to store Host data in DB")
		libhttp.HandleErrorJson(w, err)
		return
	}

	_, err = cassandra.NewHostData(r.Context()).CreateOrUpdate(accessTokenRow, dataJson)
	if err != nil {
		println("2")
		errLogger.WithFields(logrus.Fields{
			"Error": err.Error(),
		}).Error("Failed to store Host data in DB")
		libhttp.HandleErrorJson(w, err)
		return
	}

	// Asynchronously write timeseries data
	go func() {
		metricsMap, err := cassandra.NewMetric(r.Context()).AllByClusterIDAsMap(hostRow.ClusterID)
		if err != nil {
			errLogger.WithFields(logrus.Fields{
				"Error": err.Error(),
			}).Error("Failed to get map of metrics by cluster id")
			return
		}

		clusterRow, err := cassandra.NewCluster(r.Context()).GetByID(hostRow.ClusterID)
		if err != nil {
			errLogger.WithFields(logrus.Fields{
				"Error": err.Error(),
			}).Error("Failed to get cluster by id")
			return
		}

		err = cassandra.NewTSMetric(r.Context()).CreateByHostRow(hostRow, metricsMap, clusterRow.GetTTLDurationForInsert("ts_metrics"))
		if err != nil {
			errLogger.Error(err)
			return
		}

		go func() {
			// Publish evey graphed metric to message bus.
			hostData, err := cassandra.NewHostData(r.Context()).AllByID(hostRow.ID)
			if err == nil {
				bus.PublishMetricsByHostRow(hostRow, hostData, metricsMap)
			}
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

	accessTokenRow := r.Context().Value("accessToken").(*cassandra.AccessTokenRow)

	query := r.URL.Query().Get("q")
	count := r.URL.Query().Get("count")
	interval := strings.TrimSpace(r.URL.Query().Get("interval"))

	if interval == "" {
		interval = "1h"
	}

	hosts, err := cassandra.NewHost(r.Context()).AllCompactByClusterIDQueryAndUpdatedInterval(accessTokenRow.ClusterID, query, interval)
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

func GetApiHostsID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var hostRow *cassandra.HostRow

	id := chi.URLParam(r, "id")

	hostRow, err := cassandra.NewHost(r.Context()).GetByID(id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	hostRowJSON, err := json.Marshal(hostRow)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(hostRowJSON)
}

func PutApiHostsIDMasterTags(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	id := chi.URLParam(r, "id")

	masterTags := make(map[string]string)

	err = json.Unmarshal(dataJson, &masterTags)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	err = cassandra.NewHost(r.Context()).UpdateMasterTagsByID(id, masterTags)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write([]byte(`{"Message": "Success"}`))
}
