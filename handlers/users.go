package handlers

import (
	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/libtemplate"
	"net/http"
)

func GetSignup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	riceBoxes := context.Get(r, "riceBoxes").(map[string]*rice.Box)

	tmpl, err := libtemplate.GetFromRicebox(riceBoxes["templates"], false, "users/login-signup-parent.html.tmpl", "users/signup.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tmpl.Execute(w, nil)
}

func GetLoginWithoutSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	riceBoxes := context.Get(r, "riceBoxes").(map[string]*rice.Box)

	tmpl, err := libtemplate.GetFromRicebox(riceBoxes["templates"], false, "users/login-signup-parent.html.tmpl", "users/login.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tmpl.Execute(w, nil)
}

func GetLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)

	session, _ := cookieStore.Get(r, "resourcedmaster-session")

	currentUserInterface := session.Values["user"]
	if currentUserInterface != nil {
		http.Redirect(w, r, "/", 301)
		return
	}

	GetLoginWithoutSession(w, r)
}
