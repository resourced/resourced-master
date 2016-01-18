package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func GetGraphs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)
	session, _ := cookieStore.Get(r, "resourcedmaster-session")

	currentUserRow, ok := session.Values["user"].(*dal.UserRow)
	if !ok {
		http.Redirect(w, r, "/logout", 301)
		return
	}

	currentClusterInterface := session.Values["currentCluster"]
	if currentClusterInterface == nil {
		http.Redirect(w, r, "/", 301)
		return
	}

	currentCluster := currentClusterInterface.(*dal.ClusterRow)

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
		currentUserRow,
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

	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)
	session, _ := cookieStore.Get(r, "resourcedmaster-session")

	currentClusterInterface := session.Values["currentCluster"]
	if currentClusterInterface == nil {
		http.Redirect(w, r, "/", 301)
		return
	}

	currentCluster := currentClusterInterface.(*dal.ClusterRow)

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

	idString := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)
	session, _ := cookieStore.Get(r, "resourcedmaster-session")

	currentUserRow, ok := session.Values["user"].(*dal.UserRow)
	if !ok {
		http.Redirect(w, r, "/logout", 301)
		return
	}

	accessTokenRow, err := dal.NewAccessToken(db).GetByUserID(nil, currentUserRow.ID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	currentClusterInterface := session.Values["currentCluster"]
	if currentClusterInterface == nil {
		http.Redirect(w, r, "/", 301)
		return
	}

	currentCluster := currentClusterInterface.(*dal.ClusterRow)

	graphs, err := dal.NewGraph(db).AllByClusterID(nil, currentCluster.ID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	currentGraph, err := dal.NewGraph(db).GetById(nil, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	metrics, err := dal.NewMetric(db).AllByClusterID(nil, currentCluster.ID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

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
		currentUserRow,
		accessTokenRow,
		context.Get(r, "clusters").([]*dal.ClusterRow),
		string(context.Get(r, "currentClusterJson").([]byte)),
		currentGraph,
		graphs,
		metrics,
	}

	tmpl, err := template.ParseFiles("templates/dashboard.html.tmpl", "templates/graphs/dashboard.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tmpl.Execute(w, data)
}

func PutGraphsID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	idString := vars["id"]
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	data := make(map[string]interface{})
	data["name"] = r.FormValue("Name")
	data["description"] = r.FormValue("Description")

	db := context.Get(r, "db.Core").(*sqlx.DB)

	_, err = dal.NewGraph(db).UpdateByID(nil, data, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/graphs", 301)
}

func DeleteGraphsID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	idString := vars["id"]
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	db := context.Get(r, "db.Core").(*sqlx.DB)

	_, err = dal.NewGraph(db).DeleteByID(nil, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/graphs", 301)
}
