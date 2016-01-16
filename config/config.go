package config

import (
	"path"

	"github.com/BurntSushi/toml"
	"github.com/jmoiron/sqlx"
	"github.com/resourced/resourced/libstring"
)

// NewGeneralConfig is the constructor for GeneralConfig.
func NewGeneralConfig(configDir string) (GeneralConfig, error) {
	configDir = libstring.ExpandTildeAndEnv(configDir)
	fullpath := path.Join(configDir, "general.toml")

	var config GeneralConfig
	_, err := toml.DecodeFile(fullpath, &config)

	if config.LogLevel == "" {
		config.LogLevel = "info"
	}

	return config, err
}

// GeneralConfig stores all configuration data.
type GeneralConfig struct {
	Addr           string
	LogLevel       string
	Hosts          []string
	DSN            string
	CookieSecret   string
	RequestTimeout string

	HTTPS struct {
		CertFile string
		KeyFile  string
	}

	Watchers struct {
		ListFetchInterval string
		DSNs              []string

		Email struct {
			From          string
			SubjectPrefix string
			Host          string
			Port          int
			Username      string
			Password      string
			Identity      string
		}

		SMSEmailGateway map[string]string
	}

	Metrics struct {
		DSNs []string
	}
}

// NewDBConfig is the constructor for DBConfig.
func NewDBConfig(generalConfig GeneralConfig) (*DBConfig, error) {
	conf := &DBConfig{}
	conf.TSWatchers = make([]*sqlx.DB, len(generalConfig.Watchers.DSNs))
	conf.TSWatcherDSNs = make([]string, len(generalConfig.Watchers.DSNs))
	conf.TSMetrics = make([]*sqlx.DB, len(generalConfig.Metrics.DSNs))
	conf.TSMetricDSNs = make([]string, len(generalConfig.Metrics.DSNs))

	coreDB, err := sqlx.Connect("postgres", generalConfig.DSN)
	if err != nil {
		return nil, err
	}
	conf.Core = coreDB
	conf.CoreDSN = generalConfig.DSN

	for i, dsn := range generalConfig.Watchers.DSNs {
		conf.TSWatcherDSNs[i] = dsn

		db, err := sqlx.Connect("postgres", dsn)
		if err != nil {
			return nil, err
		}
		conf.TSWatchers[i] = db
	}
	for i, dsn := range generalConfig.Metrics.DSNs {
		conf.TSMetricDSNs[i] = dsn

		db, err := sqlx.Connect("postgres", dsn)
		if err != nil {
			return nil, err
		}
		conf.TSMetrics[i] = db
	}

	return conf, nil
}

type DBConfig struct {
	Core    *sqlx.DB
	CoreDSN string

	TSWatchers    []*sqlx.DB
	TSWatcherDSNs []string

	TSMetrics    []*sqlx.DB
	TSMetricDSNs []string
}
