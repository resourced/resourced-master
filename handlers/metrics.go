package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/pressly/chi"

	"github.com/resourced/resourced-master/contexthelper"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/models/pg"
	"github.com/resourced/resourced-master/models/shims"
)

func PostMetrics(w http.ResponseWriter, r *http.Request) {
	clusterID, err := getInt64SlugFromPath(w, r, "clusterID")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	key := r.FormValue("Key")

	_, err = pg.NewMetric(r.Context()).CreateOrUpdate(nil, clusterID, key)
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

	_, err = pg.NewMetric(r.Context()).DeleteByClusterIDAndID(nil, clusterID, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	err = pg.NewGraph(r.Context()).DeleteMetricFromGraphs(nil, clusterID, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

func GetApiTSMetricsByHost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	errLogger, err := contexthelper.GetLogger(r.Context(), "ErrLogger")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	qParams := r.URL.Query()

	from, err := strconv.ParseInt(qParams.Get("from"), 10, 64)
	if err != nil {
		from = -1
	}

	to, err := strconv.ParseInt(qParams.Get("to"), 10, 64)
	if err != nil {
		to = -1
	}

	downsample, err := strconv.ParseInt(qParams.Get("downsample"), 10, 64)
	if err != nil {
		downsample = -1
	}

	if from < 0 || to < 0 {
		libhttp.HandleErrorJson(w, errors.New("from or to parameters are missing"))
		return
	}

	host := chi.URLParam(r, "host")

	metricRow, err := pg.NewMetric(r.Context()).GetByID(nil, id)
	if err != nil {
		errLogger.WithFields(logrus.Fields{"Error": err}).Error("Failed to fetch metric row")
		libhttp.HandleErrorJson(w, err)
		return
	}

	clusterRow, err := pg.NewCluster(r.Context()).GetByID(nil, metricRow.ClusterID)
	if err != nil {
		errLogger.WithFields(logrus.Fields{"Error": err}).Error("Failed to fetch cluster row")
		libhttp.HandleErrorJson(w, err)
		return
	}

	shimsTSMetric := shims.NewTSMetric(r.Context(), metricRow.ClusterID)

	hcMetrics, err := shimsTSMetric.AllByMetricIDHostAndRangeForHighchart(metricRow.ClusterID, id, host, from, to, clusterRow.GetDeletedFromUNIXTimestampForSelect("ts_metrics"), downsample)
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

func GetApiTSMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	errLogger, err := contexthelper.GetLogger(r.Context(), "ErrLogger")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

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

	downsample, err := strconv.ParseInt(qParams.Get("downsample"), 10, 64)
	if err != nil {
		downsample = -1
	}

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	metricRow, err := pg.NewMetric(r.Context()).GetByID(nil, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	clusterRow, err := pg.NewCluster(r.Context()).GetByID(nil, metricRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	shimsTSMetric := shims.NewTSMetric(r.Context(), metricRow.ClusterID)

	hcMetrics, err := shimsTSMetric.AllByMetricIDAndRangeForHighchart(metricRow.ClusterID, id, from, to, clusterRow.GetDeletedFromUNIXTimestampForSelect("ts_metrics"), downsample)
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
