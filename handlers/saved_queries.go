package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/models/pg"
	"github.com/resourced/resourced-master/models/shared"
)

func PostSavedQueries(w http.ResponseWriter, r *http.Request) {
	currentUser := r.Context().Value("currentUser").(*shared.UserRow)

	accessTokenRow, err := pg.NewAccessToken(r.Context()).GetByUserID(nil, currentUser.ID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	savedQueryType := r.FormValue("Type")
	savedQuery := r.FormValue("SavedQuery")

	_, err = pg.NewSavedQuery(r.Context()).CreateOrUpdate(nil, accessTokenRow, savedQueryType, savedQuery)
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

	currentUser := r.Context().Value("currentUser").(*shared.UserRow)

	currentCluster := r.Context().Value("currentCluster").(*pg.ClusterRow)

	sq := pg.NewSavedQuery(r.Context())

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

	http.Redirect(w, r, r.Referer(), 301)
}
