package main

import (
	"encoding/gob"
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
		println(err.Error())
		os.Exit(1)
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
		println(err.Error())
		os.Exit(1)
	}

	serverAddress := libenv.EnvWithDefault("RESOURCED_MASTER_ADDR", ":55655")
	certFile := libenv.EnvWithDefault("RESOURCED_MASTER_CERT_FILE", "")
	keyFile := libenv.EnvWithDefault("RESOURCED_MASTER_KEY_FILE", "")
	requestTimeoutString := libenv.EnvWithDefault("RESOURCED_MASTER_REQUEST_TIMEOUT", "1s")

	requestTimeout, err := time.ParseDuration(requestTimeoutString)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	srv := &graceful.Server{
		Timeout: requestTimeout,
		Server:  &http.Server{Addr: serverAddress, Handler: middle},
	}

	if certFile != "" && keyFile != "" {
		srv.ListenAndServeTLS(certFile, keyFile)
	} else {
		srv.ListenAndServe()
	}
}
