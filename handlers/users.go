package handlers

import (
	"github.com/GeertJohan/go.rice"
	// chillax_storage "github.com/chillaxio/chillax/storage"
	"github.com/gorilla/context"
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

	// storages := context.Get(r, "storages").(*chillax_storage.Storages)

	// session, _ := storages.Cookie.Get(r, "chillax-session")

	// currentUserInterface := session.Values["user"]
	// if currentUserInterface != nil {
	// 	http.Redirect(w, r, "/", 301)
	// 	return
	// }

	GetLoginWithoutSession(w, r)
}
