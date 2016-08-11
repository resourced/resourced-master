package application

import (
	"fmt"

	"github.com/lib/pq"

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

// NewPGListener creates a new database connection for the purpose of listening events.
func (app *Application) NewPGListener(generalConfig config.GeneralConfig) (*pq.ListenerConn, <-chan *pq.Notification, error) {
	notificationChan := make(chan *pq.Notification)

	listener, err := pq.NewListenerConn(generalConfig.DSN, notificationChan)

	return listener, notificationChan, err
}

// ListenAllPGChannels listens to all predefined channels.
func (app *Application) ListenAllPGChannels(listener *pq.ListenerConn) (bool, error) {
	ok, err := listener.Listen("peers_add")
	if err != nil {
		return ok, err
	}

	ok, err = listener.Listen("peers_remove")
	if err != nil {
		return ok, err
	}

	ok, err = listener.Listen("checks_refetch")
	if err != nil {
		return ok, err
	}

	return ok, err
}

func (app *Application) HandleAllTypesOfNotification(notification *pq.Notification) error {
	err := app.HandlePGNotificationPeersAdd(notification)
	if err != nil {
		return err
	}

	err = app.HandlePGNotificationPeersRemove(notification)
	if err != nil {
		return err
	}

	err = app.PGNotifyChecksRefetch()
	if err != nil {
		return err
	}

	return nil
}

// HandlePGNotificationPeersAdd responds to peers_add channel.
func (app *Application) HandlePGNotificationPeersAdd(notification *pq.Notification) error {
	if notification.Channel == "peers_add" {
		hostAndPort := notification.Extra
		app.Peers.Set(hostAndPort, hostAndPort)
	}

	return nil
}

// HandlePGNotificationPeersRemove responds to peers_remove channel.
func (app *Application) HandlePGNotificationPeersRemove(notification *pq.Notification) error {
	if notification.Channel == "peers_remove" {
		hostAndPort := notification.Extra
		app.Peers.Delete(hostAndPort)
	}

	return nil
}

// PGNotifyPeersAdd sends message to peers_add channel.
func (app *Application) PGNotifyPeersAdd() error {
	_, err := app.DBConfig.Core.Exec(fmt.Sprintf("NOTIFY peers_add, '%v'", app.FullAddr()))
	return err
}

// PGNotifyPeersRemove sends message to peers_remove channel.
func (app *Application) PGNotifyPeersRemove() error {
	_, err := app.DBConfig.Core.Exec(fmt.Sprintf("NOTIFY peers_remove, '%v'", app.FullAddr()))
	return err
}

// PGNotifyChecksRefetch sends message to checks_refetch channel.
func (app *Application) PGNotifyChecksRefetch() error {
	_, err := app.DBConfig.Core.Exec("NOTIFY checks_refetch")
	return err
}
