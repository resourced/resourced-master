package config

import (
	"path"

	"github.com/BurntSushi/toml"
	"github.com/jmoiron/sqlx"
	"github.com/resourced/resourced-master/multidb"
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

		DSNs                  []string
		ReplicationPercentage int
	}

	Metrics struct {
		DSNs                  []string
		ReplicationPercentage int
	}
}

// NewDBConfig is the constructor for DBConfig.
func NewDBConfig(generalConfig GeneralConfig) (*DBConfig, error) {
	// Set defaults
	if generalConfig.Watchers.ReplicationPercentage == 0 {
		generalConfig.Watchers.ReplicationPercentage = 100
	}
	if generalConfig.Metrics.ReplicationPercentage == 0 {
		generalConfig.Metrics.ReplicationPercentage = 100
	}

	conf := &DBConfig{}

	coreDB, err := sqlx.Connect("postgres", generalConfig.DSN)
	if err != nil {
		return nil, err
	}
	conf.Core = coreDB
	conf.CoreDSN = generalConfig.DSN

	tsWatcherMultiDB, err := multidb.New(generalConfig.Watchers.DSNs, generalConfig.Watchers.ReplicationPercentage)
	if err != nil {
		return nil, err
	}
	conf.TSWatchers = tsWatcherMultiDB

	tsMetricMultiDB, err := multidb.New(generalConfig.Metrics.DSNs, generalConfig.Metrics.ReplicationPercentage)
	if err != nil {
		return nil, err
	}
	conf.TSMetrics = tsMetricMultiDB

	return conf, nil
}

type DBConfig struct {
	Core       *sqlx.DB
	CoreDSN    string
	TSWatchers *multidb.MultiDB
	TSMetrics  *multidb.MultiDB
}
