package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/csrf"

	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/models/cassandra"
)

func GetGraphs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	currentUser := r.Context().Value("currentUser").(*cassandra.UserRow)

	currentCluster := r.Context().Value("currentCluster").(*cassandra.ClusterRow)

	graphs, err := cassandra.NewGraph(r.Context()).AllByClusterID(currentCluster.ID)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	data := struct {
		CSRFToken      string
		Addr           string
		CurrentUser    *cassandra.UserRow
		Clusters       []*cassandra.ClusterRow
		CurrentCluster *cassandra.ClusterRow
		Graphs         []*cassandra.GraphRow
	}{
		csrf.Token(r),
		r.Context().Value("Addr").(string),
		currentUser,
		r.Context().Value("clusters").([]*cassandra.ClusterRow),
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

	currentCluster := r.Context().Value("currentCluster").(*cassandra.ClusterRow)

	name := r.FormValue("Name")
	description := r.FormValue("Description")
	range_ := r.FormValue("Range")

	data := make(map[string]interface{})
	data["name"] = name
	data["description"] = description
	data["range"] = range_

	_, err := cassandra.NewGraph(r.Context()).Create(currentCluster.ID, data)
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

	currentUser := r.Context().Value("currentUser").(*cassandra.UserRow)

	currentCluster := r.Context().Value("currentCluster").(*cassandra.ClusterRow)

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

	// -----------------------------------
	// Create channels to receive SQL rows
	// -----------------------------------
	metricsChan := make(chan *cassandra.MetricRowsWithError)
	defer close(metricsChan)

	graphsChan := make(chan *cassandra.GraphRowsWithError)
	defer close(graphsChan)

	// --------------------------
	// Fetch SQL rows in parallel
	// --------------------------
	go func(currentCluster *cassandra.ClusterRow, id int64) {
		graphsWithError := &cassandra.GraphRowsWithError{}
		graphsWithError.Graphs, graphsWithError.Error = cassandra.NewGraph(r.Context()).AllByClusterID(currentCluster.ID)
		graphsChan <- graphsWithError
	}(currentCluster, id)

	go func(currentCluster *cassandra.ClusterRow) {
		metricsWithError := &cassandra.MetricRowsWithError{}
		metricsWithError.Metrics, metricsWithError.Error = cassandra.NewMetric(r.Context()).AllByClusterID(currentCluster.ID)
		metricsChan <- metricsWithError
	}(currentCluster)

	// -----------------------------------
	// Wait for channels to return results
	// -----------------------------------
	hasError := false

	metricsWithError := <-metricsChan
	if metricsWithError.Error != nil {
		libhttp.HandleErrorJson(w, metricsWithError.Error)
		hasError = true
	}

	var currentGraph *cassandra.GraphRow
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
		CurrentGraph   *cassandra.GraphRow
		Graphs         []*cassandra.GraphRow
		Metrics        []*cassandra.MetricRow
	}{
		csrf.Token(r),
		r.Context().Value("Addr").(string),
		currentUser,
		accessToken,
		r.Context().Value("clusters").([]*cassandra.ClusterRow),
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
	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	err = r.ParseForm()
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	name := r.FormValue("Name")
	description := r.FormValue("Description")
	range_ := r.FormValue("Range")

	data := make(map[string]string)
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
		data["metrics"] = "[]"
	} else {
		data["metrics"] = metricsJSON
	}

	if len(data) > 0 {
		_, err = cassandra.NewGraph(r.Context()).UpdateByID(id, data)
		if err != nil {
			libhttp.HandleErrorHTML(w, err, 500)
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/graphs/%v", id), 301)
}

func DeleteGraphsID(w http.ResponseWriter, r *http.Request) {
	currentCluster := r.Context().Value("currentCluster").(*cassandra.ClusterRow)

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	err = cassandra.NewGraph(r.Context()).DeleteByClusterIDAndID(currentCluster.ID, id)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	http.Redirect(w, r, "/graphs", 301)
}

func PutApiGraphsIDMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accessTokenRow := r.Context().Value("accessToken").(*cassandra.AccessTokenRow)

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	dataJSON, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	row, err := cassandra.NewGraph(r.Context()).UpdateMetricsByClusterIDAndID(accessTokenRow.ClusterID, id, dataJSON)

	rowJSON, err := json.Marshal(row)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	w.Write(rowJSON)
}
