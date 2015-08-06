package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/context"
	"github.com/jmoiron/sqlx"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func GetApiStacks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := context.Get(r, "db").(*sqlx.DB)

	accessTokenRow := context.Get(r, "accessTokenRow").(*dal.AccessTokenRow)

	stacksRow, err := dal.NewStacks(db).GetByClusterID(nil, accessTokenRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	stacksRowJson, err := json.Marshal(stacksRow)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(stacksRowJson)
}

func PostApiStacks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := context.Get(r, "db").(*sqlx.DB)

	accessTokenRow := context.Get(r, "accessTokenRow").(*dal.AccessTokenRow)

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	stacksRow, err := dal.NewStacks(db).CreateOrUpdate(nil, accessTokenRow.ClusterID, dataJson)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	stacksRowJson, err := json.Marshal(stacksRow)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(stacksRowJson)
}
