package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/models/pg"
	"github.com/resourced/resourced-master/models/shims"
)

func GetApiEventsLine(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accessTokenRow := r.Context().Value("accessToken").(*pg.AccessTokenRow)

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

	clusterRow, err := pg.NewCluster(r.Context()).GetByID(nil, accessTokenRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	rows, err := shims.NewTSEvent(r.Context(), accessTokenRow.ClusterID).AllLinesByClusterIDAndCreatedFromRangeForHighchart(
		accessTokenRow.ClusterID,
		from,
		to,
		clusterRow.GetDeletedFromUNIXTimestampForSelect("ts_events"),
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

	accessTokenRow := r.Context().Value("accessToken").(*pg.AccessTokenRow)

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

	clusterRow, err := pg.NewCluster(r.Context()).GetByID(nil, accessTokenRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	rows, err := shims.NewTSEvent(r.Context(), accessTokenRow.ClusterID).AllBandsByClusterIDAndCreatedFromRangeForHighchart(accessTokenRow.ClusterID, from, to, clusterRow.GetDeletedFromUNIXTimestampForSelect("ts_events"))
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

	accessTokenRow := r.Context().Value("accessToken").(*pg.AccessTokenRow)

	dataJSON, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	clusterRow, err := pg.NewCluster(r.Context()).GetByID(nil, accessTokenRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tsEventRow, err := shims.NewTSEvent(r.Context(), accessTokenRow.ClusterID).CreateFromJSON(
		shims.NewExplicitID(),
		accessTokenRow.ClusterID,
		dataJSON,
		clusterRow.GetDeletedFromUNIXTimestampForInsert("ts_events"),
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

	accessTokenRow := r.Context().Value("accessToken").(*pg.AccessTokenRow)

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	_, err = pg.NewTSEvent(r.Context(), accessTokenRow.ClusterID).DeleteByClusterIDAndID(nil, accessTokenRow.ClusterID, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write([]byte(fmt.Sprintf(`{"Message": "Deleted event", "ID": %v"}`, id)))
}
