package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/pressly/chi"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/messagebus"
	"github.com/resourced/resourced-master/models/pg"
)

func ApiMetricStreams(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		libhttp.HandleErrorHTML(w, fmt.Errorf("Event streaming is unsupported"), 500)
		return
	}

	bus := r.Context().Value("bus").(*messagebus.MessageBus)

	accessTokenRow := r.Context().Value("accessToken").(*pg.AccessTokenRow)

	errLogger := r.Context().Value("errLogger").(*logrus.Logger)

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
			errLogger.WithFields(logrus.Fields{
				"Method": "ApiMetricStreams",
				"Error":  err,
			}).Errorf("Failed to unmarshal JSON payload")
			continue
		}

		clusterIDInterface, ok := payload["ClusterID"]
		if !ok {
			errLogger.WithFields(logrus.Fields{
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
			errLogger.WithFields(logrus.Fields{
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
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		libhttp.HandleErrorHTML(w, fmt.Errorf("Event streaming is unsupported"), 500)
		return
	}

	dbs := r.Context().Value("dbs").(*config.DBConfig)

	bus := r.Context().Value("bus").(*messagebus.MessageBus)

	accessTokenRow := r.Context().Value("accessToken").(*pg.AccessTokenRow)

	metricID, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	metricRow, err := pg.NewMetric(dbs.Core).GetByID(nil, metricID)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	host := chi.URLParam(r, "host")

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
			}).Errorf("Failed to unmarshal metric-%v JSON payload", metricRow.Key)
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

		// Don't funnel payload if metricRow.ClusterID does not belong.
		if metricRow.ClusterID != accessTokenRow.ClusterID {
			continue
		}

		// Make sure to only return metrics with matching hostname if host value exists.
		if host != "" {
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
