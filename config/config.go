package config

import (
	"io/ioutil"
	"path"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/jmoiron/sqlx"
	"github.com/resourced/resourced-master/multidb"
	"github.com/resourced/resourced/libstring"
)

// NewGeneralConfig is the constructor for GeneralConfig.
func NewGeneralConfig(configDir string) (config GeneralConfig, err error) {
	configDir = libstring.ExpandTildeAndEnv(configDir)

	files, err := ioutil.ReadDir(configDir)
	if err != nil {
		return config, err
	}

	contentSlice := make([]string, len(files))
	var generalTomlIndex int

	for i, f := range files {
		if strings.HasSuffix(f.Name(), ".toml") {
			newContent, err := ioutil.ReadFile(path.Join(configDir, f.Name()))
			if err != nil {
				return config, err
			}

			contentSlice[i] = string(newContent)

			if f.Name() == "general.toml" {
				generalTomlIndex = i
			}
		}
	}

	// general.toml must always be first.
	firstContent := contentSlice[0]
	contentSlice[0] = contentSlice[generalTomlIndex]
	contentSlice[generalTomlIndex] = firstContent

	_, err = toml.Decode(strings.Join(contentSlice, "\n"), &config)

	if config.LogLevel == "" {
		config.LogLevel = "info"
	}

	return config, err
}

type EmailConfig struct {
	From          string
	SubjectPrefix string
	Host          string
	Port          int
	Username      string
	Password      string
	Identity      string
}

// GeneralConfig stores all configuration data.
type GeneralConfig struct {
	Addr           string
	LogLevel       string
	DSN            string
	CookieSecret   string
	RequestTimeout string

	HTTPS struct {
		CertFile string
		KeyFile  string
	}

	Watchers struct {
		ListFetchInterval string

		Email *EmailConfig

		SMSEmailGateway map[string]string

		DSNs                  []string
		ReplicationPercentage int
		DataRetention         int
	}

	Metrics struct {
		DSNs                  []string
		ReplicationPercentage int
		DataRetention         int
	}

	Events struct {
		DSNs                  []string
		ReplicationPercentage int
		DataRetention         int
	}

	Email *EmailConfig
}

// NewDBConfig is the constructor for DBConfig.
func NewDBConfig(generalConfig GeneralConfig) (*DBConfig, error) {
	// Set defaults
	if generalConfig.Watchers.ReplicationPercentage <= 0 {
		generalConfig.Watchers.ReplicationPercentage = 100
	}
	if generalConfig.Metrics.ReplicationPercentage <= 0 {
		generalConfig.Metrics.ReplicationPercentage = 100
	}
	if generalConfig.Events.ReplicationPercentage <= 0 {
		generalConfig.Events.ReplicationPercentage = 100
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

	tsEventMultiDB, err := multidb.New(generalConfig.Events.DSNs, generalConfig.Events.ReplicationPercentage)
	if err != nil {
		return nil, err
	}
	conf.TSEvents = tsEventMultiDB

	return conf, nil
}

type DBConfig struct {
	Core       *sqlx.DB
	CoreDSN    string
	TSWatchers *multidb.MultiDB
	TSMetrics  *multidb.MultiDB
	TSEvents   *multidb.MultiDB
}
