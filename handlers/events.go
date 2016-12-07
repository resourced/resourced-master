package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/models/cassandra"
	"github.com/resourced/resourced-master/models/shims"
)

func GetApiEventsLine(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accessTokenRow := r.Context().Value("accessToken").(*cassandra.AccessTokenRow)

	qParams := r.URL.Query()

	from, err := strconv.ParseInt(qParams.Get("from"), 10, 64)
	if err != nil {
		from = -1
	}

	to, err := strconv.ParseInt(qParams.Get("to"), 10, 64)
	if err != nil {
		to = -1
	}

	if from < 0 || to < 0 {
		libhttp.HandleErrorJson(w, errors.New("From or To parameters are missing"))
		return
	}

	rows, err := cassandra.NewTSEvent(r.Context()).AllLinesByClusterIDAndCreatedFromRangeForHighchart(
		accessTokenRow.ClusterID,
		from,
		to,
	)
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

	accessTokenRow := r.Context().Value("accessToken").(*cassandra.AccessTokenRow)

	qParams := r.URL.Query()

	from, err := strconv.ParseInt(qParams.Get("from"), 10, 64)
	if err != nil {
		from = -1
	}

	to, err := strconv.ParseInt(qParams.Get("to"), 10, 64)
	if err != nil {
		to = -1
	}

	if from < 0 || to < 0 {
		libhttp.HandleErrorJson(w, errors.New("From or To parameters are missing"))
		return
	}

	rows, err := cassandra.NewTSEvent(r.Context()).AllBandsByClusterIDAndCreatedFromRangeForHighchart(accessTokenRow.ClusterID, from, to)
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

	accessTokenRow := r.Context().Value("accessToken").(*cassandra.AccessTokenRow)

	dataJSON, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	clusterRow, err := cassandra.NewCluster(r.Context()).GetByID(accessTokenRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tsEventRow, err := cassandra.NewTSEvent(r.Context()).CreateFromJSON(
		shims.NewExplicitID(),
		accessTokenRow.ClusterID,
		dataJSON,
		clusterRow.GetTTLDurationForInsert("ts_events"),
	)
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

	accessTokenRow := r.Context().Value("accessToken").(*cassandra.AccessTokenRow)

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	err = cassandra.NewTSEvent(r.Context()).DeleteByClusterIDAndID(accessTokenRow.ClusterID, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write([]byte(fmt.Sprintf(`{"Message": "Deleted event", "ID": %v"}`, id)))
}
