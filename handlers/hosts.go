package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/csrf"

	"github.com/resourced/resourced-master/contexthelper"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/models/pg"
	"github.com/resourced/resourced-master/models/shims"
)

func GetHosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	currentUser := r.Context().Value("currentUser").(*pg.UserRow)

	currentCluster := r.Context().Value("currentCluster").(*pg.ClusterRow)

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
	hostsChan := make(chan *pg.HostRowsWithError)
	defer close(hostsChan)

	savedQueriesChan := make(chan *pg.SavedQueryRowsWithError)
	defer close(savedQueriesChan)

	// --------------------------
	// Fetch SQL rows in parallel
	// --------------------------
	go func(currentCluster *pg.ClusterRow, query string) {
		hostsWithError := &pg.HostRowsWithError{}
		hostsWithError.Hosts, hostsWithError.Error = pg.NewHost(r.Context(), currentCluster.ID).AllCompactByClusterIDQueryAndUpdatedInterval(nil, currentCluster.ID, query, interval)
		hostsChan <- hostsWithError
	}(currentCluster, query)

	go func(currentCluster *pg.ClusterRow) {
		savedQueriesWithError := &pg.SavedQueryRowsWithError{}
		savedQueriesWithError.SavedQueries, savedQueriesWithError.Error = pg.NewSavedQuery(r.Context()).AllByClusterIDAndType(nil, currentCluster.ID, "hosts")
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
		CurrentUser    *pg.UserRow
		AccessToken    *pg.AccessTokenRow
		Clusters       []*pg.ClusterRow
		CurrentCluster *pg.ClusterRow
		Hosts          []*pg.HostRow
		SavedQueries   []*pg.SavedQueryRow
	}{
		csrf.Token(r),
		r.Context().Value("Addr").(string),
		currentUser,
		accessToken,
		r.Context().Value("clusters").([]*pg.ClusterRow),
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

	currentUser := r.Context().Value("currentUser").(*pg.UserRow)

	currentCluster := r.Context().Value("currentCluster").(*pg.ClusterRow)

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	accessToken, err := getAccessToken(w, r, "read")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	host, err := pg.NewHost(r.Context(), currentCluster.ID).GetByID(nil, id)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	// -----------------------------------
	// Create channels to receive SQL rows
	// -----------------------------------
	savedQueriesChan := make(chan *pg.SavedQueryRowsWithError)
	defer close(savedQueriesChan)

	metricsMapChan := make(chan *pg.MetricsMapWithError)
	defer close(metricsMapChan)

	// --------------------------
	// Fetch SQL rows in parallel
	// --------------------------
	go func(currentCluster *pg.ClusterRow) {
		savedQueriesWithError := &pg.SavedQueryRowsWithError{}
		savedQueriesWithError.SavedQueries, savedQueriesWithError.Error = pg.NewSavedQuery(r.Context()).AllByClusterIDAndType(nil, currentCluster.ID, "hosts")
		savedQueriesChan <- savedQueriesWithError
	}(currentCluster)

	go func(currentCluster *pg.ClusterRow) {
		metricsMapWithError := &pg.MetricsMapWithError{}
		metricsMapWithError.MetricsMap, metricsMapWithError.Error = pg.NewMetric(r.Context()).AllByClusterIDAsMap(nil, currentCluster.ID)
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
		CurrentUser    *pg.UserRow
		AccessToken    *pg.AccessTokenRow
		Clusters       []*pg.ClusterRow
		CurrentCluster *pg.ClusterRow
		Host           *pg.HostRow
		SavedQueries   []*pg.SavedQueryRow
		MetricsMap     map[string]int64
	}{
		csrf.Token(r),
		r.Context().Value("Addr").(string),
		currentUser,
		accessToken,
		r.Context().Value("clusters").([]*pg.ClusterRow),
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

	currentCluster := r.Context().Value("currentCluster").(*pg.ClusterRow)

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	masterTagsKVs := strings.Split(r.FormValue("MasterTags"), "\n")

	masterTags := make(map[string]interface{})

	for _, masterTagsKV := range masterTagsKVs {
		masterTagsKVSlice := strings.Split(masterTagsKV, ":")
		if len(masterTagsKVSlice) >= 2 {
			tagKey := strings.Replace(strings.TrimSpace(masterTagsKVSlice[0]), " ", "-", -1)
			tagValueString := strings.TrimSpace(masterTagsKVSlice[1])

			tagValueFloat, err := strconv.ParseFloat(tagValueString, 64)
			if err == nil {
				masterTags[strings.TrimSpace(tagKey)] = tagValueFloat
			} else {
				masterTags[strings.TrimSpace(tagKey)] = tagValueString
			}
		}
	}

	err = pg.NewHost(r.Context(), currentCluster.ID).UpdateMasterTagsByID(nil, id, masterTags)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

func PostApiHosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accessTokenRow := r.Context().Value("accessToken").(*pg.AccessTokenRow)

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

	hostRow, err := pg.NewHost(r.Context(), accessTokenRow.ClusterID).CreateOrUpdate(nil, accessTokenRow, dataJson)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	// Asynchronously write timeseries data
	go func() {
		metricsMap, err := pg.NewMetric(r.Context()).AllByClusterIDAsMap(nil, hostRow.ClusterID)
		if err != nil {
			errLogger.WithFields(logrus.Fields{
				"Error": err.Error(),
			}).Error("Failed to get map of metrics by cluster id")

			libhttp.HandleErrorJson(w, err)
			return
		}

		clusterRow, err := pg.NewCluster(r.Context()).GetByID(nil, hostRow.ClusterID)
		if err != nil {
			errLogger.WithFields(logrus.Fields{
				"Error": err.Error(),
			}).Error("Failed to get cluster by id")
			return
		}

		// Create ts_metrics row
		shimsTSMetric := shims.NewTSMetric(r.Context(), hostRow.ClusterID)

		err = shimsTSMetric.CreateByHostRow(hostRow, metricsMap, clusterRow.GetDeletedFromUNIXTimestampForInsert("ts_metrics"), clusterRow.GetTTLDurationForInsert("ts_metrics"))
		if err != nil {
			errLogger.Error(err)
			return
		}

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

	accessTokenRow := r.Context().Value("accessToken").(*pg.AccessTokenRow)

	query := r.URL.Query().Get("q")
	count := r.URL.Query().Get("count")
	interval := strings.TrimSpace(r.URL.Query().Get("interval"))

	if interval == "" {
		interval = "1h"
	}

	hosts, err := pg.NewHost(r.Context(), accessTokenRow.ClusterID).AllCompactByClusterIDQueryAndUpdatedInterval(nil, accessTokenRow.ClusterID, query, interval)
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
