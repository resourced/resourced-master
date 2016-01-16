package config

import (
	"path"

	"github.com/BurntSushi/toml"
	"github.com/jmoiron/sqlx"
	"github.com/resourced/resourced-master/multidbs"
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

	coreDB, err := sqlx.Connect("postgres", generalConfig.DSN)
	if err != nil {
		return nil, err
	}
	conf.Core = coreDB
	conf.CoreDSN = generalConfig.DSN

	tsWatcherMultiDBs, err := multidbs.New(generalConfig.Watchers.DSNs, 100)
	if err != nil {
		return nil, err
	}
	conf.TSWatchers = tsWatcherMultiDBs

	tsMetricMultiDBs, err := multidbs.New(generalConfig.Metrics.DSNs, 100)
	if err != nil {
		return nil, err
	}
	conf.TSMetrics = tsMetricMultiDBs

	return conf, nil
}

type DBConfig struct {
	Core       *sqlx.DB
	CoreDSN    string
	TSWatchers *multidbs.MultiDBs
	TSMetrics  *multidbs.MultiDBs
}
