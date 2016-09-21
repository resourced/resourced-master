package main

import (
	"encoding/gob"
	"net"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/alecthomas/kingpin"
	metrics_graphite "github.com/cyberdelia/go-metrics-graphite"
	gocache "github.com/patrickmn/go-cache"

	"github.com/resourced/resourced-master/application"
	"github.com/resourced/resourced-master/dal"
)

var (
	appConfDirFromEnv  = os.Getenv("RESOURCED_MASTER_CONFIG_DIR")
	appConfDirFromFlag = kingpin.Flag("conf", "Path to config directory").Short('c').String()

	appServerArg  = kingpin.Command("server", "Run resourced-master server.").Default()
	appMigrateArg = kingpin.Command("migrate", "CLI interface for resourced-master database migration.")

	appMigrateUpArg = appMigrateArg.Command("up", "Run all migrations to the most current.").Default()

	appVersion = "4.2.3"
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
		// Create MessageBus
		bus, err := app.NewMessageBus(app.GeneralConfig)
		if err != nil {
			logrus.Fatal(err)
		}
		app.MessageBus = bus

		go app.MessageBus.ManageClients()

		go app.MessageBus.OnReceive(app.MessageBusHandlers())

		// Broadcast heartbeat
		go app.SendHeartbeat()

		go func() {
			// On boot, assign self to peers map and send a message to RefetchChecksChan
			app.Peers.Set(app.FullAddr(), true, gocache.DefaultExpiration)
			app.RefetchChecksChan <- true
		}()

		// Run all checks
		app.CheckAndRunTriggers()

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
