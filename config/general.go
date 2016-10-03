// Package config provides data structures for Application configurations.
package config

import (
	"io/ioutil"
	"path"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/resourced/resourced-master/libstring"
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

type PostgreSQLPerClusterConfig struct {
	DSN                string
	MaxOpenConnections int64
	DSNByClusterID     map[string]string
}

// GeneralConfig stores all configuration data.
type GeneralConfig struct {
	Addr                    string
	LogLevel                string
	CookieSecret            string
	RequestShutdownTimeout  string
	VIPAddr                 string
	VIPProtocol             string
	EnablePeriodicPruneJobs bool
	JustAPI                 bool

	PostgreSQL struct {
		DSN                string
		MaxOpenConnections int64
	}

	LocalAgent struct {
		GraphiteTCPPort       string
		ReportMetricsInterval string
	}

	RateLimiters struct {
		PostSignup int
		GeneralAPI int
	}

	HTTPS struct {
		CertFile string
		KeyFile  string
	}

	MessageBus struct {
		URL   string
		Peers []string
	}

	Hosts struct {
		PostgreSQL PostgreSQLPerClusterConfig
	}

	Metrics struct {
		PostgreSQL    PostgreSQLPerClusterConfig
		DataRetention int
	}

	MetricsAggr15m struct {
		PostgreSQL    PostgreSQLPerClusterConfig
		DataRetention int
	}

	Events struct {
		PostgreSQL    PostgreSQLPerClusterConfig
		DataRetention int
	}

	Logs struct {
		PostgreSQL    PostgreSQLPerClusterConfig
		DataRetention int
	}

	Checks struct {
		Email *EmailConfig

		SMSEmailGateway map[string]string

		PostgreSQL PostgreSQLPerClusterConfig

		DataRetention int
	}

	Email *EmailConfig
}
