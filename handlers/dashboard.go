package handlers

import (
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	resourcedmaster_dal "github.com/resourced/resourced-master/dal"
	"net/http"
)

func GetDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)

	session, _ := cookieStore.Get(r, "resourcedmaster-session")
	user, ok := session.Values["user"]
	if !ok {
		http.Redirect(w, r, "/logout", 301)
		return
	}

	if user.(*resourcedmaster_dal.UserRow).ApplicationID.Int64 <= 0 {
		http.Redirect(w, r, "/applications/create", 301)
		return
	}
}
