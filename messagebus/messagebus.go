package messagebus

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/bus"
	"github.com/go-mangos/mangos/transport/ipc"
	"github.com/go-mangos/mangos/transport/tcp"

	"github.com/resourced/resourced-master/dal"
)

func New(url string) (*MessageBus, error) {
	mb := &MessageBus{}
	mb.URL = url

	sock, err := bus.NewSocket()
	if err != nil {
		return nil, err
	}

	sock.AddTransport(ipc.NewTransport())
	sock.AddTransport(tcp.NewTransport())

	err = sock.Listen(url)
	if err != nil {
		return nil, err
	}

	mb.Socket = sock

	return mb, nil
}

type MessageBus struct {
	URL    string
	Socket mangos.Socket
}

func (mb *MessageBus) DialOthers(urls []string) error {
	for _, url := range urls {
		err := mb.Socket.Dial(url)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mb *MessageBus) GetContent(payload string) (string, error) {
	payloadChunks := strings.Split(payload, "|")
	for _, chunk := range payloadChunks {
		keyValue := strings.Split(chunk, ":")

		if keyValue[0] == "content" {
			return keyValue[1], nil
		}
	}

	return "", fmt.Errorf("Failed to look for content from payload")
}

func (mb *MessageBus) GetJSONStringContent(payload string) (string, error) {
	if !strings.Contains(payload, "type:json") {
		return "", fmt.Errorf("Payload type is not JSON")
	}

	payloadChunks := strings.Split(payload, "|")
	for _, chunk := range payloadChunks {
		if strings.HasPrefix(chunk, "content:") {
			return strings.TrimPrefix(chunk, "content:"), nil
		}
	}

	return "", fmt.Errorf("Failed to look for content from payload")
}

// Publish a plain text message to a topic.
func (mb *MessageBus) Publish(topic, message string) error {
	payload := fmt.Sprintf("topic:%s|type:plain|created:%v|content:%s", topic, time.Now().UTC().Unix(), message)
	return mb.Socket.Send([]byte(payload))
}

// Publish a JSON message to a topic.
func (mb *MessageBus) PublishJSON(topic string, jsonBytes []byte) error {
	payload := fmt.Sprintf("topic:%s|type:json|created:%v|content:%v", topic, time.Now().UTC().Unix(), string(jsonBytes))
	return mb.Socket.Send([]byte(payload))
}

// PublishMetricsByHostRow publish many metrics, based on a single host data payload.
func (mb *MessageBus) PublishMetricsByHostRow(hostRow *dal.HostRow, metricsMap map[string]int64) {
	// Loop through every host's data and see if they are part of graph metrics.
	// If they are, publish the metric to pubsub pipe.
	for path, data := range hostRow.DataAsFlatKeyValue() {
		for dataKey, value := range data {
			metricKey := path + "." + dataKey

			if metricID, ok := metricsMap[metricKey]; ok {
				// Deserialized JSON number -> interface{} always have float64 as type.
				if trueValueFloat64, ok := value.(float64); ok {
					metricPayload := logrus.Fields{
						"ClusterID": hostRow.ClusterID,
						"MetricID":  metricID,
						"MetricKey": metricKey,
						"Hostname":  hostRow.Hostname,
						"Value":     trueValueFloat64,
					}

					metricPayloadJSON, err := json.Marshal(metricPayload)
					if err != nil {
						metricPayload["Method"] = "PubSub.PublishMetricsByHostRow"
						metricPayload["Error"] = err
						logrus.WithFields(metricPayload).Error("Failed to serialize metric for pubsub pipe")
					}

					println("publishing metric-" + metricKey)
					println(string(metricPayloadJSON))

					err = mb.PublishJSON("metric-"+metricKey, metricPayloadJSON)
					if err != nil {
						metricPayload["Method"] = "PubSub.PublishMetricsByHostRow"
						metricPayload["Error"] = err
						logrus.WithFields(metricPayload).Error("Failed to publish metric to pubsub pipe")
					}
				}
			}
		}
	}
}

// OnReceive handles various different type of payload based on topic.
func (mb *MessageBus) OnReceive(handlers map[string]func(msg string)) {
	for {
		msgBytes, err := mb.Socket.Recv()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Error": err,
			}).Error("Error when receiving message from bus")
		}

		msg := string(msgBytes)

		for topic, handler := range handlers {
			println("topic: " + topic)
			if strings.HasPrefix(msg, "topic:"+topic) {
				go handler(msg)
			}
		}
	}
}
