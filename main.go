package main

import (
	"encoding/gob"
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/alecthomas/kingpin"
	_ "github.com/mattes/migrate/driver/postgres"
	"github.com/mattes/migrate/migrate"
	"github.com/stretchr/graceful"

	"github.com/resourced/resourced-master/application"
	"github.com/resourced/resourced-master/dal"
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

	configDir = appConfDirFromEnv
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

		// Register daemon before launching
		_, err = dal.NewDaemon(app.DBConfig.Core).CreateOrUpdate(nil, app.Hostname)
		if err != nil {
			logrus.Fatal(err)
		}

		// Run all checks
		app.CheckAndRunTriggers()

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
