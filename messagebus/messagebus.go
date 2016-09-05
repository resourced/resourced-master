package messagebus

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/bus"
	"github.com/go-mangos/mangos/transport/ipc"
	"github.com/go-mangos/mangos/transport/tcp"

	resourced_wire "github.com/resourced/resourced-wire"
)

type IHostRow interface {
	DataAsFlatKeyValue() map[string]map[string]interface{}
	GetClusterID() int64
	GetHostname() string
}

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
	mb.Clients = make(map[chan string]bool)
	mb.NewClientChan = make(chan (chan string))
	mb.CloseClientChan = make(chan (chan string))

	return mb, nil
}

type MessageBus struct {
	URL             string
	Socket          mangos.Socket
	Clients         map[chan string]bool
	NewClientChan   chan chan string
	CloseClientChan chan chan string
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

func (mb *MessageBus) ManageClients() {
	for {
		select {
		case newChan := <-mb.NewClientChan:
			mb.Clients[newChan] = true

		case closeChan := <-mb.CloseClientChan:
			delete(mb.Clients, closeChan)
			close(closeChan)
		}
	}
}

// Publish a plain text message to a topic.
func (mb *MessageBus) Publish(topic, message string) error {
	wire := resourced_wire.Wire{
		Topic:   topic,
		Type:    "plain",
		Created: time.Now().UTC().Unix(),
		Content: message,
	}
	return mb.Socket.Send([]byte(wire.EncodePlain()))
}

// Publish a JSON message to a topic.
func (mb *MessageBus) PublishJSON(topic string, jsonBytes []byte) error {
	wire := resourced_wire.Wire{
		Topic:   topic,
		Type:    "json",
		Created: time.Now().UTC().Unix(),
		Content: string(jsonBytes),
	}
	return mb.Socket.Send([]byte(wire.EncodeJSON()))
}

// PublishMetricsByHostRow publish many metrics, based on a single host data payload.
func (mb *MessageBus) PublishMetricsByHostRow(hostRow IHostRow, metricsMap map[string]int64) {
	// Loop through every host's data and see if they are part of graph metrics.
	// If they are, publish the metric to message bus.
	for path, data := range hostRow.DataAsFlatKeyValue() {
		for dataKey, value := range data {
			metricKey := path + "." + dataKey

			if metricID, ok := metricsMap[metricKey]; ok {
				// Deserialized JSON number -> interface{} always have float64 as type.
				if trueValueFloat64, ok := value.(float64); ok {
					metricPayload := logrus.Fields{
						"ClusterID":          hostRow.GetClusterID(),
						"MetricID":           metricID,
						"MetricKey":          metricKey,
						"Hostname":           hostRow.GetHostname(),
						"Value":              trueValueFloat64,
						"CreatedMillisecond": time.Now().UTC().UnixNano() / int64(time.Millisecond),
					}

					metricPayloadJSON, err := json.Marshal(metricPayload)
					if err != nil {
						metricPayload["Method"] = "PubSub.PublishMetricsByHostRow"
						metricPayload["Error"] = err
						logrus.WithFields(metricPayload).Error("Failed to serialize metric for message bus")
					}

					err = mb.PublishJSON("metric-"+metricKey, metricPayloadJSON)
					if err != nil {
						metricPayload["Method"] = "PubSub.PublishMetricsByHostRow"
						metricPayload["Error"] = err
						logrus.WithFields(metricPayload).Error("Failed to publish metric to message bus")
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

		// NOTE: At this point, we already received duplicate message.

		for topic, handler := range handlers {
			if strings.HasPrefix(msg, "topic:"+topic) {
				go handler(msg)
			}
		}
	}
}
