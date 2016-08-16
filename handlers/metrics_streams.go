package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func ApiMetricStreams(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		libhttp.HandleErrorHTML(w, fmt.Errorf("Event streaming is unsupported"), 500)
		return
	}

	dbs := context.Get(r, "dbs").(*config.DBConfig)

	metricStreamChan := context.Get(r, "metricStreamChan").(chan string)

	// TODO
	// Make sure we deny user who has no access to metricID.

	metricID, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	metricRow, err := dal.NewMetric(dbs.Core).GetByID(nil, metricID)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	host, foundHostVar := mux.Vars(r)["host"]

	for {
		jsonContentString := <-metricStreamChan

		// Make sure to only return metrics with matching hostname if foundHostVar == true.
		if foundHostVar {
			payload := make(map[string]interface{})

			err := json.Unmarshal([]byte(jsonContentString), &payload)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"Method": "ApiMetricStreams",
					"Error":  err,
				}).Errorf("Failed to metric-%v JSON payload", metricRow.Key)
			}

			hostnameInterface, ok := payload["Hostname"]
			if ok {
				if hostnameInterface.(string) == host {
					fmt.Fprintf(w, "data: %v\n\n", jsonContentString)
					flusher.Flush()
				}
			}

		} else {
			fmt.Fprintf(w, "data: %v\n\n", jsonContentString)
			flusher.Flush()
		}
	}
}
