package handlers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/context"
	"github.com/jmoiron/sqlx"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func GetGraphs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	currentUser := context.Get(r, "currentUser").(*dal.UserRow)

	currentCluster := context.Get(r, "currentCluster").(*dal.ClusterRow)

	db := context.Get(r, "db.Core").(*sqlx.DB)

	graphs, err := dal.NewGraph(db).AllByClusterID(nil, currentCluster.ID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	data := struct {
		Addr               string
		CurrentUser        *dal.UserRow
		Clusters           []*dal.ClusterRow
		CurrentClusterJson string
		Graphs             []*dal.GraphRow
	}{
		context.Get(r, "addr").(string),
		currentUser,
		context.Get(r, "clusters").([]*dal.ClusterRow),
		string(context.Get(r, "currentClusterJson").([]byte)),
		graphs,
	}

	tmpl, err := template.ParseFiles("templates/dashboard.html.tmpl", "templates/graphs/list.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tmpl.Execute(w, data)
}

func PostGraphs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	currentCluster := context.Get(r, "currentCluster").(*dal.ClusterRow)

	name := r.FormValue("Name")
	description := r.FormValue("Description")

	db := context.Get(r, "db.Core").(*sqlx.DB)

	_, err := dal.NewGraph(db).Create(nil, currentCluster.ID, name, description)
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

	db := context.Get(r, "db.Core").(*sqlx.DB)

	currentUser := context.Get(r, "currentUser").(*dal.UserRow)

	currentCluster := context.Get(r, "currentCluster").(*dal.ClusterRow)

	id, err := getIdFromPath(w, r)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	accessTokenChan := make(chan *dal.AccessTokenRow)
	graphsChan := make(chan []*dal.GraphRow)
	currentGraphChan := make(chan *dal.GraphRow)
	metricsChan := make(chan []*dal.MetricRow)
	errChan := make(chan error)

	go func(currentUser *dal.UserRow) {
		accessToken, err := dal.NewAccessToken(db).GetByUserID(nil, currentUser.ID)
		accessTokenChan <- accessToken
		errChan <- err
	}(currentUser)

	go func(currentCluster *dal.ClusterRow, id int64) {
		graphs, err := dal.NewGraph(db).AllByClusterID(nil, currentCluster.ID)

		if err == nil {
			for _, graph := range graphs {
				if graph.ID == id {
					currentGraphChan <- graph
					break
				}
			}
		}

		graphsChan <- graphs
		errChan <- err
	}(currentCluster, id)

	go func(currentCluster *dal.ClusterRow) {
		metrics, err := dal.NewMetric(db).AllByClusterID(nil, currentCluster.ID)
		metricsChan <- metrics
		errChan <- err
	}(currentCluster)

	defer close(accessTokenChan)
	defer close(graphsChan)
	defer close(currentGraphChan)
	defer close(metricsChan)
	// defer close(errChan)

	// err = <-errChan
	// if err != nil {
	// 	libhttp.HandleErrorJson(w, err)
	// 	return
	// }

	data := struct {
		Addr               string
		CurrentUser        *dal.UserRow
		AccessToken        *dal.AccessTokenRow
		Clusters           []*dal.ClusterRow
		CurrentClusterJson string
		CurrentGraph       *dal.GraphRow
		Graphs             []*dal.GraphRow
		Metrics            []*dal.MetricRow
	}{
		context.Get(r, "addr").(string),
		currentUser,
		<-accessTokenChan,
		context.Get(r, "clusters").([]*dal.ClusterRow),
		string(context.Get(r, "currentClusterJson").([]byte)),
		<-currentGraphChan,
		<-graphsChan,
		<-metricsChan,
	}

	tmpl, err := template.ParseFiles("templates/dashboard.html.tmpl", "templates/graphs/dashboard.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tmpl.Execute(w, data)
}

func PutGraphsID(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "db.Core").(*sqlx.DB)

	currentCluster := context.Get(r, "currentCluster").(*dal.ClusterRow)

	id, err := getIdFromPath(w, r)
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

	data := make(map[string]interface{})
	if name != "" {
		data["name"] = name
	}
	if description != "" {
		data["description"] = description
	}

	metrics := r.Form["MetricsWithOrder"]

	if len(metrics) > 0 {
		metricsJSONBytes, err := dal.NewGraph(db).BuildMetricsJSONForSave(nil, currentCluster.ID, metrics)
		if err != nil {
			libhttp.HandleErrorJson(w, err)
			return
		}
		data["metrics"] = metricsJSONBytes
	}

	_, err = dal.NewGraph(db).UpdateByID(nil, data, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/graphs/%v", id), 301)
}

func DeleteGraphsID(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "db.Core").(*sqlx.DB)

	accessTokenRow := context.Get(r, "accessTokenRow").(*dal.AccessTokenRow)

	id, err := getIdFromPath(w, r)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	_, err = dal.NewGraph(db).DeleteByClusterIDAndID(nil, accessTokenRow.ClusterID, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/graphs", 301)
}
