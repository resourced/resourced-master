package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

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

# DSN to core PostgreSQL database
DSN = "postgres://localhost:5432/resourced-master?sslmode=disable"

RequestTimeout = "1s"

# Change this!
CookieSecret = "T0PS3CR3T"

# In production, change this to your VIP host and port
VIPAddr = "localhost:55655"

VIPProtocol = "http"

[HTTPS]
# Path to HTTPS cert file
CertFile = ""

# Path to HTTPS key file
KeyFile = ""

[Email]
From = "dontreply@example.com"
SubjectPrefix = "[ResourceDMaster]"
Host = "smtp.gmail.com"
Port = 587
Username = ""
Password = ""
Identity = ""

[Metrics]
DSNs = [
    "postgres://localhost:5432/resourced-master-ts-metrics-1?sslmode=disable",
    "postgres://localhost:5432/resourced-master-ts-metrics-2?sslmode=disable"
]

# Valid values for ReplicationPercentage are: 0 < x <= 100
# ReplicationPercentage == 100 means that every time series data is replicated to 100% of databases defined in DSNs.
# ReplicationPercentage == 50 means that every time series data is replicated to 50% of databases defined in DSNs.
# ReplicationPercentage == 0 is invalid, it will be converted to 100 instead.
ReplicationPercentage = 100

# DataRetention defines how long time series data are kept.
# The unit is defined in days.
DataRetention = 7

[Watchers]
ListFetchInterval = "60s"
DSNs = [
    "postgres://localhost:5432/resourced-master-ts-watchers-1?sslmode=disable",
    "postgres://localhost:5432/resourced-master-ts-watchers-2?sslmode=disable"
]

# DataRetention defines how long time series data are kept.
# The unit is defined in days.
DataRetention = 1

[Watchers.Email]
From = "alert@example.com"
SubjectPrefix = "[ERROR]"
Host = "smtp.gmail.com"
Port = 587
Username = ""
Password = ""
Identity = ""

[Watchers.SMSEmailGateway]
att = "txt.att.net"
alltel = "message.alltel.com"
sprint = "messaging.sprintpcs.com"
tmobile = "tmomail.com"
verizon = "vtext.com"
virgin = "vmobl.com"

[Events]
DSNs = [
    "postgres://localhost:5432/resourced-master-ts-events-1?sslmode=disable",
    "postgres://localhost:5432/resourced-master-ts-events-2?sslmode=disable"
]

# DataRetention defines how long time series data are kept.
# The unit is defined in days.
DataRetention = 7
`

	// Generate totally random cookie secret
	randomCookieSecret, err := libstring.GeneratePassword(32)
	if err != nil {
		return err
	}
	generalToml = strings.Replace(generalToml, `CookieSecret = "T0PS3CR3T"`, fmt.Sprintf(`CookieSecret = "%v"`, randomCookieSecret), 1)

	err = ioutil.WriteFile(path.Join(configDir, "general.toml"), []byte(generalToml), 0644)
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{
		"File": path.Join(configDir, "general.toml"),
	}).Infof("Created general config file")

	return nil
}
