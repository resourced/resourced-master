package handlers

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func GetHosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)

	session, _ := cookieStore.Get(r, "resourcedmaster-session")
	currentUserRow, ok := session.Values["user"].(*dal.UserRow)
	if !ok {
		http.Redirect(w, r, "/logout", 301)
		return
	}

	db := context.Get(r, "db").(*sqlx.DB)

	query := r.URL.Query().Get("q")

	var hosts []*dal.HostRow
	var savedQueries []*dal.SavedQueryRow

	accessTokenRow, _ := dal.NewAccessToken(db).GetByUserId(nil, currentUserRow.ID)

	if accessTokenRow != nil {
		hosts, _ = dal.NewHost(db).AllByAccessTokenIdAndQuery(nil, accessTokenRow.ID, query)
		savedQueries, _ = dal.NewSavedQuery(db).AllByAccessToken(nil, accessTokenRow)
	}

	data := struct {
		Addr               string
		CurrentUser        *dal.UserRow
		AccessToken        *dal.AccessTokenRow
		Clusters           []*dal.ClusterRow
		CurrentClusterJson string
		Hosts              []*dal.HostRow
		SavedQueries       []*dal.SavedQueryRow
	}{
		context.Get(r, "addr").(string),
		currentUserRow,
		accessTokenRow,
		context.Get(r, "clusters").([]*dal.ClusterRow),
		string(context.Get(r, "currentClusterJson").([]byte)),
		hosts,
		savedQueries,
	}

	tmpl, err := template.ParseFiles("templates/dashboard.html.tmpl", "templates/hosts/list.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tmpl.Execute(w, data)
}

func PostApiHosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := context.Get(r, "db").(*sqlx.DB)

	accessTokenRow := context.Get(r, "accessTokenRow").(*dal.AccessTokenRow)

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	hostRow, err := dal.NewHost(db).CreateOrUpdate(nil, accessTokenRow.ID, dataJson)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	hostRowJson, err := json.Marshal(hostRow)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(hostRowJson)
}

func GetApiHosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := context.Get(r, "db").(*sqlx.DB)

	accessTokenRow := context.Get(r, "accessTokenRow").(*dal.AccessTokenRow)

	query := r.URL.Query().Get("q")

	hosts, err := dal.NewHost(db).AllByAccessTokenIdAndQuery(nil, accessTokenRow.ID, query)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	hostRowsJson, err := json.Marshal(hosts)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(hostRowsJson)
}
