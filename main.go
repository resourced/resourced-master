package main

import (
	"encoding/gob"
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/resourced/resourced-master/application"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libenv"
	"github.com/stretchr/graceful"
)

func init() {
	gob.Register(&dal.UserRow{})
	gob.Register(&dal.ClusterRow{})
}

func main() {
	app, err := application.New()
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

	certFile := libenv.EnvWithDefault("RESOURCED_MASTER_CERT_FILE", "")
	keyFile := libenv.EnvWithDefault("RESOURCED_MASTER_KEY_FILE", "")
	requestTimeoutString := libenv.EnvWithDefault("RESOURCED_MASTER_REQUEST_TIMEOUT", "1s")

	requestTimeout, err := time.ParseDuration(requestTimeoutString)
	if err != nil {
		logrus.Fatal(err)
	}

	srv := &graceful.Server{
		Timeout: requestTimeout,
		Server:  &http.Server{Addr: app.Addr, Handler: middle},
	}

	if certFile != "" && keyFile != "" {
		logrus.WithFields(logrus.Fields{"Addr": app.Addr}).Info("Running HTTPS server")
		srv.ListenAndServeTLS(certFile, keyFile)
	} else {
		logrus.WithFields(logrus.Fields{"Addr": app.Addr}).Info("Running HTTP server")
		srv.ListenAndServe()
	}
}
