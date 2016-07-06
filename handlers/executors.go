package handlers

import (
	"io/ioutil"
	"net/http"

	"github.com/gorilla/context"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func PostApiExecutors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accessTokenRow := context.Get(r, "accessToken").(*dal.AccessTokenRow)

	dbs := context.Get(r, "dbs").(*config.DBConfig)

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	clusterRow, err := dal.NewCluster(dbs.Core).GetByID(nil, accessTokenRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	deletedFrom := clusterRow.GetDeletedFromUNIXTimestampForInsert("ts_executor_logs")

	err = dal.NewTSExecutorLog(dbs.TSExecutorLog).CreateFromJSON(nil, accessTokenRow.ClusterID, dataJson, deletedFrom)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write([]byte(`{"Message": "Success"}`))
}
