package handlers

import (
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	resourcedmaster_dal "github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
	"net/http"
)

func PostSavedQueries(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "db").(*sqlx.DB)

	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)

	session, _ := cookieStore.Get(r, "resourcedmaster-session")

	currentUser := session.Values["user"].(*resourcedmaster_dal.UserRow)

	accessTokenRow, err := resourcedmaster_dal.NewAccessToken(db).GetByUserId(nil, currentUser.ID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	savedQuery := r.FormValue("SavedQuery")

	_, err = resourcedmaster_dal.NewSavedQuery(db).CreateOrUpdate(nil, accessTokenRow.ID, savedQuery)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/?q="+savedQuery, 301)
}
