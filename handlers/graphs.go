package handlers

import (
	"html/template"
	"net/http"

	"github.com/gorilla/context"
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
