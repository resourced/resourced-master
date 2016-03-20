package handlers

import (
	"net/http"

	"github.com/gorilla/context"
	"github.com/jmoiron/sqlx"

	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func PostAccessTokens(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "db.Core").(*sqlx.DB)

	currentUser := context.Get(r, "currentUser").(*dal.UserRow)

	clusterID, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	level := r.FormValue("Level")

	_, err = dal.NewAccessToken(db).Create(nil, currentUser.ID, clusterID, level)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/clusters", 301)
}

func PostAccessTokensLevel(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "db.Core").(*sqlx.DB)

	tokenID, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	level := r.FormValue("Level")

	data := make(map[string]interface{})
	data["level"] = level

	_, err = dal.NewAccessToken(db).UpdateByID(nil, data, tokenID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/clusters", 301)
}

func PostAccessTokensEnabled(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "db.Core").(*sqlx.DB)

	tokenID, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	at := dal.NewAccessToken(db)

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
