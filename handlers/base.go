package handlers

import (
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	resourcedmaster_dal "github.com/resourced/resourced-master/dal"
	"net/http"
)

func getCurrentUser(w http.ResponseWriter, r *http.Request) *resourcedmaster_dal.UserRow {
	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)
	session, _ := cookieStore.Get(r, "resourcedmaster-session")
	return session.Values["user"].(*resourcedmaster_dal.UserRow)
}
