package application

import (
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/lib/pq"
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

func (app *Application) setupInternalSubscriptions() error {
	for _, psub := range app.PubSubSubscribers {
		err := psub.Subscribe("peers-heartbeat")
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *Application) sendHeartbeatOnce() error {
	return app.PubSubPublisher.Publish("peers-heartbeat", app.FullAddr())
}

// SendHeartbeat every minute to all subscribers.
func (app *Application) SendHeartbeat() {
	for range time.Tick(30 * time.Second) {
		err := app.sendHeartbeatOnce()
		if err != nil {
			logrus.Error(err)
		}
	}
}

//
func (app *Application) OnPubSubReceivePayload(url string, subscriber *pubsub.PubSub) {
	if subscriber.Mode != "sub" {
		logrus.Error("Unable to receive message if Mode != sub")
	}

	for {
		payloadBytes, err := subscriber.Socket.Recv()
		if err != nil {
			logrus.Error(err)
		}

		payload := string(payloadBytes)
		println("am i receiving this? " + payload)

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
				println("Am i setting the peers? " + fullAddr)
				app.Peers.Set(fullAddr, true, gocache.DefaultExpiration)
			}
		}
	}
}

// NewPGListener creates a new database connection for the purpose of listening events.
func (app *Application) NewPGListener(generalConfig config.GeneralConfig) (*pq.ListenerConn, <-chan *pq.Notification, error) {
	notificationChan := make(chan *pq.Notification)

	listener, err := pq.NewListenerConn(generalConfig.DSN, notificationChan)

	return listener, notificationChan, err
}

// ListenAllPGChannels listens to all predefined channels.
func (app *Application) ListenAllPGChannels(listener *pq.ListenerConn) (bool, error) {
	ok, err := listener.Listen("checks_refetch")
	if err != nil {
		return ok, err
	}

	return ok, err
}

func (app *Application) HandleAllTypesOfNotification(notification *pq.Notification) error {
	err := app.PGNotifyChecksRefetch()
	if err != nil {
		return err
	}

	return nil
}

// PGNotifyChecksRefetch sends message to checks_refetch channel.
func (app *Application) PGNotifyChecksRefetch() error {
	_, err := app.DBConfig.Core.Exec("NOTIFY checks_refetch")
	return err
}
