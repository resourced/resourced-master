package handlers

import (
	"encoding/json"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	rm_dal "github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
	"html/template"
	"io/ioutil"
	"net/http"
)

func GetHosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)

	session, _ := cookieStore.Get(r, "resourcedmaster-session")
	currentUser, ok := session.Values["user"].(*rm_dal.UserRow)
	if !ok {
		http.Redirect(w, r, "/logout", 301)
		return
	}

	db := context.Get(r, "db").(*sqlx.DB)

	query := r.URL.Query().Get("q")

	var hosts []*rm_dal.HostRow
	var savedQueries []*rm_dal.SavedQueryRow

	accessTokenRow, _ := rm_dal.NewAccessToken(db).GetByUserId(nil, currentUser.ID)

	if accessTokenRow != nil {
		hosts, _ = rm_dal.NewHost(db).AllByAccessTokenIdAndQuery(nil, accessTokenRow.ID, query)
		savedQueries, _ = rm_dal.NewSavedQuery(db).AllByAccessToken(nil, accessTokenRow)
	}

	data := struct {
		CurrentUser  *rm_dal.UserRow
		Hosts        []*rm_dal.HostRow
		SavedQueries []*rm_dal.SavedQueryRow
	}{
		currentUser,
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

	accessTokenRow := context.Get(r, "accessTokenRow").(*rm_dal.AccessTokenRow)

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	hostRow, err := rm_dal.NewHost(db).CreateOrUpdate(nil, accessTokenRow.ID, dataJson)
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

	accessTokenRow := context.Get(r, "accessTokenRow").(*rm_dal.AccessTokenRow)

	query := r.URL.Query().Get("q")

	hosts, err := rm_dal.NewHost(db).AllByAccessTokenIdAndQuery(nil, accessTokenRow.ID, query)
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
