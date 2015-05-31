package handlers

import (
	"encoding/json"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	resourcedmaster_dal "github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
	"html/template"
	"io/ioutil"
	"net/http"
)

func GetHosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)

	session, _ := cookieStore.Get(r, "resourcedmaster-session")
	currentUser, ok := session.Values["user"].(*resourcedmaster_dal.UserRow)
	if !ok {
		http.Redirect(w, r, "/logout", 301)
		return
	}

	db := context.Get(r, "db").(*sqlx.DB)

	query := r.URL.Query().Get("q")

	var hosts []*resourcedmaster_dal.HostRow
	var savedQueries []*resourcedmaster_dal.SavedQueryRow

	accessTokenRow, _ := resourcedmaster_dal.NewAccessToken(db).GetByUserId(nil, currentUser.ID)

	if accessTokenRow != nil {
		hosts, _ = resourcedmaster_dal.NewHost(db).AllByAccessTokenIdAndQuery(nil, accessTokenRow.ID, query)
		savedQueries, _ = resourcedmaster_dal.NewSavedQuery(db).AllByAccessToken(nil, accessTokenRow)
	}

	data := struct {
		CurrentUser  *resourcedmaster_dal.UserRow
		Hosts        []*resourcedmaster_dal.HostRow
		SavedQueries []*resourcedmaster_dal.SavedQueryRow
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

	accessTokenRow := context.Get(r, "accessTokenRow").(*resourcedmaster_dal.AccessTokenRow)

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	hostRow, err := resourcedmaster_dal.NewHost(db).CreateOrUpdate(nil, accessTokenRow.ID, dataJson)
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

	accessTokenRow := context.Get(r, "accessTokenRow").(*resourcedmaster_dal.AccessTokenRow)

	query := r.URL.Query().Get("q")

	hosts, err := resourcedmaster_dal.NewHost(db).AllByAccessTokenIdAndQuery(nil, accessTokenRow.ID, query)
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
