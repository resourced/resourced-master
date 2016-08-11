package main

import (
	"encoding/gob"
	"net"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/alecthomas/kingpin"
	metrics_graphite "github.com/cyberdelia/go-metrics-graphite"
	"github.com/lib/pq"

	"github.com/resourced/resourced-master/application"
	"github.com/resourced/resourced-master/dal"
)

var (
	appConfDirFromEnv  = os.Getenv("RESOURCED_MASTER_CONFIG_DIR")
	appConfDirFromFlag = kingpin.Flag("conf", "Path to config directory").Short('c').String()

	appServerArg  = kingpin.Command("server", "Run resourced-master server.").Default()
	appMigrateArg = kingpin.Command("migrate", "CLI interface for resourced-master database migration.")

	appMigrateUpArg = appMigrateArg.Command("up", "Run all migrations to the most current.").Default()

	appVersion = "4.1.1"
)

func init() {
	gob.Register(&dal.UserRow{})
	gob.Register(&dal.ClusterRow{})
}

func main() {
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version(appVersion).Author("Didip Kerabat")
	parsedCLIArgs := kingpin.Parse()

	if appConfDirFromEnv == "" && *appConfDirFromFlag == "" {
		logrus.Fatal("Path to config directory is required. You must set RESOURCED_MASTER_CONFIG_DIR environment variable or -c flag.")
	}

	configDir := appConfDirFromEnv
	if configDir == "" {
		configDir = *appConfDirFromFlag
	}

	app, err := application.New(configDir)
	if err != nil {
		logrus.Fatal(err)
	}

	switch parsedCLIArgs {
	case "server":
		// Listens to all pubsub messages
		for url, subscriber := range app.PubSubSubscribers {
			go app.OnPubSubReceivePayload(url, subscriber)
		}

		// Create a database listener
		pgListener, pgNotificationChan, err := app.NewPGListener(app.GeneralConfig)
		if err != nil {
			logrus.Fatal(err)
		}

		// Listen on all database channels
		_, err = app.ListenAllPGChannels(pgListener)
		if err != nil {
			logrus.Fatal(err)
		}

		refetchChecksChan := make(chan bool)

		// Handle all database notification
		go func(notificationChan <-chan *pq.Notification) {
			for {
				select {
				case notification := <-notificationChan:
					if notification != nil {
						logrus.WithFields(logrus.Fields{"Channel": notification.Channel}).Info("Received notification from PostgreSQL")

						if notification.Channel == "checks_refetch" {
							refetchChecksChan <- true
						} else {
							err := app.HandleAllTypesOfNotification(notification)
							if err != nil {
								logrus.Error(err)
							}
						}
					}
				}
			}
		}(pgNotificationChan)

		// Broadcast heartbeat
		go app.SendHeartbeat()

		// Run all checks
		app.CheckAndRunTriggers(refetchChecksChan)

		// Prune old timeseries data
		go app.PruneAll()

		// Publish metrics to local agent, which is a graphite endpoint.
		addr, err := net.ResolveTCPAddr("tcp", "localhost:"+app.GeneralConfig.LocalAgent.GraphiteTCPPort)
		if err != nil {
			logrus.Fatal(err)
		}
		statsInterval, err := time.ParseDuration(app.GeneralConfig.LocalAgent.ReportMetricsInterval)
		if err != nil {
			logrus.Fatal(err)
		}
		go metrics_graphite.Graphite(app.MetricsRegistry, statsInterval, "ResourcedMaster", addr)

		// Create HTTP server
		srv, err := app.NewHTTPServer()
		if err != nil {
			logrus.Fatal(err)
		}

		if app.GeneralConfig.HTTPS.CertFile != "" && app.GeneralConfig.HTTPS.KeyFile != "" {
			logrus.WithFields(logrus.Fields{"Addr": app.GeneralConfig.Addr}).Info("Running HTTPS server")
			srv.ListenAndServeTLS(app.GeneralConfig.HTTPS.CertFile, app.GeneralConfig.HTTPS.KeyFile)
		} else {
			logrus.WithFields(logrus.Fields{"Addr": app.GeneralConfig.Addr}).Info("Running HTTP server")
			srv.ListenAndServe()
		}

	case "migrate up":
		err := app.MigrateUpAll()
		if err != nil {
			logrus.Fatal(err)
		}
	}
}
