package migrator

import (
	_ "github.com/mattes/migrate/driver/postgres"
	"github.com/mattes/migrate/migrate"
	"github.com/resourced/resourced-master/config"
)

func New(generalConfig config.GeneralConfig) *Migrator {
	m := &Migrator{}
	m.AppGeneralConfig = generalConfig

	return m
}

type Migrator struct {
	AppGeneralConfig config.GeneralConfig
}

func (m *Migrator) CoreMigrateUp() ([]error, bool) {
	return migrate.UpSync(m.AppGeneralConfig.DSN, "./migrations/core")
}

func (m *Migrator) TSWatchersMigrateUp() ([]error, bool) {
	for _, dsn := range m.AppGeneralConfig.Watchers.DSNs {
		errs, ok := migrate.UpSync(dsn, "./migrations/ts-watchers")
		if errs != nil && len(errs) > 0 {
			return errs, ok
		}
	}

	return nil, true
}

func (m *Migrator) TSMetricsMigrateUp() ([]error, bool) {
	for _, dsn := range m.AppGeneralConfig.Metrics.DSNs {
		errs, ok := migrate.UpSync(dsn, "./migrations/ts-metrics")
		if errs != nil && len(errs) > 0 {
			return errs, ok
		}
	}

	return nil, true
}
