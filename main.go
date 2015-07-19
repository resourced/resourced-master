package main

import (
	"encoding/gob"
	"github.com/Sirupsen/logrus"
	"github.com/resourced/resourced-master/application"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libenv"
	"github.com/stretchr/graceful"
	"net/http"
	"os"
	"time"
)

func init() {
	gob.Register(&dal.UserRow{})
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
			println(err.Error())
		}
		os.Exit(1)
	}

	middle, err := app.MiddlewareStruct()
	if err != nil {
		logrus.Fatal(err)
	}

	addr := libenv.EnvWithDefault("RESOURCED_MASTER_ADDR", ":55655")
	certFile := libenv.EnvWithDefault("RESOURCED_MASTER_CERT_FILE", "")
	keyFile := libenv.EnvWithDefault("RESOURCED_MASTER_KEY_FILE", "")
	requestTimeoutString := libenv.EnvWithDefault("RESOURCED_MASTER_REQUEST_TIMEOUT", "1s")

	requestTimeout, err := time.ParseDuration(requestTimeoutString)
	if err != nil {
		logrus.Fatal(err)
	}

	srv := &graceful.Server{
		Timeout: requestTimeout,
		Server:  &http.Server{Addr: addr, Handler: middle},
	}

	logrus.WithFields(logrus.Fields{
		"addr": addr,
	}).Info("Running HTTP server")

	if certFile != "" && keyFile != "" {
		srv.ListenAndServeTLS(certFile, keyFile)
	} else {
		srv.ListenAndServe()
	}
}
