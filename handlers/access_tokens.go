package handlers

import (
	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	resourcedmaster_dal "github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/libtemplate"
	"net/http"
)

func GetAccessTokens(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	db := context.Get(r, "db").(*sqlx.DB)

	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)

	session, _ := cookieStore.Get(r, "resourcedmaster-session")
	currentUser, ok := session.Values["user"].(*resourcedmaster_dal.UserRow)
	if !ok {
		http.Redirect(w, r, "/logout", 301)
		return
	}

	riceBoxes := context.Get(r, "riceBoxes").(map[string]*rice.Box)

	tmpl, err := libtemplate.GetFromRicebox(riceBoxes["templates"], false, "dashboard.html.tmpl", "access-tokens/list.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	accessTokens, err := resourcedmaster_dal.NewAccessToken(db).AllAccessTokens(nil)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	data := struct {
		CurrentUser  *resourcedmaster_dal.UserRow
		AccessTokens []*resourcedmaster_dal.AccessTokenRow
	}{
		currentUser,
		accessTokens,
	}

	tmpl.Execute(w, data)
}
