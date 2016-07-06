package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/context"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func PostSavedQueries(w http.ResponseWriter, r *http.Request) {
	dbs := context.Get(r, "dbs").(*config.DBConfig)

	currentUser := context.Get(r, "currentUser").(*dal.UserRow)

	accessTokenRow, err := dal.NewAccessToken(dbs.Core).GetByUserID(nil, currentUser.ID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	savedQueryType := r.FormValue("Type")
	savedQuery := r.FormValue("SavedQuery")

	_, err = dal.NewSavedQuery(dbs.Core).CreateOrUpdate(nil, accessTokenRow, savedQueryType, savedQuery)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

func PostPutDeleteSavedQueriesID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	method := r.FormValue("_method")
	if method == "" || strings.ToLower(method) == "post" || strings.ToLower(method) == "put" {
		PutSavedQueriesID(w, r)
	} else if strings.ToLower(method) == "delete" {
		DeleteSavedQueriesID(w, r)
	}
}

func PutSavedQueriesID(w http.ResponseWriter, r *http.Request) {
	err := errors.New("PUT method is not implemented.")
	libhttp.HandleErrorJson(w, err)
	return
}

func DeleteSavedQueriesID(w http.ResponseWriter, r *http.Request) {
	savedQueryID, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	dbs := context.Get(r, "dbs").(*config.DBConfig)

	currentUser := context.Get(r, "currentUser").(*dal.UserRow)

	currentCluster := context.Get(r, "currentCluster").(*dal.ClusterRow)

	sq := dal.NewSavedQuery(dbs.Core)

	savedQueryRow, err := sq.GetByID(nil, savedQueryID)

	if currentUser.ID != savedQueryRow.UserID {
		err := errors.New("Modifying other user's saved query is not allowed.")
		libhttp.HandleErrorJson(w, err)
		return
	}

	_, err = sq.DeleteByClusterIDAndID(nil, currentCluster.ID, savedQueryID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/", 301)
}
