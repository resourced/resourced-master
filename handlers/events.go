package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/context"
	"github.com/jmoiron/sqlx"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/multidb"
)

func GetApiEventsLine(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accessTokenRow := context.Get(r, "accessTokenRow").(*dal.AccessTokenRow)

	createdInterval := r.URL.Query().Get("CreatedInterval")
	if createdInterval == "" {
		createdInterval = "1 hour"
	}

	tsEventsDB := context.Get(r, "multidb.TSEvents").(*multidb.MultiDB).PickRandom()

	rows, err := dal.NewTSEvent(tsEventsDB).AllLinesByClusterIDAndCreatedFromIntervalForHighchart(nil, accessTokenRow.ClusterID, createdInterval)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	rowsJSONBytes, err := json.Marshal(rows)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(rowsJSONBytes)
}

func GetApiEventsBand(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accessTokenRow := context.Get(r, "accessTokenRow").(*dal.AccessTokenRow)

	createdInterval := r.URL.Query().Get("CreatedInterval")
	if createdInterval == "" {
		createdInterval = "1 hour"
	}

	tsEventsDB := context.Get(r, "multidb.TSEvents").(*multidb.MultiDB).PickRandom()

	rows, err := dal.NewTSEvent(tsEventsDB).AllBandsByClusterIDAndCreatedFromIntervalForHighchart(nil, accessTokenRow.ClusterID, createdInterval)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	rowsJSONBytes, err := json.Marshal(rows)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(rowsJSONBytes)
}

func PostApiEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := context.Get(r, "db.Core").(*sqlx.DB)

	accessTokenRow := context.Get(r, "accessTokenRow").(*dal.AccessTokenRow)

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	var tsEventRow *dal.TSEventRow

	// Asynchronously write time series data to ts_metrics
	id := dal.NewTSEvent(db).NewExplicitID()

	dbs := context.Get(r, "multidb.TSEvents").(*multidb.MultiDB).PickMultipleForWrites()
	for _, db := range dbs {
		tsEventRow, err = dal.NewTSEvent(db).CreateFromJSON(nil, id, accessTokenRow.ClusterID, dataJson)
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

func DeleteApiEventsID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accessTokenRow := context.Get(r, "accessTokenRow").(*dal.AccessTokenRow)

	id, err := getIdFromPath(w, r)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	// Asynchronously write time series data to ts_metrics
	dbs := context.Get(r, "multidb.TSEvents").(*multidb.MultiDB).PickMultipleForWrites()
	for _, db := range dbs {
		_, err = dal.NewTSEvent(db).DeleteByClusterIDAndID(nil, accessTokenRow.ClusterID, id)
		if err != nil {
			libhttp.HandleErrorJson(w, err)
			return
		}
	}

	w.Write([]byte(fmt.Sprintf(`{"Message": "Deleted event", "ID": %v"}`, id)))
}
