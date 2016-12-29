// Package application allows the creation of Application struct.
// There's only one Application per main().
package application

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	gocache "github.com/patrickmn/go-cache"
	"github.com/rcrowley/go-metrics"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/mailer"
	"github.com/resourced/resourced-master/messagebus"
)

// New is the constructor for Application struct.
func New(configDir string) (*Application, error) {
	generalConfig, err := config.NewGeneralConfig(configDir)
	if err != nil {
		return nil, err
	}

	pgDBConfig, err := config.NewPGDBConfig(generalConfig)
	if err != nil {
		return nil, err
	}

	cassandraDBConfig, err := config.NewCassandraDBConfig(generalConfig)
	if err != nil {
		return nil, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	app := &Application{}
	app.Hostname = hostname
	app.GeneralConfig = generalConfig
	app.PGDBConfig = pgDBConfig
	app.CassandraDBConfig = cassandraDBConfig
	app.cookieStore = sessions.NewCookieStore([]byte(app.GeneralConfig.CookieSecret))
	app.Mailers = make(map[string]*mailer.Mailer)
	app.HandlerInstruments = app.NewHandlerInstruments()
	app.LatencyGauges = make(map[string]metrics.Gauge)
	app.MetricsRegistry = app.NewMetricsRegistry(app.HandlerInstruments, app.LatencyGauges)
	app.Peers = gocache.New(1*time.Minute, 10*time.Minute)
	app.RefetchChecksChan = make(chan bool)

	if app.GeneralConfig.Email != nil {
		mailer, err := mailer.New(app.GeneralConfig.Email)
		if err != nil {
			return nil, err
		}
		app.Mailers["GeneralConfig"] = mailer
	}

	if app.GeneralConfig.Checks.Email != nil {
		mailer, err := mailer.New(app.GeneralConfig.Checks.Email)
		if err != nil {
			return nil, err
		}
		app.Mailers["GeneralConfig.Checks"] = mailer
	}

	// Setup loggers
	app.OutLogger = logrus.New()
	app.OutLogger.Out = os.Stdout

	app.ErrLogger = logrus.New()
	app.ErrLogger.Out = os.Stderr

	return app, err
}

// Application is the application object that runs HTTP server.
type Application struct {
	Hostname           string
	GeneralConfig      config.GeneralConfig
	PGDBConfig         *config.PGDBConfig
	CassandraDBConfig  *config.CassandraDBConfig
	cookieStore        *sessions.CookieStore
	Mailers            map[string]*mailer.Mailer
	HandlerInstruments map[string]chan int64
	LatencyGauges      map[string]metrics.Gauge
	MetricsRegistry    metrics.Registry
	MessageBus         *messagebus.MessageBus
	Peers              *gocache.Cache
	RefetchChecksChan  chan bool
	OutLogger          *logrus.Logger
	ErrLogger          *logrus.Logger
	sync.RWMutex
}

func (app *Application) GetContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "GeneralConfig", app.GeneralConfig)
	ctx = context.WithValue(ctx, "PGDBConfig", app.PGDBConfig)
	ctx = context.WithValue(ctx, "CassandraDBConfig", app.CassandraDBConfig)
	ctx = context.WithValue(ctx, "OutLogger", app.OutLogger)
	ctx = context.WithValue(ctx, "ErrLogger", app.ErrLogger)
	ctx = context.WithValue(ctx, "Addr", app.FullAddr())
	ctx = context.WithValue(ctx, "CookieStore", app.cookieStore)

	for key, mailr := range app.Mailers {
		ctx = context.WithValue(ctx, "mailer."+key, mailr)
	}

	ctx = context.WithValue(ctx, "bus", app.MessageBus)

	return ctx
}

func (app *Application) FullAddr() string {
	addr := app.GeneralConfig.Addr
	if strings.HasPrefix(addr, ":") {
		addr = app.Hostname + addr
	}
	if strings.HasPrefix(addr, "localhost") {
		addr = strings.Replace(addr, "localhost", app.Hostname, 1)
	}
	if strings.HasPrefix(addr, "127.0.0.1") {
		addr = strings.Replace(addr, "127.0.0.1", app.Hostname, 1)
	}
	if strings.HasPrefix(addr, "0.0.0.0") {
		addr = strings.Replace(addr, "0.0.0.0", app.Hostname, 1)
	}

	return addr
}

// MigrateUpAll runs all migration files to be up-to-date.
func (app *Application) MigrateAllPG(direction string) error {
	migrationDir := filepath.Join(".", "migrations", "pg", direction)

	files, _ := ioutil.ReadDir(migrationDir)
	for _, f := range files {
		fullFilename := filepath.Join(migrationDir, f.Name())

		sqlBytes, err := ioutil.ReadFile(fullFilename)
		if err != nil {
			return fmt.Errorf("Failed to read migration file. File: %v, Error: %v", fullFilename, err)
		}

		sql := string(sqlBytes)

		// -----------------------------------------------
		// Core
		_, err = app.PGDBConfig.Core.Exec(sql)
		if err != nil {
			return fmt.Errorf("Failed to execute migration file on %v. File: %v, Error: %v", app.GeneralConfig.PostgreSQL.DSN, fullFilename, err)
		}

		// -----------------------------------------------
		// Hosts
		_, err = app.PGDBConfig.Host.Exec(sql)
		if err != nil {
			return fmt.Errorf("Failed to execute migration file on %v. File: %v, Error: %v", app.GeneralConfig.Hosts.PostgreSQL.DSN, fullFilename, err)
		}

		for clusterIDString, dsn := range app.GeneralConfig.Hosts.PostgreSQL.DSNByClusterID {
			clusterID, err := strconv.ParseInt(clusterIDString, 10, 64)
			if err != nil {
				return fmt.Errorf("Failed to execute migration file on %v. File: %v, Error: %v", dsn, fullFilename, err)
			}

			_, err = app.PGDBConfig.GetHost(clusterID).Exec(sql)
			if err != nil {
				return fmt.Errorf("Failed to execute migration file on %v. File: %v, Error: %v", dsn, fullFilename, err)
			}
		}

		// -----------------------------------------------
		// Checks
		_, err = app.PGDBConfig.TSCheck.Exec(sql)
		if err != nil {
			return fmt.Errorf("Failed to execute migration file on %v. File: %v, Error: %v", app.GeneralConfig.Checks.PostgreSQL.DSN, fullFilename, err)
		}

		for clusterIDString, dsn := range app.GeneralConfig.Checks.PostgreSQL.DSNByClusterID {
			clusterID, err := strconv.ParseInt(clusterIDString, 10, 64)
			if err != nil {
				return fmt.Errorf("Failed to execute migration file on %v. File: %v, Error: %v", dsn, fullFilename, err)
			}

			_, err = app.PGDBConfig.GetTSCheck(clusterID).Exec(sql)
			if err != nil {
				return fmt.Errorf("Failed to execute migration file on %v. File: %v, Error: %v", dsn, fullFilename, err)
			}
		}

		// -----------------------------------------------
		// Events
		_, err = app.PGDBConfig.TSEvent.Exec(sql)
		if err != nil {
			return fmt.Errorf("Failed to execute migration file on %v. File: %v, Error: %v", app.GeneralConfig.Events.PostgreSQL.DSN, fullFilename, err)
		}

		for clusterIDString, dsn := range app.GeneralConfig.Events.PostgreSQL.DSNByClusterID {
			clusterID, err := strconv.ParseInt(clusterIDString, 10, 64)
			if err != nil {
				return fmt.Errorf("Failed to execute migration file on %v. File: %v, Error: %v", dsn, fullFilename, err)
			}

			_, err = app.PGDBConfig.GetTSEvent(clusterID).Exec(sql)
			if err != nil {
				return fmt.Errorf("Failed to execute migration file on %v. File: %v, Error: %v", dsn, fullFilename, err)
			}
		}

		// -----------------------------------------------
		// Logs
		_, err = app.PGDBConfig.TSLog.Exec(sql)
		if err != nil {
			return fmt.Errorf("Failed to execute migration file on %v. File: %v, Error: %v", app.GeneralConfig.Logs.PostgreSQL.DSN, fullFilename, err)
		}

		for clusterIDString, dsn := range app.GeneralConfig.Logs.PostgreSQL.DSNByClusterID {
			clusterID, err := strconv.ParseInt(clusterIDString, 10, 64)
			if err != nil {
				return fmt.Errorf("Failed to execute migration file on %v. File: %v, Error: %v", dsn, fullFilename, err)
			}

			_, err = app.PGDBConfig.GetTSLog(clusterID).Exec(sql)
			if err != nil {
				return fmt.Errorf("Failed to execute migration file on %v. File: %v, Error: %v", dsn, fullFilename, err)
			}
		}

		// -----------------------------------------------
		// Metrics
		_, err = app.PGDBConfig.TSMetric.Exec(sql)
		if err != nil {
			return fmt.Errorf("Failed to execute migration file on %v. File: %v, Error: %v", app.GeneralConfig.Metrics.PostgreSQL.DSN, fullFilename, err)
		}

		for clusterIDString, dsn := range app.GeneralConfig.Metrics.PostgreSQL.DSNByClusterID {
			clusterID, err := strconv.ParseInt(clusterIDString, 10, 64)
			if err != nil {
				return fmt.Errorf("Failed to execute migration file on %v. File: %v, Error: %v", dsn, fullFilename, err)
			}

			_, err = app.PGDBConfig.GetTSMetric(clusterID).Exec(sql)
			if err != nil {
				return fmt.Errorf("Failed to execute migration file on %v. File: %v, Error: %v", dsn, fullFilename, err)
			}
		}
	}

	return nil
}

func (app *Application) MigrateAllCassandra(direction string) error {
	migrationDir := filepath.Join(".", "migrations", "cassandra", direction)

	files, _ := ioutil.ReadDir(migrationDir)
	for _, f := range files {
		fullFilename := filepath.Join(migrationDir, f.Name())

		sqlBytes, err := ioutil.ReadFile(fullFilename)
		if err != nil {
			return fmt.Errorf("Failed to read migration file. File: %v, Error: %v", fullFilename, err)
		}

		sql := string(sqlBytes)

		// -----------------------------------------------
		// Core
		err = app.CassandraDBConfig.CoreSession.Query(sql).Exec()
		if err != nil {
			return fmt.Errorf("Failed to execute migration file on %v. File: %v, Error: %v", app.GeneralConfig.Cassandra.Keyspace, fullFilename, err)
		}

		// -----------------------------------------------
		// Metrics
		err = app.CassandraDBConfig.TSMetricSession.Query(sql).Exec()
		if err != nil {
			return fmt.Errorf("Failed to execute migration file on %v. File: %v, Error: %v", app.GeneralConfig.Metrics.Cassandra.Keyspace, fullFilename, err)
		}

		// -----------------------------------------------
		// Logs
		err = app.CassandraDBConfig.TSLogSession.Query(sql).Exec()
		if err != nil {
			return fmt.Errorf("Failed to execute migration file on %v. File: %v, Error: %v", app.GeneralConfig.Logs.Cassandra.Keyspace, fullFilename, err)
		}

		// -----------------------------------------------
		// Events
		err = app.CassandraDBConfig.TSEventSession.Query(sql).Exec()
		if err != nil {
			return fmt.Errorf("Failed to execute migration file on %v. File: %v, Error: %v", app.GeneralConfig.Events.Cassandra.Keyspace, fullFilename, err)
		}

		// -----------------------------------------------
		// Checks
		err = app.CassandraDBConfig.TSCheckSession.Query(sql).Exec()
		if err != nil {
			return fmt.Errorf("Failed to execute migration file on %v. File: %v, Error: %v", app.GeneralConfig.Checks.Cassandra.Keyspace, fullFilename, err)
		}
	}

	return nil
}
