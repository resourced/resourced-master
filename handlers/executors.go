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

func GetApiExecutors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := context.Get(r, "db").(*sqlx.DB)

	accessTokenRow := context.Get(r, "accessTokenRow").(*dal.AccessTokenRow)

	executorRows, err := dal.NewExecutor(db).AllByClusterID(nil, accessTokenRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	executorRowJson, err := json.Marshal(executorRows)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(executorRowJson)
}

func PostApiExecutorsByHostname(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := context.Get(r, "db").(*sqlx.DB)

	accessTokenRow := context.Get(r, "accessTokenRow").(*dal.AccessTokenRow)

	vars := mux.Vars(r)
	hostname := vars["hostname"]

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	executorRow, err := dal.NewExecutor(db).CreateOrUpdate(nil, accessTokenRow.ClusterID, hostname, dataJson)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	executorRowJson, err := json.Marshal(executorRow)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(executorRowJson)
}
