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

func GetTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)

	session, _ := cookieStore.Get(r, "resourcedmaster-session")
	currentUser, ok := session.Values["user"].(*dal.UserRow)
	if !ok {
		http.Redirect(w, r, "/logout", 301)
		return
	}

	db := context.Get(r, "db").(*sqlx.DB)

	query := r.URL.Query().Get("q")

	var hosts []*dal.HostRow
	var savedQueries []*dal.SavedQueryRow

	accessTokenRow, _ := dal.NewAccessToken(db).GetByUserId(nil, currentUser.ID)

	if accessTokenRow != nil {
		hosts, _ = dal.NewHost(db).AllByAccessTokenIdAndQuery(nil, accessTokenRow.ID, query)
		savedQueries, _ = dal.NewSavedQuery(db).AllByAccessToken(nil, accessTokenRow)
	}

	data := struct {
		Addr         string
		CurrentUser  *dal.UserRow
		AccessToken  *dal.AccessTokenRow
		Hosts        []*dal.HostRow
		SavedQueries []*dal.SavedQueryRow
	}{
		context.Get(r, "addr").(string),
		currentUser,
		accessTokenRow,
		hosts,
		savedQueries,
	}

	tmpl, err := template.ParseFiles("templates/dashboard.html.tmpl", "templates/tasks/list.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tmpl.Execute(w, data)
}
