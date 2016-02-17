package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/context"
	"github.com/jmoiron/sqlx"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func PostSavedQueries(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "db.Core").(*sqlx.DB)

	currentUser := context.Get(r, "currentUser").(*dal.UserRow)

	accessTokenRow, err := dal.NewAccessToken(db).GetByUserID(nil, currentUser.ID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	savedQuery := r.FormValue("SavedQuery")

	_, err = dal.NewSavedQuery(db).CreateOrUpdate(nil, accessTokenRow, savedQuery)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/?q="+savedQuery, 301)
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
	savedQueryID, err := getIdFromPath(w, r)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	db := context.Get(r, "db.Core").(*sqlx.DB)

	currentUser := context.Get(r, "currentUser").(*dal.UserRow)

	currentCluster := context.Get(r, "currentCluster").(*dal.ClusterRow)

	sq := dal.NewSavedQuery(db)

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
