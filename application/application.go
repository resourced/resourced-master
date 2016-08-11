// Package application allows the creation of Application struct.
// There's only one Application per main().
package application

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/carbocation/interpose"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	_ "github.com/mattes/migrate/driver/postgres"
	"github.com/mattes/migrate/migrate"
	"github.com/rcrowley/go-metrics"
	"gopkg.in/tylerb/graceful.v1"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/libmap"
	"github.com/resourced/resourced-master/mailer"
	"github.com/resourced/resourced-master/middlewares"
	"github.com/resourced/resourced-master/pubsub"
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
	app.Peers = libmap.NewTSafeMapString(nil)
	app.Mailers = make(map[string]*mailer.Mailer)
	app.HandlerInstruments = app.NewHandlerInstruments()
	app.LatencyGauges = make(map[string]metrics.Gauge)
	app.MetricsRegistry = app.NewMetricsRegistry(app.HandlerInstruments, app.LatencyGauges)

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

	// Create PubSub Publisher
	publisher, err := app.NewPublisher(app.GeneralConfig)
	if err != nil {
		return nil, err
	}
	app.PubSubPublisher = publisher

	// Create PubSub Subscribers
	subscribers, err := app.NewSubscribers(app.GeneralConfig)
	if err != nil {
		return nil, err
	}
	app.PubSubSubscribers = subscribers

	err = app.setupInternalSubscriptions()
	if err != nil {
		return nil, err
	}

	return app, err
}

// Application is the application object that runs HTTP server.
type Application struct {
	Hostname           string
	GeneralConfig      config.GeneralConfig
	DBConfig           *config.DBConfig
	cookieStore        *sessions.CookieStore
	Mailers            map[string]*mailer.Mailer
	Peers              *libmap.TSafeMapString // Peers include self
	HandlerInstruments map[string]chan int64
	LatencyGauges      map[string]metrics.Gauge
	MetricsRegistry    metrics.Registry
	PubSubPublisher    *pubsub.PubSub
	PubSubSubscribers  map[string]*pubsub.PubSub
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

// MiddlewareStruct configures all the middlewares that are in-use for all request handlers.
func (app *Application) MiddlewareStruct() (*interpose.Middleware, error) {
	middle := interpose.New()
	middle.Use(middlewares.SetAddr(app.GeneralConfig.Addr))
	middle.Use(middlewares.SetVIPAddr(app.GeneralConfig.VIPAddr))
	middle.Use(middlewares.SetVIPProtocol(app.GeneralConfig.VIPProtocol))
	middle.Use(middlewares.SetDBs(app.DBConfig))
	middle.Use(middlewares.SetCookieStore(app.cookieStore))
	middle.Use(middlewares.SetMailers(app.Mailers))

	middle.UseHandler(app.mux())

	return middle, nil
}

// NewHTTPServer returns an instance of HTTP server.
func (app *Application) NewHTTPServer() (*graceful.Server, error) {
	// Create HTTP middlewares
	middle, err := app.MiddlewareStruct()
	if err != nil {
		return nil, err
	}

	requestTimeout, err := time.ParseDuration(app.GeneralConfig.RequestShutdownTimeout)
	if err != nil {
		return nil, err
	}

	// Create HTTP server
	srv := &graceful.Server{
		Timeout: requestTimeout,
		Server:  &http.Server{Addr: app.GeneralConfig.Addr, Handler: middle},
	}

	return srv, nil
}

// MigrateUpAll runs all migration files to be up-to-date.
func (app *Application) MigrateUpAll() error {
	errs, ok := migrate.UpSync(app.GeneralConfig.DSN, "./migrations/core")
	if !ok {
		return fmt.Errorf("DSN: %v, Error: %v", app.GeneralConfig.DSN, errs[0])
	}

	errs, ok = migrate.UpSync(app.GeneralConfig.Hosts.DSN, "./migrations/hosts")
	if !ok {
		return fmt.Errorf("DSN: %v, Error: %v", app.GeneralConfig.Hosts.DSN, errs[0])
	}

	for _, dsn := range app.GeneralConfig.Hosts.DSNByClusterID {
		errs, ok = migrate.UpSync(dsn, "./migrations/hosts")
		if !ok {
			return fmt.Errorf("DSN: %v, Error: %v", dsn, errs[0])
		}
	}

	errs, ok = migrate.UpSync(app.GeneralConfig.Checks.DSN, "./migrations/ts-checks")
	if !ok {
		return fmt.Errorf("DSN: %v, Error: %v", app.GeneralConfig.Checks.DSN, errs[0])
	}

	for _, dsn := range app.GeneralConfig.Checks.DSNByClusterID {
		errs, ok = migrate.UpSync(dsn, "./migrations/ts-checks")
		if !ok {
			return fmt.Errorf("DSN: %v, Error: %v", dsn, errs[0])
		}
	}

	errs, ok = migrate.UpSync(app.GeneralConfig.Events.DSN, "./migrations/ts-events")
	if !ok {
		return fmt.Errorf("DSN: %v, Error: %v", app.GeneralConfig.Events.DSN, errs[0])
	}

	for _, dsn := range app.GeneralConfig.Events.DSNByClusterID {
		errs, ok = migrate.UpSync(dsn, "./migrations/ts-events")
		if !ok {
			return fmt.Errorf("DSN: %v, Error: %v", dsn, errs[0])
		}
	}

	errs, ok = migrate.UpSync(app.GeneralConfig.ExecutorLogs.DSN, "./migrations/ts-executor-logs")
	if !ok {
		return fmt.Errorf("DSN: %v, Error: %v", app.GeneralConfig.ExecutorLogs.DSN, errs[0])
	}

	for _, dsn := range app.GeneralConfig.ExecutorLogs.DSNByClusterID {
		errs, ok = migrate.UpSync(dsn, "./migrations/ts-executor-logs")
		if !ok {
			return fmt.Errorf("DSN: %v, Error: %v", dsn, errs[0])
		}
	}

	errs, ok = migrate.UpSync(app.GeneralConfig.Logs.DSN, "./migrations/ts-logs")
	if !ok {
		return fmt.Errorf("DSN: %v, Error: %v", app.GeneralConfig.Logs.DSN, errs[0])
	}

	for _, dsn := range app.GeneralConfig.Logs.DSNByClusterID {
		errs, ok = migrate.UpSync(dsn, "./migrations/ts-logs")
		if !ok {
			return fmt.Errorf("DSN: %v, Error: %v", dsn, errs[0])
		}
	}

	errs, ok = migrate.UpSync(app.GeneralConfig.Metrics.DSN, "./migrations/ts-metrics")
	if !ok {
		return fmt.Errorf("DSN: %v, Error: %v", app.GeneralConfig.Metrics.DSN, errs[0])
	}

	for _, dsn := range app.GeneralConfig.Metrics.DSNByClusterID {
		errs, ok = migrate.UpSync(dsn, "./migrations/ts-metrics")
		if !ok {
			return fmt.Errorf("DSN: %v, Error: %v", dsn, errs[0])
		}
	}

	errs, ok = migrate.UpSync(app.GeneralConfig.MetricsAggr15m.DSN, "./migrations/ts-metrics")
	if !ok {
		return fmt.Errorf("DSN: %v, Error: %v", app.GeneralConfig.MetricsAggr15m.DSN, errs[0])
	}

	for _, dsn := range app.GeneralConfig.MetricsAggr15m.DSNByClusterID {
		errs, ok = migrate.UpSync(dsn, "./migrations/ts-metrics")
		if !ok {
			return fmt.Errorf("DSN: %v, Error: %v", dsn, errs[0])
		}
	}

	return nil
}
