package handlers

import (
	"net/http"

	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/models/cassandra"
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

	_, err = cassandra.NewAccessToken(r.Context()).Create(currentUser.ID, clusterID, level)
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

	err = cassandra.NewAccessToken(r.Context()).UpdateLevelByID(tokenID, level)
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

	at := cassandra.NewAccessToken(r.Context())

	accessTokenRow, err := at.GetByID(tokenID)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	err = at.UpdateEnabledByID(tokenID, !accessTokenRow.Enabled)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
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

	err = cassandra.NewAccessToken(r.Context()).DeleteByID(tokenID)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	http.Redirect(w, r, "/clusters", 301)
}
