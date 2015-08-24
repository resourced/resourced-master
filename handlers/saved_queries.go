package handlers

import (
	"errors"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	rm_dal "github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
	"net/http"
	"strings"
)

func PostSavedQueries(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "db").(*sqlx.DB)

	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)

	session, _ := cookieStore.Get(r, "resourcedmaster-session")

	currentUser := session.Values["user"].(*rm_dal.UserRow)

	accessTokenRow, err := rm_dal.NewAccessToken(db).GetByUserId(nil, currentUser.ID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	savedQuery := r.FormValue("SavedQuery")

	_, err = rm_dal.NewSavedQuery(db).CreateOrUpdate(nil, accessTokenRow, savedQuery)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/?q="+savedQuery, 301)
}

func PostPutDeleteSavedQueriesID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	method := r.FormValue("_method")
	if method == "" || strings.ToLower(method) == "post" || strings.ToLower(method) == "put" {
		PutSavedQueriesID(w, r)
	} else if strings.ToLower(method) == "delete" {
		DeleteSavedQueriesID(w, r)
	}
}

func PutSavedQueriesID(w http.ResponseWriter, r *http.Request) {
	err := errors.New("PUT method is not implemented.")
	libhttp.HandleErrorJson(w, err)
	return
}

func DeleteSavedQueriesID(w http.ResponseWriter, r *http.Request) {
	savedQueryID, err := getIdFromPath(w, r)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	db := context.Get(r, "db").(*sqlx.DB)

	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)

	session, _ := cookieStore.Get(r, "resourcedmaster-session")

	currentUser := session.Values["user"].(*rm_dal.UserRow)

	sq := rm_dal.NewSavedQuery(db)

	savedQueryRow, err := sq.GetByID(nil, savedQueryID)

	if currentUser.ID != savedQueryRow.UserID {
		err := errors.New("Modifying other user's saved query is not allowed.")
		libhttp.HandleErrorJson(w, err)
		return
	}

	err = sq.DeleteByID(nil, savedQueryID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/", 301)
}
