package handlers

import (
	"net/http"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/models/pg"
)

func PostAccessTokens(w http.ResponseWriter, r *http.Request) {
	dbs := r.Context().Value("dbs").(*config.PGDBConfig)

	currentUser := r.Context().Value("currentUser").(*pg.UserRow)

	clusterID, err := getInt64SlugFromPath(w, r, "clusterID")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	level := r.FormValue("Level")

	_, err = pg.NewAccessToken(dbs.Core).Create(nil, currentUser.ID, clusterID, level)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/clusters", 301)
}

func PostAccessTokensLevel(w http.ResponseWriter, r *http.Request) {
	dbs := r.Context().Value("dbs").(*config.PGDBConfig)

	tokenID, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	level := r.FormValue("Level")

	data := make(map[string]interface{})
	data["level"] = level

	_, err = pg.NewAccessToken(dbs.Core).UpdateByID(nil, data, tokenID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/clusters", 301)
}

func PostAccessTokensEnabled(w http.ResponseWriter, r *http.Request) {
	dbs := r.Context().Value("dbs").(*config.PGDBConfig)

	tokenID, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	at := pg.NewAccessToken(dbs.Core)

	accessTokenRow, err := at.GetByID(nil, tokenID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
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
	dbs := r.Context().Value("dbs").(*config.PGDBConfig)

	tokenID, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	_, err = pg.NewAccessToken(dbs.Core).DeleteByID(nil, tokenID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/clusters", 301)
}
