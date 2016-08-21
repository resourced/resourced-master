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
	"github.com/resourced/resourced-master/messagebus"
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

	bus := context.Get(r, "bus").(*messagebus.MessageBus)

	accessTokenRow := context.Get(r, "accessToken").(*dal.AccessTokenRow)

	// Create a new channel for this connected client.
	newClientChan := make(chan string)

	// Inform message bus of this new channel.
	bus.NewClientChan <- newClientChan

	// Listen to the closing of the http connection via the CloseNotifier,
	// and close the corresponding channel.
	connClosedChan := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-connClosedChan
		bus.CloseClientChan <- newClientChan
	}()

	for {
		jsonContentString, isOpen := <-newClientChan
		if !isOpen {
			// If newClientChan was closed, client has disconnected.
			break
		}

		payload := make(map[string]interface{})

		err := json.Unmarshal([]byte(jsonContentString), &payload)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Method": "ApiMetricStreams",
				"Error":  err,
			}).Errorf("Failed to unmarshal JSON payload")
			continue
		}

		clusterIDInterface, ok := payload["ClusterID"]
		if !ok {
			logrus.WithFields(logrus.Fields{
				"Method": "ApiMetricStreams",
			}).Errorf("ClusterID is missing from payload")
			continue
		}

		clusterID := int64(clusterIDInterface.(float64))

		// Don't funnel payload if ClusterID is incorrect.
		if clusterID != accessTokenRow.ClusterID {
			continue
		}

		// Emit payload by MetricID
		metricIDInterface, ok := payload["MetricID"]
		if !ok {
			logrus.WithFields(logrus.Fields{
				"Method": "ApiMetricStreams",
			}).Errorf("MetricID is missing from payload")
			continue
		}

		metricID := int64(metricIDInterface.(float64))

		fmt.Fprintf(w, "event: metric|%v\n", metricID)
		fmt.Fprintf(w, "data: %v\n\n", jsonContentString)

		// Emit payload by MetricID and Hostname
		hostnameInterface, ok := payload["Hostname"]
		if ok {
			fmt.Fprintf(w, "event: metric|%v|host|%v\n", metricID, hostnameInterface.(string))
			fmt.Fprintf(w, "data: %v\n\n", jsonContentString)
		}

		flusher.Flush()
	}

}

func ApiMetricIDStreams(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		libhttp.HandleErrorHTML(w, fmt.Errorf("Event streaming is unsupported"), 500)
		return
	}

	dbs := context.Get(r, "dbs").(*config.DBConfig)

	bus := context.Get(r, "bus").(*messagebus.MessageBus)

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

	// Create a new channel for this connected client.
	newClientChan := make(chan string)

	// Inform message bus of this new channel.
	bus.NewClientChan <- newClientChan

	// Listen to the closing of the http connection via the CloseNotifier,
	// and close the corresponding channel.
	connClosedChan := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-connClosedChan
		bus.CloseClientChan <- newClientChan
	}()

	for {
		jsonContentString, isOpen := <-newClientChan
		if !isOpen {
			// If newClientChan was closed, client has disconnected.
			break
		}

		// Make sure to only return metrics with matching hostname if foundHostVar == true.
		if foundHostVar {
			payload := make(map[string]interface{})

			err := json.Unmarshal([]byte(jsonContentString), &payload)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"Method": "ApiMetricIDStreams",
					"Error":  err,
				}).Errorf("Failed to unmarshal metric-%v JSON payload", metricRow.Key)
			}

			hostnameInterface, ok := payload["Hostname"]
			if ok {
				if hostnameInterface.(string) == host {
					fmt.Fprintf(w, "event: metric|%v|host|%v\n", metricID, host)
					fmt.Fprintf(w, "data: %v\n\n", jsonContentString)
					flusher.Flush()
				}
			}

		} else {
			fmt.Fprintf(w, "event: metric|%v\n", metricID)
			fmt.Fprintf(w, "data: %v\n\n", jsonContentString)
			flusher.Flush()
		}
	}
}
