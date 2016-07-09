// Package config provides data structures for Application configurations.
package config

import (
	"io/ioutil"
	"path"
	"strconv"
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
	Addr                    string
	LogLevel                string
	DSN                     string
	CookieSecret            string
	RequestShutdownTimeout  string
	VIPAddr                 string
	VIPProtocol             string
	EnablePeriodicPruneJobs bool

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

	Hosts struct {
		DSN            string
		DSNByClusterID map[string]string
	}

	Metrics struct {
		DSN            string
		DSNByClusterID map[string]string

		DataRetention int
	}

	MetricsAggr15m struct {
		DSN            string
		DSNByClusterID map[string]string

		DataRetention int
	}

	Events struct {
		DSN            string
		DSNByClusterID map[string]string

		DataRetention int
	}

	ExecutorLogs struct {
		DSN            string
		DSNByClusterID map[string]string

		DataRetention int
	}

	Logs struct {
		DSN            string
		DSNByClusterID map[string]string

		DataRetention int
	}

	Checks struct {
		Email *EmailConfig

		SMSEmailGateway map[string]string

		DSN            string
		DSNByClusterID map[string]string

		DataRetention int
	}

	Email *EmailConfig
}

// NewDBConfig connects to all the databases and returns them in DBConfig instance.
func NewDBConfig(generalConfig GeneralConfig) (*DBConfig, error) {
	conf := &DBConfig{}
	conf.HostByClusterID = make(map[int64]*sqlx.DB)
	conf.TSMetricByClusterID = make(map[int64]*sqlx.DB)
	conf.TSMetricAggr15mByClusterID = make(map[int64]*sqlx.DB)
	conf.TSEventByClusterID = make(map[int64]*sqlx.DB)
	conf.TSExecutorLogByClusterID = make(map[int64]*sqlx.DB)
	conf.TSLogByClusterID = make(map[int64]*sqlx.DB)
	conf.TSCheckByClusterID = make(map[int64]*sqlx.DB)

	db, err := sqlx.Connect("postgres", generalConfig.DSN)
	if err != nil {
		return nil, err
	}
	conf.Core = db

	// ---------------------------------------------------------
	// hosts table
	//
	db, err = sqlx.Connect("postgres", generalConfig.Hosts.DSN)
	if err != nil {
		return nil, err
	}
	conf.Host = db

	for clusterIDString, dsn := range generalConfig.Hosts.DSNByClusterID {
		clusterID, err := strconv.ParseInt(clusterIDString, 10, 64)
		if err != nil {
			return nil, err
		}

		db, err = sqlx.Connect("postgres", dsn)
		if err != nil {
			return nil, err
		}
		conf.HostByClusterID[clusterID] = db
	}

	// ---------------------------------------------------------
	// ts_metrics table
	//
	db, err = sqlx.Connect("postgres", generalConfig.Metrics.DSN)
	if err != nil {
		return nil, err
	}
	conf.TSMetric = db

	for clusterIDString, dsn := range generalConfig.Metrics.DSNByClusterID {
		clusterID, err := strconv.ParseInt(clusterIDString, 10, 64)
		if err != nil {
			return nil, err
		}

		db, err = sqlx.Connect("postgres", dsn)
		if err != nil {
			return nil, err
		}
		conf.TSMetricByClusterID[clusterID] = db
	}

	// ---------------------------------------------------------
	// ts_metrics_aggr_15m table
	//
	db, err = sqlx.Connect("postgres", generalConfig.MetricsAggr15m.DSN)
	if err != nil {
		return nil, err
	}
	conf.TSMetricAggr15m = db

	for clusterIDString, dsn := range generalConfig.MetricsAggr15m.DSNByClusterID {
		clusterID, err := strconv.ParseInt(clusterIDString, 10, 64)
		if err != nil {
			return nil, err
		}

		db, err = sqlx.Connect("postgres", dsn)
		if err != nil {
			return nil, err
		}
		conf.TSMetricAggr15mByClusterID[clusterID] = db
	}

	// ---------------------------------------------------------
	// ts_events table
	//
	db, err = sqlx.Connect("postgres", generalConfig.Events.DSN)
	if err != nil {
		return nil, err
	}
	conf.TSEvent = db

	for clusterIDString, dsn := range generalConfig.Events.DSNByClusterID {
		clusterID, err := strconv.ParseInt(clusterIDString, 10, 64)
		if err != nil {
			return nil, err
		}

		db, err = sqlx.Connect("postgres", dsn)
		if err != nil {
			return nil, err
		}
		conf.TSEventByClusterID[clusterID] = db
	}

	// ---------------------------------------------------------
	// ts_executor_logs table
	//
	db, err = sqlx.Connect("postgres", generalConfig.ExecutorLogs.DSN)
	if err != nil {
		return nil, err
	}
	conf.TSExecutorLog = db

	for clusterIDString, dsn := range generalConfig.ExecutorLogs.DSNByClusterID {
		clusterID, err := strconv.ParseInt(clusterIDString, 10, 64)
		if err != nil {
			return nil, err
		}

		db, err = sqlx.Connect("postgres", dsn)
		if err != nil {
			return nil, err
		}
		conf.TSExecutorLogByClusterID[clusterID] = db
	}

	// ---------------------------------------------------------
	// ts_logs table
	//
	db, err = sqlx.Connect("postgres", generalConfig.Logs.DSN)
	if err != nil {
		return nil, err
	}
	conf.TSLog = db

	for clusterIDString, dsn := range generalConfig.Logs.DSNByClusterID {
		clusterID, err := strconv.ParseInt(clusterIDString, 10, 64)
		if err != nil {
			return nil, err
		}

		db, err = sqlx.Connect("postgres", dsn)
		if err != nil {
			return nil, err
		}
		conf.TSLogByClusterID[clusterID] = db
	}

	// ---------------------------------------------------------
	// ts_checks table
	//
	db, err = sqlx.Connect("postgres", generalConfig.Checks.DSN)
	if err != nil {
		return nil, err
	}
	conf.TSCheck = db

	for clusterIDString, dsn := range generalConfig.Checks.DSNByClusterID {
		clusterID, err := strconv.ParseInt(clusterIDString, 10, 64)
		if err != nil {
			return nil, err
		}

		db, err = sqlx.Connect("postgres", dsn)
		if err != nil {
			return nil, err
		}
		conf.TSCheckByClusterID[clusterID] = db
	}

	return conf, nil
}

// DBConfig stores all database configuration data.
type DBConfig struct {
	Core                       *sqlx.DB
	Host                       *sqlx.DB
	HostByClusterID            map[int64]*sqlx.DB
	TSMetric                   *sqlx.DB
	TSMetricByClusterID        map[int64]*sqlx.DB
	TSMetricAggr15m            *sqlx.DB
	TSMetricAggr15mByClusterID map[int64]*sqlx.DB
	TSEvent                    *sqlx.DB
	TSEventByClusterID         map[int64]*sqlx.DB
	TSExecutorLog              *sqlx.DB
	TSExecutorLogByClusterID   map[int64]*sqlx.DB
	TSLog                      *sqlx.DB
	TSLogByClusterID           map[int64]*sqlx.DB
	TSCheck                    *sqlx.DB
	TSCheckByClusterID         map[int64]*sqlx.DB
}

func (dbconf *DBConfig) GetHost(clusterID int64) *sqlx.DB {
	conn, ok := dbconf.HostByClusterID[clusterID]
	if !ok {
		conn = dbconf.Host
	}

	return conn
}

func (dbconf *DBConfig) GetTSMetric(clusterID int64) *sqlx.DB {
	conn, ok := dbconf.TSMetricByClusterID[clusterID]
	if !ok {
		conn = dbconf.TSMetric
	}

	return conn
}

func (dbconf *DBConfig) GetTSMetricAggr15m(clusterID int64) *sqlx.DB {
	conn, ok := dbconf.TSMetricAggr15mByClusterID[clusterID]
	if !ok {
		conn = dbconf.TSMetricAggr15m
	}

	return conn
}

func (dbconf *DBConfig) GetTSEvent(clusterID int64) *sqlx.DB {
	conn, ok := dbconf.TSEventByClusterID[clusterID]
	if !ok {
		conn = dbconf.TSEvent
	}

	return conn
}

func (dbconf *DBConfig) GetTSExecutorLog(clusterID int64) *sqlx.DB {
	conn, ok := dbconf.TSExecutorLogByClusterID[clusterID]
	if !ok {
		conn = dbconf.TSExecutorLog
	}

	return conn
}

func (dbconf *DBConfig) GetTSLog(clusterID int64) *sqlx.DB {
	conn, ok := dbconf.TSLogByClusterID[clusterID]
	if !ok {
		conn = dbconf.TSLog
	}

	return conn
}

func (dbconf *DBConfig) GetTSCheck(clusterID int64) *sqlx.DB {
	conn, ok := dbconf.TSCheckByClusterID[clusterID]
	if !ok {
		conn = dbconf.TSCheck
	}

	return conn
}
