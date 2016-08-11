package pubsub

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/pub"
	"github.com/go-mangos/mangos/protocol/sub"
	"github.com/go-mangos/mangos/transport/ipc"
	"github.com/go-mangos/mangos/transport/tcp"

	"github.com/resourced/resourced-master/dal"
)

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

func NewPubSub(mode, url string) (*PubSub, error) {
	p := &PubSub{}
	p.Mode = mode

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
	Socket mangos.Socket
}

func (p *PubSub) Publish(topic, message string) error {
	if p.Mode != "pub" {
		return fmt.Errorf("Publish method cannot be called if Mode != pub")
	}

	payload := fmt.Sprintf("topic:%s|type:plain|created:%v|content:%s", topic, time.Now().UTC().Unix(), message)
	return p.Socket.Send([]byte(payload))
}

func (p *PubSub) PublishJSON(topic string, jsonBytes []byte) error {
	if p.Mode != "pub" {
		return fmt.Errorf("Publish method cannot be called if Mode != pub")
	}

	payload := fmt.Sprintf("topic:%s|type:json|created:%v|content:%v", topic, time.Now().UTC().Unix(), string(jsonBytes))
	return p.Socket.Send([]byte(payload))
}

func (p *PubSub) Subscribe(topic string) error {
	if p.Mode != "sub" {
		return fmt.Errorf("Subscribe method cannot be called if Mode != sub")
	}
	return p.Socket.SetOption(mangos.OptionSubscribe, []byte(fmt.Sprintf("topic:%v|", topic)))
}

func (p *PubSub) PublishMetricsByHostRow(hostRow *dal.HostRow, metricsMap map[string]int64) {
	// Loop through every host's data and see if they are part of graph metrics.
	// If they are, publish the metric to pubsub pipe.
	for path, data := range hostRow.DataAsFlatKeyValue() {
		for dataKey, value := range data {
			metricKey := path + "." + dataKey

			if metricID, ok := metricsMap[metricKey]; ok {
				// Deserialized JSON number -> interface{} always have float64 as type.
				if trueValueFloat64, ok := value.(float64); ok {
					metricPayload := make(map[string]interface{})
					metricPayload["Key"] = metricKey
					metricPayload["Value"] = trueValueFloat64

					metricPayloadJSON, err := json.Marshal(metricPayload)
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"Method":    "PubSub.PublishMetricsByHostRow",
							"ClusterID": hostRow.ClusterID,
							"MetricID":  metricID,
							"MetricKey": metricKey,
							"Hostname":  hostRow.Hostname,
							"Value":     trueValueFloat64,
							"Error":     err,
						}).Error("Failed to serialize metric for pubsub pipe")
					}

					err = p.PublishJSON("metric-"+metricKey, metricPayloadJSON)
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"Method":    "PubSub.PublishMetricsByHostRow",
							"ClusterID": hostRow.ClusterID,
							"MetricID":  metricID,
							"MetricKey": metricKey,
							"Hostname":  hostRow.Hostname,
							"Value":     trueValueFloat64,
							"Error":     err,
						}).Error("Failed to publish metric to pubsub pipe")
					}
				}
			}
		}
	}
}
