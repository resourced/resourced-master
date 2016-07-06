package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/context"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func GetApiEventsLine(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbs := context.Get(r, "dbs").(*config.DBConfig)

	accessTokenRow := context.Get(r, "accessToken").(*dal.AccessTokenRow)

	qParams := r.URL.Query()

	fromString := qParams.Get("From")
	if fromString == "" {
		fromString = qParams.Get("from")
	}
	from, err := strconv.ParseInt(fromString, 10, 64)
	if err != nil {
		from = -1
	}

	toString := qParams.Get("To")
	if toString == "" {
		toString = qParams.Get("to")
	}
	to, err := strconv.ParseInt(toString, 10, 64)
	if err != nil {
		to = -1
	}

	if from < 0 || to < 0 {
		libhttp.HandleErrorJson(w, errors.New("From or To parameters are missing"))
		return
	}

	clusterRow, err := dal.NewCluster(dbs.Core).GetByID(nil, accessTokenRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	deletedFrom := clusterRow.GetDeletedFromUNIXTimestampForSelect("ts_events")

	rows, err := dal.NewTSEvent(dbs.TSEvent).AllLinesByClusterIDAndCreatedFromRangeForHighchart(nil, accessTokenRow.ClusterID, from, to, deletedFrom)
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

	dbs := context.Get(r, "dbs").(*config.DBConfig)

	accessTokenRow := context.Get(r, "accessToken").(*dal.AccessTokenRow)

	qParams := r.URL.Query()

	fromString := qParams.Get("From")
	if fromString == "" {
		fromString = qParams.Get("from")
	}
	from, err := strconv.ParseInt(fromString, 10, 64)
	if err != nil {
		from = -1
	}

	toString := qParams.Get("To")
	if toString == "" {
		toString = qParams.Get("to")
	}
	to, err := strconv.ParseInt(toString, 10, 64)
	if err != nil {
		to = -1
	}

	if from < 0 || to < 0 {
		libhttp.HandleErrorJson(w, errors.New("From or To parameters are missing"))
		return
	}

	clusterRow, err := dal.NewCluster(dbs.Core).GetByID(nil, accessTokenRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	deletedFrom := clusterRow.GetDeletedFromUNIXTimestampForSelect("ts_events")

	rows, err := dal.NewTSEvent(dbs.TSEvent).AllBandsByClusterIDAndCreatedFromRangeForHighchart(nil, accessTokenRow.ClusterID, from, to, deletedFrom)
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

	dbs := context.Get(r, "dbs").(*config.DBConfig)

	accessTokenRow := context.Get(r, "accessToken").(*dal.AccessTokenRow)

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

	id := dal.NewTSEvent(dbs.Core).NewExplicitID()

	deletedFrom := clusterRow.GetDeletedFromUNIXTimestampForInsert("ts_events")

	tsEventRow, err := dal.NewTSEvent(dbs.TSEvent).CreateFromJSON(nil, id, accessTokenRow.ClusterID, dataJson, deletedFrom)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
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

	dbs := context.Get(r, "dbs").(*config.DBConfig)

	accessTokenRow := context.Get(r, "accessToken").(*dal.AccessTokenRow)

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	_, err = dal.NewTSEvent(dbs.TSEvent).DeleteByClusterIDAndID(nil, accessTokenRow.ClusterID, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write([]byte(fmt.Sprintf(`{"Message": "Deleted event", "ID": %v"}`, id)))
}
