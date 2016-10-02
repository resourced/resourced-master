// Package application allows the creation of Application struct.
// There's only one Application per main().
package application

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	_ "github.com/mattes/migrate/driver/postgres"
	"github.com/mattes/migrate/migrate"
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

	dbConfig, err := config.NewDBConfig(generalConfig)
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
	app.DBConfig = dbConfig
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
	DBConfig           *config.DBConfig
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
func (app *Application) MigrateUpAllPG() error {
	errs, ok := migrate.UpSync(app.GeneralConfig.DSN, "./migrations/pg/core")
	if !ok {
		return fmt.Errorf("DSN: %v, Errors: %v", app.GeneralConfig.DSN, errs)
	}

	errs, ok = migrate.UpSync(app.GeneralConfig.Hosts.DSN, "./migrations/pg/hosts")
	if !ok {
		return fmt.Errorf("DSN: %v, Errors: %v", app.GeneralConfig.Hosts.DSN, errs)
	}

	for _, dsn := range app.GeneralConfig.Hosts.DSNByClusterID {
		errs, ok = migrate.UpSync(dsn, "./migrations/pg/hosts")
		if !ok {
			return fmt.Errorf("DSN: %v, Errors: %v", dsn, errs)
		}
	}

	errs, ok = migrate.UpSync(app.GeneralConfig.Checks.DSN, "./migrations/pg/ts-checks")
	if !ok {
		return fmt.Errorf("DSN: %v, Errors: %v", app.GeneralConfig.Checks.DSN, errs)
	}

	for _, dsn := range app.GeneralConfig.Checks.DSNByClusterID {
		errs, ok = migrate.UpSync(dsn, "./migrations/pg/ts-checks")
		if !ok {
			return fmt.Errorf("DSN: %v, Errors: %v", dsn, errs)
		}
	}

	errs, ok = migrate.UpSync(app.GeneralConfig.Events.DSN, "./migrations/pg/ts-events")
	if !ok {
		return fmt.Errorf("DSN: %v, Errors: %v", app.GeneralConfig.Events.DSN, errs)
	}

	for _, dsn := range app.GeneralConfig.Events.DSNByClusterID {
		errs, ok = migrate.UpSync(dsn, "./migrations/pg/ts-events")
		if !ok {
			return fmt.Errorf("DSN: %v, Errors: %v", dsn, errs)
		}
	}

	errs, ok = migrate.UpSync(app.GeneralConfig.Logs.DSN, "./migrations/pg/ts-logs")
	if !ok {
		return fmt.Errorf("DSN: %v, Errors: %v", app.GeneralConfig.Logs.DSN, errs)
	}

	for _, dsn := range app.GeneralConfig.Logs.DSNByClusterID {
		errs, ok = migrate.UpSync(dsn, "./migrations/pg/ts-logs")
		if !ok {
			return fmt.Errorf("DSN: %v, Errors: %v", dsn, errs)
		}
	}

	errs, ok = migrate.UpSync(app.GeneralConfig.Metrics.DSN, "./migrations/pg/ts-metrics")
	if !ok {
		return fmt.Errorf("DSN: %v, Error: %v", app.GeneralConfig.Metrics.DSN, errs[0])
	}

	for _, dsn := range app.GeneralConfig.Metrics.DSNByClusterID {
		errs, ok = migrate.UpSync(dsn, "./migrations/pg/ts-metrics")
		if !ok {
			return fmt.Errorf("DSN: %v, Errors: %v", dsn, errs)
		}
	}

	errs, ok = migrate.UpSync(app.GeneralConfig.MetricsAggr15m.DSN, "./migrations/pg/ts-metrics")
	if !ok {
		return fmt.Errorf("DSN: %v, Errors: %v", app.GeneralConfig.MetricsAggr15m.DSN, errs)
	}

	for _, dsn := range app.GeneralConfig.MetricsAggr15m.DSNByClusterID {
		errs, ok = migrate.UpSync(dsn, "./migrations/pg/ts-metrics")
		if !ok {
			return fmt.Errorf("DSN: %v, Errors: %v", dsn, errs)
		}
	}

	return nil
}
