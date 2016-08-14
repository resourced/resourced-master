package pubsub

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/pub"
	"github.com/go-mangos/mangos/protocol/sub"
	"github.com/go-mangos/mangos/transport/ipc"
	"github.com/go-mangos/mangos/transport/tcp"

	"github.com/resourced/resourced-master/dal"
)

// GetSelfSubscriber returns one subscriber given a map of subscribers.
// It uses hostname or localhost as criteria.
func GetSelfSubscriber(subscribers map[string]*PubSub) (*PubSub, error) {
	var subscriber *PubSub

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	for _, sub := range subscribers {
		if sub.IsSelf(hostname) {
			subscriber = sub
			break
		}
	}

	if subscriber == nil {
		return nil, fmt.Errorf("Failed to find subscriber for hostname: %v", hostname)
	}

	return subscriber, nil
}

// NewPubSocket creates a socket for publishing messages.
func NewPubSocket(url string) (mangos.Socket, error) {
	sock, err := pub.NewSocket()
	if err != nil {
		return nil, err
	}

	sock.AddTransport(ipc.NewTransport())
	sock.AddTransport(tcp.NewTransport())

	err = sock.Listen(url)
	if err != nil {
		return nil, err
	}

	return sock, nil
}

// NewSubSocket creates a socket for subscribing messages.
func NewSubSocket(url string) (mangos.Socket, error) {
	sock, err := sub.NewSocket()
	if err != nil {
		return nil, err
	}
	sock.AddTransport(ipc.NewTransport())
	sock.AddTransport(tcp.NewTransport())
	err = sock.Dial(url)
	if err != nil {
		return nil, err
	}
	return sock, nil
}

// NewPubSub is the constructor to create *PubSub.
func NewPubSub(mode, url string) (*PubSub, error) {
	p := &PubSub{}
	p.Mode = mode
	p.URL = url

	if mode == "pub" {
		sock, err := NewPubSocket(url)
		if err != nil {
			return nil, err
		}
		p.Socket = sock

	} else if mode == "sub" {
		sock, err := NewSubSocket(url)
		if err != nil {
			return nil, err
		}
		p.Socket = sock
	}

	return p, nil
}

type PubSub struct {
	Mode   string
	URL    string
	Socket mangos.Socket
}

// IsSelf checks if current pubsub is running on the host.
func (p *PubSub) IsSelf(selfHostname string) bool {
	if strings.Contains(p.URL, "localhost") {
		return true
	}
	if strings.Contains(p.URL, "127.0.0.1") {
		return true
	}
	if strings.Contains(p.URL, selfHostname) {
		return true
	}

	return false
}

// GetJSONContent returns JSON content from payload in bytes
func (p *PubSub) GetJSONContent(payload string) ([]byte, error) {
	payloadChunks := strings.Split(payload, "|")
	for _, chunk := range payloadChunks {
		keyValue := strings.Split(chunk, ":")

		if keyValue[0] == "type" {
			if keyValue[1] != "json" {
				return nil, fmt.Errorf("Payload type must be json")
			}

			if keyValue[0] == "content" {
				return []byte(keyValue[1]), nil
			}
		}
	}

	return nil, fmt.Errorf("Failed to look for json content from payload")
}

// Publish a plain text message to a topic.
func (p *PubSub) Publish(topic, message string) error {
	if p.Mode != "pub" {
		return fmt.Errorf("Publish method cannot be called if Mode != pub")
	}

	payload := fmt.Sprintf("topic:%s|type:plain|created:%v|content:%s", topic, time.Now().UTC().Unix(), message)
	return p.Socket.Send([]byte(payload))
}

// Publish a JSON message to a topic.
func (p *PubSub) PublishJSON(topic string, jsonBytes []byte) error {
	if p.Mode != "pub" {
		return fmt.Errorf("Publish method cannot be called if Mode != pub")
	}

	payload := fmt.Sprintf("topic:%s|type:json|created:%v|content:%v", topic, time.Now().UTC().Unix(), string(jsonBytes))
	return p.Socket.Send([]byte(payload))
}

// PublishMetricsByHostRow publish many metrics, based on a single host data payload, to the corresponding pubsub pipe.
func (p *PubSub) PublishMetricsByHostRow(hostRow *dal.HostRow, metricsMap map[string]int64) {
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

					err = p.PublishJSON("metric-"+metricKey, metricPayloadJSON)
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

// Subscribe to a topic.
func (p *PubSub) Subscribe(topic string) error {
	if p.Mode != "sub" {
		return fmt.Errorf("Subscribe method cannot be called if Mode != sub")
	}
	return p.Socket.SetOption(mangos.OptionSubscribe, []byte(fmt.Sprintf("topic:%v|", topic)))
}

func (p *PubSub) SubscribeMetric(metricKey string) error {
	return p.Subscribe("metric-" + metricKey)
}
