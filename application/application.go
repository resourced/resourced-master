// Package application allows the creation of Application struct.
// There's only one Application per main().
package application

import (
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/carbocation/interpose"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	_ "github.com/mattes/migrate/driver/postgres"
	"github.com/mattes/migrate/migrate"
	"github.com/stretchr/graceful"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libmap"
	"github.com/resourced/resourced-master/libtime"
	"github.com/resourced/resourced-master/mailer"
	"github.com/resourced/resourced-master/middlewares"
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

	return app, err
}

// Application is the application object that runs HTTP server.
type Application struct {
	Hostname      string
	GeneralConfig config.GeneralConfig
	DBConfig      *config.DBConfig
	cookieStore   *sessions.CookieStore
	Mailers       map[string]*mailer.Mailer
	Peers         *libmap.TSafeMapString // Peers include self
}

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

func (app *Application) NewHTTPServer() (*graceful.Server, error) {
	// Create HTTP middlewares
	middle, err := app.MiddlewareStruct()
	if err != nil {
		return nil, err
	}

	requestTimeout, err := time.ParseDuration(app.GeneralConfig.RequestTimeout)
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

func (app *Application) MigrateUpAll() error {
	errs, ok := migrate.UpSync(app.GeneralConfig.DSN, "./migrations/core")
	if !ok {
		return errs[0]
	}

	errs, ok = migrate.UpSync(app.GeneralConfig.Checks.DSN, "./migrations/ts-checks")
	if !ok {
		return errs[0]
	}

	errs, ok = migrate.UpSync(app.GeneralConfig.Events.DSN, "./migrations/ts-events")
	if !ok {
		return errs[0]
	}

	errs, ok = migrate.UpSync(app.GeneralConfig.ExecutorLogs.DSN, "./migrations/ts-executor-logs")
	if !ok {
		return errs[0]
	}

	errs, ok = migrate.UpSync(app.GeneralConfig.Logs.DSN, "./migrations/ts-logs")
	if !ok {
		return errs[0]
	}

	errs, ok = migrate.UpSync(app.GeneralConfig.Metrics.DSN, "./migrations/ts-metrics")
	if !ok {
		return errs[0]
	}

	return nil
}

func (app *Application) PruneAll() {
	for {
		app.PruneTSCheckOnce()
		app.PruneTSMetricOnce()
		app.PruneTSEventOnce()
		app.PruneTSExecutorLogOnce()
		app.PruneTSLogOnce()

		libtime.SleepString("24h")
	}
}

func (app *Application) PruneTSCheckOnce() {
	go func() {
		err := dal.NewTSCheck(app.DBConfig.TSCheck).DeleteByDayInterval(nil, app.GeneralConfig.Checks.DataRetention)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Method":        "TSCheck.DeleteByDayInterval",
				"DataRetention": app.GeneralConfig.Checks.DataRetention,
			}).Error(err)
		}
	}()
}

func (app *Application) PruneTSMetricOnce() {
	go func() {
		err := dal.NewTSMetric(app.DBConfig.TSMetric).DeleteByDayInterval(nil, app.GeneralConfig.Metrics.DataRetention)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Method":        "TSMetric.DeleteByDayInterval",
				"DataRetention": app.GeneralConfig.Metrics.DataRetention,
			}).Error(err)
		}
	}()
}

func (app *Application) PruneTSEventOnce() {
	go func() {
		err := dal.NewTSEvent(app.DBConfig.TSEvent).DeleteByDayInterval(nil, app.GeneralConfig.Events.DataRetention)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Method":        "TSEvent.DeleteByDayInterval",
				"DataRetention": app.GeneralConfig.Events.DataRetention,
			}).Error(err)
		}
	}()
}

func (app *Application) PruneTSExecutorLogOnce() {
	go func() {
		err := dal.NewTSExecutorLog(app.DBConfig.TSExecutorLog).DeleteByDayInterval(nil, app.GeneralConfig.ExecutorLogs.DataRetention)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Method":        "TSExecutorLog.DeleteByDayInterval",
				"DataRetention": app.GeneralConfig.ExecutorLogs.DataRetention,
			}).Error(err)
		}
	}()
}

func (app *Application) PruneTSLogOnce() {
	go func() {
		err := dal.NewTSLog(app.DBConfig.TSLog).DeleteByDayInterval(nil, app.GeneralConfig.Logs.DataRetention)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Method":        "TSLog.DeleteByDayInterval",
				"DataRetention": app.GeneralConfig.Logs.DataRetention,
			}).Error(err)
		}
	}()
}
