package application

import (
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	gocache "github.com/patrickmn/go-cache"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/pubsub"
)

// NewPublisher creates a new PubSub instance with "pub" mode.
func (app *Application) NewPublisher(generalConfig config.GeneralConfig) (*pubsub.PubSub, error) {
	return pubsub.NewPubSub("pub", "tcp://127.0.0.1:"+generalConfig.PubSub.PublisherPort)
}

// NewSubscribers creates a map of PubSub instances with "sub" mode.
func (app *Application) NewSubscribers(generalConfig config.GeneralConfig) (map[string]*pubsub.PubSub, error) {
	subscribers := make(map[string]*pubsub.PubSub)

	for _, url := range generalConfig.PubSub.SubscriberURLs {
		psub, err := pubsub.NewPubSub("sub", url)
		if err != nil {
			return subscribers, err
		}
		subscribers[url] = psub
	}

	return subscribers, nil
}

// Application uses pubsub for various internal features.
func (app *Application) setupInternalSubscriptions() error {
	for _, psub := range app.PubSubSubscribers {
		err := psub.Subscribe("peers-heartbeat")
		if err != nil {
			return err
		}

		err = psub.Subscribe("checks-refetch")
		if err != nil {
			return err
		}
	}

	return nil
}

// sendHeartbeatOnce payload using the pubsub mechanism
func (app *Application) sendHeartbeatOnce() error {
	return app.PubSubPublisher.Publish("peers-heartbeat", app.FullAddr())
}

// SendHeartbeat every minute to all "peers-heartbeat" subscribers.
func (app *Application) SendHeartbeat() {
	for range time.Tick(30 * time.Second) {
		err := app.sendHeartbeatOnce()
		if err != nil {
			logrus.Error(err)
		}
	}
}

// OnPubSubReceivePeersHeartbeat responds to peers-heartbeat topic.
func (app *Application) OnPubSubReceivePeersHeartbeat(url string, subscriber *pubsub.PubSub) {
	if subscriber.Mode != "sub" {
		logrus.Error("Unable to receive message if Mode != sub")
	}

	for {
		payloadBytes, err := subscriber.Socket.Recv()
		if err != nil {
			logrus.Error(err)
		}

		payload := string(payloadBytes)

		payloadChunks := strings.Split(payload, "|")

		if strings.HasPrefix(payload, "topic:peers-heartbeat") {
			var fullAddr string

			for _, chunk := range payloadChunks {
				if strings.HasPrefix(chunk, "content:") {
					fullAddr = strings.TrimPrefix(chunk, "content:")
					break
				}
			}

			if fullAddr != "" {
				app.Peers.Set(fullAddr, true, gocache.DefaultExpiration)
			}
		}
	}
}

// OnPubSubReceiveChecksRefetch responds to checks-refetch topic.
func (app *Application) OnPubSubReceiveChecksRefetch(url string, subscriber *pubsub.PubSub, refetchChecksChan chan bool) {
	if subscriber.Mode != "sub" {
		logrus.Error("Unable to receive message if Mode != sub")
	}

	for {
		payloadBytes, err := subscriber.Socket.Recv()
		if err != nil {
			logrus.Error(err)
		}

		if strings.HasPrefix(string(payloadBytes), "topic:checks-refetch") {
			refetchChecksChan <- true
		}
	}
}
