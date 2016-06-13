package handlers

import (
	"io/ioutil"
	"net/http"

	"github.com/gorilla/context"
	"github.com/jmoiron/sqlx"

	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func PostApiExecutors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accessTokenRow := context.Get(r, "accessToken").(*dal.AccessTokenRow)

	db := context.Get(r, "db.Core").(*sqlx.DB)

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	clusterRow, err := dal.NewCluster(db).GetByID(nil, accessTokenRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	deletedFrom := clusterRow.GetDeletedFromUNIXTimestampForInsert("ts_executor_logs")

	tsExecutorLogDB := context.Get(r, "db.TSExecutorLog").(*sqlx.DB)

	err = dal.NewTSExecutorLog(tsExecutorLogDB).CreateFromJSON(nil, accessTokenRow.ClusterID, dataJson, deletedFrom)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write([]byte(`{"Message": "Success"}`))
}
