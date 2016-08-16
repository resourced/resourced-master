package application

import (
	"time"

	"github.com/Sirupsen/logrus"
	gocache "github.com/patrickmn/go-cache"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/messagebus"
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
		fullAddr, err := app.MessageBus.GetContent(msg)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Method": "app.MessageBusHandlers",
				"Error":  err,
			}).Error("Error when parsing content from peers-heartbeat topic")
		}

		if fullAddr != "" {
			app.Peers.Set(fullAddr, true, gocache.DefaultExpiration)
		}
	}

	checksRefetch := func(msg string) {
		_, err := app.MessageBus.GetContent(msg)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Method": "app.MessageBusHandlers",
				"Error":  err,
			}).Error("Error when parsing content from checks-refetch topic")
		}

		app.RefetchChecksChan <- true
	}

	metricStream := func(msg string) {
		content, err := app.MessageBus.GetJSONStringContent(msg)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Method": "app.MessageBusHandlers",
				"Error":  err,
			}).Error("Error when parsing content from checks-refetch topic")
		}

		// NOTE: At this point, we are already doubling the message that's received.
		app.MetricStreamChan <- content
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
