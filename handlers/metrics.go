package handlers

import (
	"encoding/json"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"net/http"
	"strconv"

	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/multidb"
)

func PostMetrics(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "db.Core").(*sqlx.DB)

	vars := mux.Vars(r)

	clusterIDString := vars["clusterid"]
	clusterID, err := strconv.ParseInt(clusterIDString, 10, 64)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	key := r.FormValue("Key")

	_, err = dal.NewMetric(db).CreateOrUpdate(nil, clusterID, key)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/", 301)
}

func GetApiTSMetricsByHost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	createdInterval := r.URL.Query().Get("CreatedInterval")
	if createdInterval == "" {
		createdInterval = "1 hour"
	}

	db := context.Get(r, "db.Core").(*sqlx.DB)

	id, err := getIdFromPath(w, r)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	host := mux.Vars(r)["host"]

	metricRow, err := dal.NewMetric(db).GetById(nil, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tsMetricsDB := context.Get(r, "multidb.TSMetrics").(*multidb.MultiDB).PickRandom()

	hcMetrics, err := dal.NewTSMetric(tsMetricsDB).AllByMetricIDHostAndIntervalForHighchart(nil, metricRow.ClusterID, id, host, createdInterval)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	hcMetricsJSON, err := json.Marshal(hcMetrics)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(hcMetricsJSON)
}

func GetApiTSMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	createdInterval := r.URL.Query().Get("CreatedInterval")
	if createdInterval == "" {
		createdInterval = "1 hour"
	}

	db := context.Get(r, "db.Core").(*sqlx.DB)

	id, err := getIdFromPath(w, r)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	metricRow, err := dal.NewMetric(db).GetById(nil, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tsMetricsDB := context.Get(r, "multidb.TSMetrics").(*multidb.MultiDB).PickRandom()

	hcMetrics, err := dal.NewTSMetric(tsMetricsDB).AllByMetricIDAndIntervalForHighchart(nil, metricRow.ClusterID, id, createdInterval)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	hcMetricsJSON, err := json.Marshal(hcMetrics)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(hcMetricsJSON)
}

func GetApiTSMetrics15Min(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	createdInterval := r.URL.Query().Get("CreatedInterval")
	if createdInterval == "" {
		createdInterval = "1 hour"
	}

	db := context.Get(r, "db.Core").(*sqlx.DB)

	id, err := getIdFromPath(w, r)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	metricRow, err := dal.NewMetric(db).GetById(nil, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tsMetricsDB := context.Get(r, "multidb.TSMetrics").(*multidb.MultiDB).PickRandom()

	hcMetrics, err := dal.NewTSMetricAggr15m(tsMetricsDB).AllByMetricIDAndIntervalForHighchart(nil, metricRow.ClusterID, id, createdInterval)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	hcMetricsJSON, err := json.Marshal(hcMetrics)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(hcMetricsJSON)
}

// PostPutDeleteMetricID handles POST, PUT, and DELETE
func PostPutDeleteMetricID(w http.ResponseWriter, r *http.Request) {
	method := r.FormValue("_method")
	if method == "" {
		method = "put"
	}

	if method == "post" || method == "put" {
		PutMetricID(w, r)
	} else if method == "delete" {
		DeleteMetricID(w, r)
	}
}

// PutMetricID is not supported
func PutMetricID(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", 301)
}

// DeleteMetricID deletes metrics by ID
func DeleteMetricID(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "db.Core").(*sqlx.DB)

	vars := mux.Vars(r)

	clusterIDString := vars["clusterid"]
	clusterID, err := strconv.ParseInt(clusterIDString, 10, 64)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	id, err := getIdFromPath(w, r)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	_, err = dal.NewMetric(db).DeleteByClusterIDAndID(nil, clusterID, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	err = dal.NewGraph(db).DeleteMetricFromGraphs(nil, clusterID, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/", 301)
}
