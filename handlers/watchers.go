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

func GetWatchers(w http.ResponseWriter, r *http.Request) {
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

	db := context.Get(r, "db").(*sqlx.DB)

	watchers, err := dal.NewWatcher(db).AllByClusterID(nil, currentCluster.ID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	savedQueries, err := dal.NewSavedQuery(db).AllByClusterID(nil, currentCluster.ID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	data := struct {
		Addr               string
		CurrentUser        *dal.UserRow
		Clusters           []*dal.ClusterRow
		CurrentClusterJson string
		Watchers           []*dal.WatcherRow
		SavedQueries       []*dal.SavedQueryRow
	}{
		context.Get(r, "addr").(string),
		currentUserRow,
		context.Get(r, "clusters").([]*dal.ClusterRow),
		string(context.Get(r, "currentClusterJson").([]byte)),
		watchers,
		savedQueries,
	}

	tmpl, err := template.ParseFiles("templates/dashboard.html.tmpl", "templates/watchers/list.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tmpl.Execute(w, data)
}
