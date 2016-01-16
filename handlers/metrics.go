package handlers

import (
	"encoding/json"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
	"net/http"
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

	hcMetrics, err := dal.NewTSMetric(db).AllByMetricIDHostAndIntervalForHighchart(nil, id, host, createdInterval)
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
