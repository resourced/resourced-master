// Package config provides data structures for Application configurations.
package config

import (
	"io/ioutil"
	"path"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/jmoiron/sqlx"

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

// EmailConfig stores all email configuration data.
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
	VIPAddr        string
	VIPProtocol    string

	HTTPS struct {
		CertFile string
		KeyFile  string
	}

	Metrics struct {
		DSN           string
		DataRetention int
	}

	Events struct {
		DSN           string
		DataRetention int
	}

	ExecutorLogs struct {
		DSN           string
		DataRetention int
	}

	Logs struct {
		DSN           string
		DataRetention int
	}

	Checks struct {
		Email *EmailConfig

		SMSEmailGateway map[string]string

		DSN           string
		DataRetention int
	}

	Email *EmailConfig
}

// NewDBConfig connects to all the databases and returns them in DBConfig instance.
func NewDBConfig(generalConfig GeneralConfig) (*DBConfig, error) {
	conf := &DBConfig{}

	db, err := sqlx.Connect("postgres", generalConfig.DSN)
	if err != nil {
		return nil, err
	}
	conf.Core = db

	db, err = sqlx.Connect("postgres", generalConfig.Metrics.DSN)
	if err != nil {
		return nil, err
	}
	conf.TSMetric = db

	db, err = sqlx.Connect("postgres", generalConfig.Events.DSN)
	if err != nil {
		return nil, err
	}
	conf.TSEvent = db

	db, err = sqlx.Connect("postgres", generalConfig.ExecutorLogs.DSN)
	if err != nil {
		return nil, err
	}
	conf.TSExecutorLog = db

	db, err = sqlx.Connect("postgres", generalConfig.Logs.DSN)
	if err != nil {
		return nil, err
	}
	conf.TSLog = db

	db, err = sqlx.Connect("postgres", generalConfig.Checks.DSN)
	if err != nil {
		return nil, err
	}
	conf.TSCheck = db

	return conf, nil
}

// DBConfig stores all database configuration data.
type DBConfig struct {
	Core          *sqlx.DB
	TSMetric      *sqlx.DB
	TSEvent       *sqlx.DB
	TSExecutorLog *sqlx.DB
	TSLog         *sqlx.DB
	TSCheck       *sqlx.DB
}
