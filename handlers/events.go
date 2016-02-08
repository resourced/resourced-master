package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/context"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/multidb"
)

func PostApiEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accessTokenRow := context.Get(r, "accessTokenRow").(*dal.AccessTokenRow)

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	var tsEventRow *dal.TSEventRow

	// Asynchronously write time series data to ts_metrics
	dbs := context.Get(r, "multidb.TSEvents").(*multidb.MultiDB).PickMultipleForWrites()
	for _, db := range dbs {
		tsEventRow, err = dal.NewTSEvent(db).CreateFromJSON(nil, accessTokenRow.ClusterID, dataJson)
		if err != nil {
			libhttp.HandleErrorJson(w, err)
			return
		}
	}

	tsEventRowJson, err := json.Marshal(tsEventRow)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(tsEventRowJson)
}
