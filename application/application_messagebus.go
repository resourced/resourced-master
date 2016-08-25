package application

import (
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	gocache "github.com/patrickmn/go-cache"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/messagebus"
	resourced_wire "github.com/resourced/resourced-wire"
)

// NewMessageBus creates a new MessageBus instance.
func (app *Application) NewMessageBus(generalConfig config.GeneralConfig) (*messagebus.MessageBus, error) {
	bus, err := messagebus.New(generalConfig.MessageBus.URL)
	if err != nil {
		return nil, err
	}

	err = bus.DialOthers(generalConfig.MessageBus.Peers)
	if err != nil {
		return nil, err
	}

	return bus, nil
}

func (app *Application) MessageBusHandlers() map[string]func(msg string) {
	peersHeartbeat := func(msg string) {
		fullAddr := resourced_wire.ParseSingle(msg).PlainContent()
		if strings.Contains(fullAddr, "Error") {
			logrus.WithFields(logrus.Fields{
				"Method": "app.MessageBusHandlers",
				"Error":  fullAddr,
			}).Error("Error when parsing content from peers-heartbeat topic")
		}

		if fullAddr != "" {
			app.Peers.Set(fullAddr, true, gocache.DefaultExpiration)
			app.RefetchChecksChan <- true
		}
	}

	checksRefetch := func(msg string) {
		content := resourced_wire.ParseSingle(msg).PlainContent()
		if strings.Contains(content, "Error") {
			logrus.WithFields(logrus.Fields{
				"Method": "app.MessageBusHandlers",
				"Error":  content,
			}).Error("Error when parsing content from checks-refetch topic")
		}

		app.RefetchChecksChan <- true
	}

	metricStream := func(msg string) {
		content := resourced_wire.ParseSingle(msg).JSONStringContent()
		if strings.Contains(content, "Error") {
			logrus.WithFields(logrus.Fields{
				"Method": "app.MessageBusHandlers",
				"Error":  content,
			}).Error("Error when parsing content from checks-refetch topic")
		}

		for clientChan, _ := range app.MessageBus.Clients {
			clientChan <- content
		}
	}

	return map[string]func(msg string){
		"peers-heartbeat": peersHeartbeat,
		"checks-refetch":  checksRefetch,
		"metric-":         metricStream,
	}
}

// sendHeartbeatOnce payload using the messagebus mechanism
func (app *Application) sendHeartbeatOnce() error {
	return app.MessageBus.Publish("peers-heartbeat", app.FullAddr())
}

// SendHeartbeat every 30 seconds over message bus.
func (app *Application) SendHeartbeat() {
	for range time.Tick(30 * time.Second) {
		err := app.sendHeartbeatOnce()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Method": "app.SendHeartbeat",
				"Error":  err,
			}).Error("Error when sending heartbeat every 30 seconds")
		}
	}
}
