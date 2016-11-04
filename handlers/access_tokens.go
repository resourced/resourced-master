package handlers

import (
	"net/http"

	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/models/pg"
	"github.com/resourced/resourced-master/models/shared"
)

func PostAccessTokens(w http.ResponseWriter, r *http.Request) {
	currentUser := r.Context().Value("currentUser").(*shared.UserRow)

	clusterID, err := getInt64SlugFromPath(w, r, "clusterID")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	level := r.FormValue("Level")

	_, err = pg.NewAccessToken(r.Context()).Create(nil, currentUser.ID, clusterID, level)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	http.Redirect(w, r, "/clusters", 301)
}

func PostAccessTokensLevel(w http.ResponseWriter, r *http.Request) {
	tokenID, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	level := r.FormValue("Level")

	data := make(map[string]interface{})
	data["level"] = level

	_, err = pg.NewAccessToken(r.Context()).UpdateByID(nil, data, tokenID)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	http.Redirect(w, r, "/clusters", 301)
}

func PostAccessTokensEnabled(w http.ResponseWriter, r *http.Request) {
	tokenID, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	at := pg.NewAccessToken(r.Context())

	accessTokenRow, err := at.GetByID(nil, tokenID)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	data := make(map[string]interface{})
	data["enabled"] = !accessTokenRow.Enabled

	_, err = at.UpdateByID(nil, data, tokenID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/clusters", 301)
}

func PostAccessTokensDelete(w http.ResponseWriter, r *http.Request) {
	tokenID, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	_, err = pg.NewAccessToken(r.Context()).DeleteByID(nil, tokenID)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	http.Redirect(w, r, "/clusters", 301)
}
