package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/csrf"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func GetGraphs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	currentUser := context.Get(r, "currentUser").(*dal.UserRow)

	currentCluster := context.Get(r, "currentCluster").(*dal.ClusterRow)

	dbs := context.Get(r, "dbs").(*config.DBConfig)

	graphs, err := dal.NewGraph(dbs.Core).AllByClusterID(nil, currentCluster.ID)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	data := struct {
		CSRFToken      string
		Addr           string
		CurrentUser    *dal.UserRow
		Clusters       []*dal.ClusterRow
		CurrentCluster *dal.ClusterRow
		Graphs         []*dal.GraphRow
	}{
		csrf.Token(r),
		context.Get(r, "addr").(string),
		currentUser,
		context.Get(r, "clusters").([]*dal.ClusterRow),
		currentCluster,
		graphs,
	}

	var tmpl *template.Template

	currentUserPermission := currentCluster.GetLevelByUserID(currentUser.ID)
	if currentUserPermission == "read" {
		tmpl, err = template.ParseFiles("templates/dashboard.html.tmpl", "templates/graphs/list-readonly.html.tmpl")
	} else {
		tmpl, err = template.ParseFiles("templates/dashboard.html.tmpl", "templates/graphs/list.html.tmpl")
	}
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	tmpl.Execute(w, data)
}

func PostGraphs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	currentCluster := context.Get(r, "currentCluster").(*dal.ClusterRow)

	name := r.FormValue("Name")
	description := r.FormValue("Description")
	range_ := r.FormValue("Range")

	dbs := context.Get(r, "dbs").(*config.DBConfig)

	data := make(map[string]interface{})
	data["name"] = name
	data["description"] = description
	data["range"] = range_

	_, err := dal.NewGraph(dbs.Core).Create(nil, currentCluster.ID, data)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/graphs", 301)
}

func GetPostPutDeleteGraphsID(w http.ResponseWriter, r *http.Request) {
	method := r.FormValue("_method")

	if method == "" || method == "get" {
		GetGraphsID(w, r)
	} else if method == "post" || method == "put" {
		PutGraphsID(w, r)
	} else if method == "delete" {
		DeleteGraphsID(w, r)
	}
}

func GetGraphsID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	dbs := context.Get(r, "dbs").(*config.DBConfig)

	currentUser := context.Get(r, "currentUser").(*dal.UserRow)

	currentCluster := context.Get(r, "currentCluster").(*dal.ClusterRow)

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	// -----------------------------------
	// Create channels to receive SQL rows
	// -----------------------------------
	metricsChan := make(chan *dal.MetricRowsWithError)
	defer close(metricsChan)

	graphsChan := make(chan *dal.GraphRowsWithError)
	defer close(graphsChan)

	// --------------------------
	// Fetch SQL rows in parallel
	// --------------------------
	go func(currentCluster *dal.ClusterRow, id int64) {
		graphsWithError := &dal.GraphRowsWithError{}
		graphsWithError.Graphs, graphsWithError.Error = dal.NewGraph(dbs.Core).AllByClusterID(nil, currentCluster.ID)
		graphsChan <- graphsWithError
	}(currentCluster, id)

	go func(currentCluster *dal.ClusterRow) {
		metricsWithError := &dal.MetricRowsWithError{}
		metricsWithError.Metrics, metricsWithError.Error = dal.NewMetric(dbs.Core).AllByClusterID(nil, currentCluster.ID)
		metricsChan <- metricsWithError
	}(currentCluster)

	// -----------------------------------
	// Wait for channels to return results
	// -----------------------------------
	metricsWithError := <-metricsChan
	if metricsWithError.Error != nil {
		libhttp.HandleErrorJson(w, metricsWithError.Error)
		return
	}

	var currentGraph *dal.GraphRow
	graphsWithError := <-graphsChan
	if graphsWithError.Error == nil {
		for _, graph := range graphsWithError.Graphs {
			if graph.ID == id {
				currentGraph = graph
				break
			}
		}
	} else {
		libhttp.HandleErrorJson(w, graphsWithError.Error)
		return
	}

	accessToken, err := getAccessToken(w, r, "read")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	data := struct {
		CSRFToken      string
		Addr           string
		CurrentUser    *dal.UserRow
		AccessToken    *dal.AccessTokenRow
		Clusters       []*dal.ClusterRow
		CurrentCluster *dal.ClusterRow
		CurrentGraph   *dal.GraphRow
		Graphs         []*dal.GraphRow
		Metrics        []*dal.MetricRow
	}{
		csrf.Token(r),
		context.Get(r, "addr").(string),
		currentUser,
		accessToken,
		context.Get(r, "clusters").([]*dal.ClusterRow),
		currentCluster,
		currentGraph,
		graphsWithError.Graphs,
		metricsWithError.Metrics,
	}

	var tmpl *template.Template

	currentUserPermission := currentCluster.GetLevelByUserID(currentUser.ID)
	if currentUserPermission == "read" {
		tmpl, err = template.ParseFiles("templates/dashboard.html.tmpl", "templates/graphs/dashboard-readonly.html.tmpl")
	} else {
		tmpl, err = template.ParseFiles("templates/dashboard.html.tmpl", "templates/graphs/dashboard.html.tmpl")
	}
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	tmpl.Execute(w, data)
}

func PutGraphsID(w http.ResponseWriter, r *http.Request) {
	dbs := context.Get(r, "dbs").(*config.DBConfig)

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	err = r.ParseForm()
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	name := r.FormValue("Name")
	description := r.FormValue("Description")
	range_ := r.FormValue("Range")

	data := make(map[string]interface{})
	if name != "" {
		data["name"] = name
	}
	if description != "" {
		data["description"] = description
	}
	if range_ != "" {
		data["range"] = range_
	}

	metricsJSON := r.FormValue("MetricsWithOrder")
	if metricsJSON == "" {
		data["metrics"] = []byte("[]")
	} else {
		data["metrics"] = []byte(metricsJSON)
	}

	if len(data) > 0 {
		_, err = dal.NewGraph(dbs.Core).UpdateByID(nil, data, id)
		if err != nil {
			libhttp.HandleErrorJson(w, err)
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/graphs/%v", id), 301)
}

func DeleteGraphsID(w http.ResponseWriter, r *http.Request) {
	dbs := context.Get(r, "dbs").(*config.DBConfig)

	currentCluster := context.Get(r, "currentCluster").(*dal.ClusterRow)

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	_, err = dal.NewGraph(dbs.Core).DeleteByClusterIDAndID(nil, currentCluster.ID, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/graphs", 301)
}

func PutApiGraphsIDMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbs := context.Get(r, "dbs").(*config.DBConfig)

	accessTokenRow := context.Get(r, "accessToken").(*dal.AccessTokenRow)

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	dataJSON, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	row, err := dal.NewGraph(dbs.Core).UpdateMetricsByClusterIDAndID(nil, accessTokenRow.ClusterID, id, dataJSON)

	rowJSON, err := json.Marshal(row)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(rowJSON)
}
