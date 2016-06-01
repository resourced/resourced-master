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

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tsExecutorLogDB := context.Get(r, "db.TSExecutorLog").(*sqlx.DB)

	err = dal.NewTSExecutorLog(tsExecutorLogDB).CreateFromJSON(nil, accessTokenRow.ClusterID, dataJson)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write([]byte(`{"Message": "Success"}`))
}
