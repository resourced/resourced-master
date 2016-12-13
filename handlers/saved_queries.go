package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/models/cassandra"
)

func PostSavedQueries(w http.ResponseWriter, r *http.Request) {
	currentUser := r.Context().Value("currentUser").(*cassandra.UserRow)

	accessTokenRow, err := cassandra.NewAccessToken(r.Context()).GetByUserID(currentUser.ID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	savedQueryType := r.FormValue("Type")
	savedQuery := r.FormValue("SavedQuery")

	_, err = cassandra.NewSavedQuery(r.Context()).CreateOrUpdate(accessTokenRow, savedQueryType, savedQuery)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

func PostPutDeleteSavedQueriesID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	method := r.FormValue("_method")

	if strings.ToLower(method) == "delete" {
		DeleteSavedQueriesID(w, r)
	}
}

func DeleteSavedQueriesID(w http.ResponseWriter, r *http.Request) {
	savedQueryID, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	currentUser := r.Context().Value("currentUser").(*cassandra.UserRow)

	sq := cassandra.NewSavedQuery(r.Context())

	savedQueryRow, err := sq.GetByID(savedQueryID)

	if currentUser.ID != savedQueryRow.UserID {
		err := errors.New("Modifying other user's saved query is not allowed.")
		libhttp.HandleErrorJson(w, err)
		return
	}

	err = sq.DeleteByID(savedQueryID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}
