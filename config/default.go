package config

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/Sirupsen/logrus"
	"github.com/resourced/resourced/libstring"
)

// NewDefaultConfigs provide default config setup.
// This function is called on first boot.
func NewDefaultConfigs(configDir string) error {
	configDir = libstring.ExpandTildeAndEnv(configDir)

	// Create configDir if it does not exist
	if _, err := os.Stat(configDir); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(configDir, 0755)
			if err != nil {
				return err
			}

			logrus.WithFields(logrus.Fields{
				"Directory": configDir,
			}).Infof("Created config directory")
		}
	}

	// Create a default general.toml
	generalToml := `# Addr is the host and port of ResourceD Master HTTP/S server
Addr = "localhost:55655"

# Valid LogLevel are: debug, info, warning, error, fatal, panic
LogLevel = "info"

# List of all instances hostnames
Hosts = []

# DSN to PostgreSQL
DSN = "postgres://localhost:5432/resourced-master?sslmode=disable"

RequestTimeout = "1s"

# Change this!
CookieSecret = "T0PS3CR3T"

[HTTPS]
# Path to HTTPS cert file
CertFile = ""

# Path to HTTPS key file
KeyFile = ""

[Watchers]
ListFetchInterval = "60s"

[Watchers.Email]
From = "alert@example.com"
Subject = ""
Host = "smtp.example.com"
Port = 25
Username = ""
Password = ""
`

	err := ioutil.WriteFile(path.Join(configDir, "general.toml"), []byte(generalToml), 0644)
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{
		"File": path.Join(configDir, "general.toml"),
	}).Infof("Created general config file")

	return nil
}
