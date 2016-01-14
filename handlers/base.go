package handlers

import (
	"errors"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/resourced/resourced-master/dal"
	"net/http"
	"strconv"
)

func getCurrentUser(w http.ResponseWriter, r *http.Request) *dal.UserRow {
	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)
	session, _ := cookieStore.Get(r, "resourcedmaster-session")
	return session.Values["user"].(*dal.UserRow)
}

func getIdFromPath(w http.ResponseWriter, r *http.Request) (int64, error) {
	idString := mux.Vars(r)["id"]
	if idString == "" {
		return -1, errors.New("user id cannot be empty.")
	}

	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		return -1, err
	}

	return id, nil
}
