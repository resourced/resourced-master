package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"

	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func GetApiMetadata(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := context.Get(r, "db.Core").(*sqlx.DB)

	accessTokenRow := context.Get(r, "accessToken").(*dal.AccessTokenRow)

	metadataRows, err := dal.NewMetadata(db).AllByClusterID(nil, accessTokenRow.ClusterID)
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

	db := context.Get(r, "db.Core").(*sqlx.DB)

	accessTokenRow := context.Get(r, "accessToken").(*dal.AccessTokenRow)

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	vars := mux.Vars(r)
	key := vars["key"]

	metadataRow, err := dal.NewMetadata(db).CreateOrUpdate(nil, accessTokenRow.ClusterID, key, dataJson)
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

	db := context.Get(r, "db.Core").(*sqlx.DB)

	accessTokenRow := context.Get(r, "accessToken").(*dal.AccessTokenRow)

	vars := mux.Vars(r)
	key := vars["key"]

	metadataRow, err := dal.NewMetadata(db).DeleteByClusterIDAndKey(nil, accessTokenRow.ClusterID, key)
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

	db := context.Get(r, "db.Core").(*sqlx.DB)

	accessTokenRow := context.Get(r, "accessToken").(*dal.AccessTokenRow)

	vars := mux.Vars(r)
	key := vars["key"]

	metadataRow, err := dal.NewMetadata(db).GetByClusterIDAndKey(nil, accessTokenRow.ClusterID, key)
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
