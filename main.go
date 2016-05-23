package main

import (
	"encoding/gob"
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/alecthomas/kingpin"
	"github.com/lib/pq"
	_ "github.com/mattes/migrate/driver/postgres"
	"github.com/mattes/migrate/migrate"
	"github.com/stretchr/graceful"

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
)

func init() {
	gob.Register(&dal.UserRow{})
	gob.Register(&dal.ClusterRow{})
}

func main() {
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version("1.0").Author("Didip Kerabat")
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
		middle, err := app.MiddlewareStruct()
		if err != nil {
			logrus.Fatal(err)
		}

		requestTimeout, err := time.ParseDuration(app.GeneralConfig.RequestTimeout)
		if err != nil {
			logrus.Fatal(err)
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

		peersChangedChan := make(chan bool)

		// Handle all database notification
		go func(notificationChan <-chan *pq.Notification) {
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

					peersChangedChan <- true
				}
			}
		}(pgNotificationChan)

		// Register self
		go func() {
			libtime.SleepString("10s")

			err := app.PGNotifyPeersAdd()
			if err != nil {
				logrus.Error(err)
			}
		}()

		// Run all checks
		app.CheckAndRunTriggers(peersChangedChan)

		srv := &graceful.Server{
			Timeout: requestTimeout,
			Server:  &http.Server{Addr: app.GeneralConfig.Addr, Handler: middle},
		}

		if app.GeneralConfig.HTTPS.CertFile != "" && app.GeneralConfig.HTTPS.KeyFile != "" {
			logrus.WithFields(logrus.Fields{"Addr": app.GeneralConfig.Addr}).Info("Running HTTPS server")
			srv.ListenAndServeTLS(app.GeneralConfig.HTTPS.CertFile, app.GeneralConfig.HTTPS.KeyFile)
		} else {
			logrus.WithFields(logrus.Fields{"Addr": app.GeneralConfig.Addr}).Info("Running HTTP server")
			srv.ListenAndServe()
		}

	case "migrate up":
		errs, ok := migrate.UpSync(app.GeneralConfig.DSN, "./migrations/core")
		if !ok {
			logrus.Fatal(errs[0])
		}

		errs, ok = migrate.UpSync(app.GeneralConfig.Checks.DSN, "./migrations/ts-checks")
		if !ok {
			logrus.Fatal(errs[0])
		}

		errs, ok = migrate.UpSync(app.GeneralConfig.Events.DSN, "./migrations/ts-events")
		if !ok {
			logrus.Fatal(errs[0])
		}

		errs, ok = migrate.UpSync(app.GeneralConfig.ExecutorLogs.DSN, "./migrations/ts-executor-logs")
		if !ok {
			logrus.Fatal(errs[0])
		}

		errs, ok = migrate.UpSync(app.GeneralConfig.Logs.DSN, "./migrations/ts-logs")
		if !ok {
			logrus.Fatal(errs[0])
		}

		errs, ok = migrate.UpSync(app.GeneralConfig.Metrics.DSN, "./migrations/ts-metrics")
		if !ok {
			logrus.Fatal(errs[0])
		}
	}
}
