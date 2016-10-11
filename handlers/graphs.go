package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/csrf"

	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/models/pg"
)

func GetGraphs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	currentUser := r.Context().Value("currentUser").(*pg.UserRow)

	currentCluster := r.Context().Value("currentCluster").(*pg.ClusterRow)

	graphs, err := pg.NewGraph(r.Context()).AllByClusterID(nil, currentCluster.ID)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	data := struct {
		CSRFToken      string
		Addr           string
		CurrentUser    *pg.UserRow
		Clusters       []*pg.ClusterRow
		CurrentCluster *pg.ClusterRow
		Graphs         []*pg.GraphRow
	}{
		csrf.Token(r),
		r.Context().Value("Addr").(string),
		currentUser,
		r.Context().Value("clusters").([]*pg.ClusterRow),
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

	currentCluster := r.Context().Value("currentCluster").(*pg.ClusterRow)

	name := r.FormValue("Name")
	description := r.FormValue("Description")
	range_ := r.FormValue("Range")

	data := make(map[string]interface{})
	data["name"] = name
	data["description"] = description
	data["range"] = range_

	_, err := pg.NewGraph(r.Context()).Create(nil, currentCluster.ID, data)
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

	// -----------------------------------
	// Create channels to receive SQL rows
	// -----------------------------------
	metricsChan := make(chan *pg.MetricRowsWithError)
	defer close(metricsChan)

	graphsChan := make(chan *pg.GraphRowsWithError)
	defer close(graphsChan)

	// --------------------------
	// Fetch SQL rows in parallel
	// --------------------------
	go func(currentCluster *pg.ClusterRow, id int64) {
		graphsWithError := &pg.GraphRowsWithError{}
		graphsWithError.Graphs, graphsWithError.Error = pg.NewGraph(r.Context()).AllByClusterID(nil, currentCluster.ID)
		graphsChan <- graphsWithError
	}(currentCluster, id)

	go func(currentCluster *pg.ClusterRow) {
		metricsWithError := &pg.MetricRowsWithError{}
		metricsWithError.Metrics, metricsWithError.Error = pg.NewMetric(r.Context()).AllByClusterID(nil, currentCluster.ID)
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

	var currentGraph *pg.GraphRow
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
		CurrentUser    *pg.UserRow
		AccessToken    *pg.AccessTokenRow
		Clusters       []*pg.ClusterRow
		CurrentCluster *pg.ClusterRow
		CurrentGraph   *pg.GraphRow
		Graphs         []*pg.GraphRow
		Metrics        []*pg.MetricRow
	}{
		csrf.Token(r),
		r.Context().Value("Addr").(string),
		currentUser,
		accessToken,
		r.Context().Value("clusters").([]*pg.ClusterRow),
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
		_, err = pg.NewGraph(r.Context()).UpdateByID(nil, data, id)
		if err != nil {
			libhttp.HandleErrorHTML(w, err, 500)
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/graphs/%v", id), 301)
}

func DeleteGraphsID(w http.ResponseWriter, r *http.Request) {
	currentCluster := r.Context().Value("currentCluster").(*pg.ClusterRow)

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	_, err = pg.NewGraph(r.Context()).DeleteByClusterIDAndID(nil, currentCluster.ID, id)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	http.Redirect(w, r, "/graphs", 301)
}

func PutApiGraphsIDMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accessTokenRow := r.Context().Value("accessToken").(*pg.AccessTokenRow)

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

	row, err := pg.NewGraph(r.Context()).UpdateMetricsByClusterIDAndID(nil, accessTokenRow.ClusterID, id, dataJSON)

	rowJSON, err := json.Marshal(row)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	w.Write(rowJSON)
}
