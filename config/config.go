package config

import (
	"path"

	"github.com/BurntSushi/toml"
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
			From     string
			Subject  string
			Host     string
			Port     int
			Username string
			Password string
		}
	}
}
