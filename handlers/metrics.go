package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/pressly/chi"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/models/pg"
)

func PostMetrics(w http.ResponseWriter, r *http.Request) {
	dbs := r.Context().Value("pg-dbs").(*config.PGDBConfig)

	clusterID, err := getInt64SlugFromPath(w, r, "clusterID")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	key := r.FormValue("Key")

	_, err = pg.NewMetric(dbs.Core).CreateOrUpdate(nil, clusterID, key)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
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
	http.Redirect(w, r, r.Referer(), 301)
}

// DeleteMetricID deletes metrics by ID
func DeleteMetricID(w http.ResponseWriter, r *http.Request) {
	dbs := r.Context().Value("pg-dbs").(*config.PGDBConfig)

	clusterID, err := getInt64SlugFromPath(w, r, "clusterID")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	id, err := getInt64SlugFromPath(w, r, "metricID")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	_, err = pg.NewMetric(dbs.Core).DeleteByClusterIDAndID(nil, clusterID, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	err = pg.NewGraph(dbs.Core).DeleteMetricFromGraphs(nil, clusterID, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

func GetApiTSMetricsByHost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbs := r.Context().Value("pg-dbs").(*config.PGDBConfig)

	errLogger := r.Context().Value("errLogger").(*logrus.Logger)

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

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

	host := chi.URLParam(r, "host")

	metricRow, err := pg.NewMetric(dbs.Core).GetByID(nil, id)
	if err != nil {
		errLogger.WithFields(logrus.Fields{"Error": err}).Error("Failed to fetch metric row")
		libhttp.HandleErrorJson(w, err)
		return
	}

	clusterRow, err := pg.NewCluster(dbs.Core).GetByID(nil, metricRow.ClusterID)
	if err != nil {
		errLogger.WithFields(logrus.Fields{"Error": err}).Error("Failed to fetch cluster row")
		libhttp.HandleErrorJson(w, err)
		return
	}

	deletedFrom := clusterRow.GetDeletedFromUNIXTimestampForSelect("ts_metrics")

	hcMetrics, err := pg.NewTSMetric(dbs.GetTSMetric(metricRow.ClusterID)).AllByMetricIDHostAndRangeForHighchart(nil, metricRow.ClusterID, id, host, from, to, deletedFrom)
	if err != nil {
		errLogger.WithFields(logrus.Fields{"Error": err}).Error("Failed to fetch metrics rows")
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

func GetApiTSMetricsByHost15Min(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbs := r.Context().Value("pg-dbs").(*config.PGDBConfig)

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	qParams := r.URL.Query()

	aggr := qParams.Get("Aggr")
	if aggr == "" {
		aggr = qParams.Get("aggr")
	}

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

	host := chi.URLParam(r, "host")

	metricRow, err := pg.NewMetric(dbs.Core).GetByID(nil, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	clusterRow, err := pg.NewCluster(dbs.Core).GetByID(nil, metricRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	deletedFrom := clusterRow.GetDeletedFromUNIXTimestampForSelect("ts_metrics_aggr_15m")

	hcMetrics, err := pg.NewTSMetricAggr15m(dbs.GetTSMetricAggr15m(metricRow.ClusterID)).AllByMetricIDHostAndRangeForHighchart(nil, metricRow.ClusterID, id, host, from, to, deletedFrom, aggr)
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

	dbs := r.Context().Value("pg-dbs").(*config.PGDBConfig)

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

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	metricRow, err := pg.NewMetric(dbs.Core).GetByID(nil, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	clusterRow, err := pg.NewCluster(dbs.Core).GetByID(nil, metricRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	deletedFrom := clusterRow.GetDeletedFromUNIXTimestampForSelect("ts_metrics")

	hcMetrics, err := pg.NewTSMetric(dbs.GetTSMetric(metricRow.ClusterID)).AllByMetricIDAndRangeForHighchart(nil, metricRow.ClusterID, id, from, to, deletedFrom)
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

	dbs := r.Context().Value("pg-dbs").(*config.PGDBConfig)

	qParams := r.URL.Query()

	aggr := qParams.Get("Aggr")
	if aggr == "" {
		aggr = qParams.Get("aggr")
	}

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

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	metricRow, err := pg.NewMetric(dbs.Core).GetByID(nil, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	clusterRow, err := pg.NewCluster(dbs.Core).GetByID(nil, metricRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	deletedFrom := clusterRow.GetDeletedFromUNIXTimestampForSelect("ts_metrics")

	hcMetrics, err := pg.NewTSMetricAggr15m(dbs.GetTSMetricAggr15m(metricRow.ClusterID)).AllByMetricIDAndRangeForHighchart(nil, metricRow.ClusterID, id, from, to, deletedFrom, aggr)
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
