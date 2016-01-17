package handlers

import (
	"html/template"
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
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

	// currentCluster := currentClusterInterface.(*dal.ClusterRow)

	data := struct {
		Addr               string
		CurrentUser        *dal.UserRow
		Clusters           []*dal.ClusterRow
		CurrentClusterJson string
	}{
		context.Get(r, "addr").(string),
		currentUserRow,
		context.Get(r, "clusters").([]*dal.ClusterRow),
		string(context.Get(r, "currentClusterJson").([]byte)),
	}

	tmpl, err := template.ParseFiles("templates/dashboard.html.tmpl", "templates/graphs/list.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tmpl.Execute(w, data)
}
