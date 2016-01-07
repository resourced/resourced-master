package main

import (
	"encoding/gob"
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/resourced/resourced-master/application"
	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/dal"
	"github.com/stretchr/graceful"
)

func init() {
	gob.Register(&dal.UserRow{})
	gob.Register(&dal.ClusterRow{})
}

func main() {
	configDir := os.Getenv("RESOURCED_MASTER_CONFIG_DIR")
	if configDir == "" {
		logrus.Fatal("RESOURCED_MASTER_CONFIG_DIR is required")
	}

	// Create default configDir if necessary
	if _, err := os.Stat(configDir); err != nil {
		if os.IsNotExist(err) {
			err := config.NewDefaultConfigs(configDir)
			if err != nil {
				logrus.Fatal(err)
			}
		}
	}

	app, err := application.New(configDir)
	if err != nil {
		logrus.Fatal(err)
	}

	// Migrate up
	errs, ok := app.MigrateUp()
	if !ok {
		for _, err := range errs {
			logrus.Fatal(err)
		}
		os.Exit(1)
	}

	middle, err := app.MiddlewareStruct()
	if err != nil {
		logrus.Fatal(err)
	}

	requestTimeout, err := time.ParseDuration(app.GeneralConfig.RequestTimeout)
	if err != nil {
		logrus.Fatal(err)
	}

	// Register daemon before launching
	_, err = dal.NewDaemon(app.DB).CreateOrUpdate(nil, app.Hostname)
	if err != nil {
		logrus.Fatal(err)
	}

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
}
