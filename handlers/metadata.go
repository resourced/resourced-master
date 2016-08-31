package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pressly/chi"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func GetApiMetadata(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbs := r.Context().Value("dbs").(*config.DBConfig)

	accessTokenRow := r.Context().Value("accessToken").(*dal.AccessTokenRow)

	metadataRows, err := dal.NewMetadata(dbs.Core).AllByClusterID(nil, accessTokenRow.ClusterID)
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

	dbs := r.Context().Value("dbs").(*config.DBConfig)

	accessTokenRow := r.Context().Value("accessToken").(*dal.AccessTokenRow)

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	key := chi.URLParam(r, "key")

	metadataRow, err := dal.NewMetadata(dbs.Core).CreateOrUpdate(nil, accessTokenRow.ClusterID, key, dataJson)
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

	dbs := r.Context().Value("dbs").(*config.DBConfig)

	accessTokenRow := r.Context().Value("accessToken").(*dal.AccessTokenRow)

	key := chi.URLParam(r, "key")

	metadataRow, err := dal.NewMetadata(dbs.Core).DeleteByClusterIDAndKey(nil, accessTokenRow.ClusterID, key)
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

	dbs := r.Context().Value("dbs").(*config.DBConfig)

	accessTokenRow := r.Context().Value("accessToken").(*dal.AccessTokenRow)

	key := chi.URLParam(r, "key")

	metadataRow, err := dal.NewMetadata(dbs.Core).GetByClusterIDAndKey(nil, accessTokenRow.ClusterID, key)
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
