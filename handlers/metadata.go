package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pressly/chi"

	"github.com/resourced/resourced-master/contexthelper"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/models/pg"
)

func GetApiMetadata(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pgdbs, err := contexthelper.GetPGDBConfig(r.Context())
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	accessTokenRow := r.Context().Value("accessToken").(*pg.AccessTokenRow)

	metadataRows, err := pg.NewMetadata(pgdbs.Core).AllByClusterID(nil, accessTokenRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	metadataRowsJson, err := json.Marshal(metadataRows)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(metadataRowsJson)
}

func PostApiMetadataKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pgdbs, err := contexthelper.GetPGDBConfig(r.Context())
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	accessTokenRow := r.Context().Value("accessToken").(*pg.AccessTokenRow)

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	key := chi.URLParam(r, "key")

	metadataRow, err := pg.NewMetadata(pgdbs.Core).CreateOrUpdate(nil, accessTokenRow.ClusterID, key, dataJson)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	metadataRowJson, err := json.Marshal(metadataRow)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(metadataRowJson)
}

func DeleteApiMetadataKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pgdbs, err := contexthelper.GetPGDBConfig(r.Context())
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	accessTokenRow := r.Context().Value("accessToken").(*pg.AccessTokenRow)

	key := chi.URLParam(r, "key")

	metadataRow, err := pg.NewMetadata(pgdbs.Core).DeleteByClusterIDAndKey(nil, accessTokenRow.ClusterID, key)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	metadataRowJson, err := json.Marshal(metadataRow)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(metadataRowJson)
}

func GetApiMetadataKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pgdbs, err := contexthelper.GetPGDBConfig(r.Context())
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	accessTokenRow := r.Context().Value("accessToken").(*pg.AccessTokenRow)

	key := chi.URLParam(r, "key")

	metadataRow, err := pg.NewMetadata(pgdbs.Core).GetByClusterIDAndKey(nil, accessTokenRow.ClusterID, key)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	metadataRowJson, err := json.Marshal(metadataRow)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(metadataRowJson)
}
