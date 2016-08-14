package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/pubsub"
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

	subscribers := context.Get(r, "pubsubSubscribers").(map[string]*pubsub.PubSub)

	subscriber, err := pubsub.GetSelfSubscriber(subscribers)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

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
		payloadBytes, err := subscriber.Socket.Recv()
		if err != nil {
			logrus.Error(err)
		}

		payloadString := string(payloadBytes)

		if strings.HasPrefix(payloadString, fmt.Sprintf(`topic:metric-%v`, metricRow.Key)) {
			jsonContent, err := subscriber.GetJSONContent(payloadString)
			if err != nil {
				logrus.Error(err)
			}

			// Make sure to only return metrics with matching hostname if foundHostVar == true.
			if foundHostVar {
				payload := make(map[string]interface{})

				err := json.Unmarshal(payloadBytes, &payload)
				if err != nil {
					logrus.Error(err)
				}

				hostnameInterface, ok := payload["Hostname"]
				if ok {
					if hostnameInterface.(string) == host {
						fmt.Fprintf(w, "data: %v\n\n", string(jsonContent))
						flusher.Flush()
					}
				}

			} else {
				fmt.Fprintf(w, "data: %v\n\n", string(jsonContent))
				flusher.Flush()
			}
		}
	}
}
