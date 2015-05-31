package handlers

import (
	"errors"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	rm_dal "github.com/resourced/resourced-master/dal"
	"net/http"
	"strconv"
)

func getCurrentUser(w http.ResponseWriter, r *http.Request) *rm_dal.UserRow {
	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)
	session, _ := cookieStore.Get(r, "resourcedmaster-session")
	return session.Values["user"].(*rm_dal.UserRow)
}

func getIdFromPath(w http.ResponseWriter, r *http.Request) (int64, error) {
	userIdString := mux.Vars(r)["id"]
	if userIdString == "" {
		return -1, errors.New("user id cannot be empty.")
	}

	userId, err := strconv.ParseInt(userIdString, 10, 64)
	if err != nil {
		return -1, err
	}

	return userId, nil
}
