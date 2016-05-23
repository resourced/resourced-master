package application

import (
	"github.com/Sirupsen/logrus"
	"github.com/lib/pq"

	"github.com/resourced/resourced-master/config"
)

func (app *Application) NewPGListener(generalConfig config.GeneralConfig) (*pq.ListenerConn, <-chan *pq.Notification, error) {
	notificationChan := make(chan *pq.Notification)

	listener, err := pq.NewListenerConn(generalConfig.DSN, notificationChan)

	return listener, notificationChan, err
}

func (app *Application) ListenAllPGChannels(listener *pq.ListenerConn) (bool, error) {
	ok, err := listener.Listen("peers_add")
	if err != nil {
		return ok, err
	}

	ok, err = listener.Listen("peers_remove")

	return ok, err
}

func (app *Application) HandleAllPGNotifications(notificationChan <-chan *pq.Notification) {
	select {
	case notification := <-notificationChan:
		if notification != nil {
			err := app.HandlePGNotificationPeersAdd(notification)
			if err != nil {
				logrus.Error(err)
			}

			err = app.HandlePGNotificationPeersRemove(notification)
			if err != nil {
				logrus.Error(err)
			}
		}
	}
}

func (app *Application) HandlePGNotificationPeersAdd(notification *pq.Notification) error {
	if notification.Channel == "peers_add" {
		hostAndPort := notification.Extra
		println(hostAndPort)
	}

	return nil
}

func (app *Application) HandlePGNotificationPeersRemove(notification *pq.Notification) error {
	if notification.Channel == "peers_remove" {
		hostAndPort := notification.Extra
		println(hostAndPort)
	}

	return nil
}
