package handlers

import (
	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/libtemplate"
	"net/http"
)

func GetHosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)

	session, _ := cookieStore.Get(r, "resourcedmaster-session")
	_, ok := session.Values["user"]
	if !ok {
		http.Redirect(w, r, "/logout", 301)
		return
	}

	riceBoxes := context.Get(r, "riceBoxes").(map[string]*rice.Box)

	tmpl, err := libtemplate.GetFromRicebox(riceBoxes["templates"], false, "dashboard.html.tmpl", "hosts/list.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tmpl.Execute(w, nil)
}
