package application

import (
	"fmt"
	"strings"

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

func (app *Application) HandlePGNotificationPeersAdd(notification *pq.Notification) error {
	if notification.Channel == "peers_add" {
		hostAndPort := notification.Extra
		app.Peers.Set(hostAndPort, hostAndPort)

		println("Added " + hostAndPort)
	}

	return nil
}

func (app *Application) HandlePGNotificationPeersRemove(notification *pq.Notification) error {
	if notification.Channel == "peers_remove" {
		hostAndPort := notification.Extra
		app.Peers.Delete(hostAndPort)

		println("Removed " + hostAndPort)
	}

	return nil
}

func (app *Application) PGNotifyPeersAdd() error {
	addr := app.GeneralConfig.Addr
	if strings.HasPrefix(addr, ":") {
		addr = app.Hostname + addr
	}

	_, err := app.DBConfig.Core.Exec(fmt.Sprintf("NOTIFY peers_add, '%v'", addr))
	return err
}

func (app *Application) PGNotifyPeersRemove() error {
	addr := app.GeneralConfig.Addr
	if strings.HasPrefix(addr, ":") {
		addr = app.Hostname + addr
	}

	_, err := app.DBConfig.Core.Exec(fmt.Sprintf("NOTIFY peers_remove, '%v'", addr))
	return err
}
