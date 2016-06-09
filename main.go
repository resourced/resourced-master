package main

import (
	"encoding/gob"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/alecthomas/kingpin"
	metrics_graphite "github.com/cyberdelia/go-metrics-graphite"
	"github.com/lib/pq"

	"github.com/resourced/resourced-master/application"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libtime"
)

var (
	appConfDirFromEnv  = os.Getenv("RESOURCED_MASTER_CONFIG_DIR")
	appConfDirFromFlag = kingpin.Flag("conf", "Path to config directory").Short('c').String()

	appServerArg  = kingpin.Command("server", "Run resourced-master server.").Default()
	appMigrateArg = kingpin.Command("migrate", "CLI interface for resourced-master database migration.")

	appMigrateUpArg = appMigrateArg.Command("up", "Run all migrations to the most current.").Default()

	appVersion = "4.0.0"
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
						if notification.Channel == "checks_refetch" {
							refetchChecksChan <- true
						} else {
							err := app.HandlePGNotificationPeersAdd(notification)
							if err != nil {
								logrus.Error(err)
							}

							err = app.HandlePGNotificationPeersRemove(notification)
							if err != nil {
								logrus.Error(err)
							}

							err = app.PGNotifyChecksRefetch()
							if err != nil {
								logrus.Error(err)
							}
						}
					}
				}
			}
		}(pgNotificationChan)

		// Register self
		go func() {
			libtime.SleepString("5s")

			err := app.PGNotifyPeersAdd()
			if err != nil {
				logrus.Error(err)
			}
		}()

		// Run all checks
		app.CheckAndRunTriggers(refetchChecksChan)

		// Prune old timeseries data
		go app.PruneAll()

		// Handle OS signals
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)

		go func() {
			for {
				s := <-sigChan
				switch s {
				case syscall.SIGHUP:

				case syscall.SIGINT:
					err := app.PGNotifyPeersRemove()
					if err != nil {
						logrus.Error(err)
					}

				case syscall.SIGTERM:
					err := app.PGNotifyPeersRemove()
					if err != nil {
						logrus.Error(err)
					}

				case syscall.SIGQUIT:
					err := app.PGNotifyPeersRemove()
					if err != nil {
						logrus.Error(err)
					}
				}
			}
		}()

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
