package handlers

import (
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func PostApiExecutors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accessTokenRow := r.Context().Value("accessToken").(*dal.AccessTokenRow)

	dbs := r.Context().Value("dbs").(*config.DBConfig)

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Error": err.Error(),
		}).Error("Failed to read JSON payload")

		libhttp.HandleErrorJson(w, err)
		return
	}

	clusterRow, err := dal.NewCluster(dbs.Core).GetByID(nil, accessTokenRow.ClusterID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Error": err.Error(),
		}).Error("Failed to get cluster by ID before storing executor logs")

		libhttp.HandleErrorJson(w, err)
		return
	}

	deletedFrom := clusterRow.GetDeletedFromUNIXTimestampForInsert("ts_executor_logs")

	err = dal.NewTSExecutorLog(dbs.GetTSExecutorLog(accessTokenRow.ClusterID)).CreateFromJSON(nil, accessTokenRow.ClusterID, dataJson, deletedFrom)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Error": err.Error(),
		}).Error("Failed to insert ts_executor_logs")

		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write([]byte(`{"Message": "Success"}`))
}
