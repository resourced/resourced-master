package handlers

import (
	"encoding/json"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"net/http"

	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/multidb"
)

func PostMetrics(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "db.Core").(*sqlx.DB)

	clusterID, err := getIdFromPath(w, r)
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
