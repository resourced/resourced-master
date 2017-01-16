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
	"github.com/resourced/resourced-master/models/cassandra"
)

var (
	appConfDirFromEnv  = os.Getenv("RESOURCED_MASTER_CONFIG_DIR")
	appConfDirFromFlag = kingpin.Flag("conf", "Path to config directory").Short('c').String()

	appServerArg  = kingpin.Command("server", "Run resourced-master server.").Default()
	appMigrateArg = kingpin.Command("migrate", "CLI interface for resourced-master database migration.")

	appMigratePG    = appMigrateArg.Command("pg", "Run PostgreSQL migrations.")
	appMigratePGCmd = appMigratePG.Arg("command", "PostgreSQL migration commands. Valid choices: up or down").Required().String()

	appMigrateCassandra    = appMigrateArg.Command("cassandra", "Run Cassandra migrations.")
	appMigrateCassandraCmd = appMigrateCassandra.Arg("command", "Cassandra migration commands. Valid choices: up or down").Required().String()
)

func init() {
	gob.Register(&cassandra.UserRow{})
	gob.Register(&cassandra.ClusterRow{})
}

func main() {
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Author("Didip Kerabat")
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

	if app.GeneralConfig.LogLevel == "debug" {
		logrus.SetLevel(logrus.DebugLevel)
	} else if app.GeneralConfig.LogLevel == "error" {
		logrus.SetLevel(logrus.ErrorLevel)
	} else if app.GeneralConfig.LogLevel == "fatal" {
		logrus.SetLevel(logrus.FatalLevel)
	} else if app.GeneralConfig.LogLevel == "panic" {
		logrus.SetLevel(logrus.PanicLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
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

		// Publish metrics to local agent, which is a graphite endpoint.
		go func() {
			statsInterval, err := time.ParseDuration(app.GeneralConfig.LocalAgent.ReportMetricsInterval)
			if err == nil {
				addr, err := net.ResolveTCPAddr("tcp", "localhost:"+app.GeneralConfig.LocalAgent.GraphiteTCPPort)
				if err == nil {
					go metrics_graphite.Graphite(app.MetricsRegistry, statsInterval, "ResourcedMaster", addr)
				}
			}
		}()

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

	case "migrate pg":
		err := app.MigrateAllPG(*appMigratePGCmd)
		if err != nil {
			logrus.Fatal(err)
		}

	case "migrate cassandra":
		err := app.MigrateAllCassandra(*appMigrateCassandraCmd)
		if err != nil {
			logrus.Fatal(err)
		}

	}
}
