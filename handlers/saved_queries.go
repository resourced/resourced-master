package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/resourced/resourced-master/contexthelper"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/models/pg"
)

func PostSavedQueries(w http.ResponseWriter, r *http.Request) {
	pgdbs, err := contexthelper.GetPGDBConfig(r.Context())
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	currentUser := r.Context().Value("currentUser").(*pg.UserRow)

	accessTokenRow, err := pg.NewAccessToken(pgdbs.Core).GetByUserID(nil, currentUser.ID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	savedQueryType := r.FormValue("Type")
	savedQuery := r.FormValue("SavedQuery")

	_, err = pg.NewSavedQuery(pgdbs.Core).CreateOrUpdate(nil, accessTokenRow, savedQueryType, savedQuery)
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

	pgdbs, err := contexthelper.GetPGDBConfig(r.Context())
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	currentUser := r.Context().Value("currentUser").(*pg.UserRow)

	currentCluster := r.Context().Value("currentCluster").(*pg.ClusterRow)

	sq := pg.NewSavedQuery(pgdbs.Core)

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
