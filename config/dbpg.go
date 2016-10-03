package config

import (
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

// NewPGDBConfig connects to all the databases and returns them in PGDBConfig instance.
func NewPGDBConfig(generalConfig GeneralConfig) (*PGDBConfig, error) {
	conf := &PGDBConfig{}
	conf.HostByClusterID = make(map[int64]*sqlx.DB)
	conf.TSMetricByClusterID = make(map[int64]*sqlx.DB)
	conf.TSMetricAggr15mByClusterID = make(map[int64]*sqlx.DB)
	conf.TSEventByClusterID = make(map[int64]*sqlx.DB)
	conf.TSLogByClusterID = make(map[int64]*sqlx.DB)
	conf.TSCheckByClusterID = make(map[int64]*sqlx.DB)

	if strings.HasPrefix(generalConfig.PostgreSQL.DSN, "postgres") {
		db, err := sqlx.Connect("postgres", generalConfig.PostgreSQL.DSN)
		if err != nil {
			return nil, err
		}
		if generalConfig.PostgreSQL.MaxOpenConnections > int64(0) {
			db.DB.SetMaxOpenConns(int(generalConfig.PostgreSQL.MaxOpenConnections))
		}
		conf.Core = db
	}

	// ---------------------------------------------------------
	// hosts table
	//
	if strings.HasPrefix(generalConfig.Hosts.PostgreSQL.DSN, "postgres") {
		db, err := sqlx.Connect("postgres", generalConfig.Hosts.PostgreSQL.DSN)
		if err != nil {
			return nil, err
		}
		if generalConfig.Hosts.PostgreSQL.MaxOpenConnections > int64(0) {
			db.DB.SetMaxOpenConns(int(generalConfig.Hosts.PostgreSQL.MaxOpenConnections))
		}
		conf.Host = db
	}

	for clusterIDString, dsn := range generalConfig.Hosts.PostgreSQL.DSNByClusterID {
		if strings.HasPrefix(dsn, "postgres") {
			clusterID, err := strconv.ParseInt(clusterIDString, 10, 64)
			if err != nil {
				return nil, err
			}

			db, err := sqlx.Connect("postgres", dsn)
			if err != nil {
				return nil, err
			}
			if generalConfig.Hosts.PostgreSQL.MaxOpenConnections > int64(0) {
				db.DB.SetMaxOpenConns(int(generalConfig.Hosts.PostgreSQL.MaxOpenConnections))
			}
			conf.HostByClusterID[clusterID] = db
		}
	}

	// ---------------------------------------------------------
	// ts_metrics table
	//
	if strings.HasPrefix(generalConfig.Metrics.PostgreSQL.DSN, "postgres") {
		db, err := sqlx.Connect("postgres", generalConfig.Metrics.PostgreSQL.DSN)
		if err != nil {
			return nil, err
		}
		if generalConfig.Metrics.PostgreSQL.MaxOpenConnections > int64(0) {
			db.DB.SetMaxOpenConns(int(generalConfig.Metrics.PostgreSQL.MaxOpenConnections))
		}
		conf.TSMetric = db
	}

	for clusterIDString, dsn := range generalConfig.Metrics.PostgreSQL.DSNByClusterID {
		if strings.HasPrefix(dsn, "postgres") {
			clusterID, err := strconv.ParseInt(clusterIDString, 10, 64)
			if err != nil {
				return nil, err
			}

			db, err := sqlx.Connect("postgres", dsn)
			if err != nil {
				return nil, err
			}
			if generalConfig.Metrics.PostgreSQL.MaxOpenConnections > int64(0) {
				db.DB.SetMaxOpenConns(int(generalConfig.Metrics.PostgreSQL.MaxOpenConnections))
			}
			conf.TSMetricByClusterID[clusterID] = db
		}
	}

	// ---------------------------------------------------------
	// ts_metrics_aggr_15m table
	//
	if strings.HasPrefix(generalConfig.MetricsAggr15m.PostgreSQL.DSN, "postgres") {
		db, err := sqlx.Connect("postgres", generalConfig.MetricsAggr15m.PostgreSQL.DSN)
		if err != nil {
			return nil, err
		}
		if generalConfig.MetricsAggr15m.PostgreSQL.MaxOpenConnections > int64(0) {
			db.DB.SetMaxOpenConns(int(generalConfig.MetricsAggr15m.PostgreSQL.MaxOpenConnections))
		}
		conf.TSMetricAggr15m = db
	}

	for clusterIDString, dsn := range generalConfig.MetricsAggr15m.PostgreSQL.DSNByClusterID {
		if strings.HasPrefix(dsn, "postgres") {
			clusterID, err := strconv.ParseInt(clusterIDString, 10, 64)
			if err != nil {
				return nil, err
			}

			db, err := sqlx.Connect("postgres", dsn)
			if err != nil {
				return nil, err
			}
			if generalConfig.MetricsAggr15m.PostgreSQL.MaxOpenConnections > int64(0) {
				db.DB.SetMaxOpenConns(int(generalConfig.MetricsAggr15m.PostgreSQL.MaxOpenConnections))
			}
			conf.TSMetricAggr15mByClusterID[clusterID] = db
		}
	}

	// ---------------------------------------------------------
	// ts_events table
	//
	if strings.HasPrefix(generalConfig.Events.PostgreSQL.DSN, "postgres") {
		db, err := sqlx.Connect("postgres", generalConfig.Events.PostgreSQL.DSN)
		if err != nil {
			return nil, err
		}
		if generalConfig.Events.PostgreSQL.MaxOpenConnections > int64(0) {
			db.DB.SetMaxOpenConns(int(generalConfig.Events.PostgreSQL.MaxOpenConnections))
		}
		conf.TSEvent = db
	}

	for clusterIDString, dsn := range generalConfig.Events.PostgreSQL.DSNByClusterID {
		if strings.HasPrefix(dsn, "postgres") {
			clusterID, err := strconv.ParseInt(clusterIDString, 10, 64)
			if err != nil {
				return nil, err
			}

			db, err := sqlx.Connect("postgres", dsn)
			if err != nil {
				return nil, err
			}
			if generalConfig.Events.PostgreSQL.MaxOpenConnections > int64(0) {
				db.DB.SetMaxOpenConns(int(generalConfig.Events.PostgreSQL.MaxOpenConnections))
			}
			conf.TSEventByClusterID[clusterID] = db
		}
	}

	// ---------------------------------------------------------
	// ts_logs table
	//
	if strings.HasPrefix(generalConfig.Logs.PostgreSQL.DSN, "postgres") {
		db, err := sqlx.Connect("postgres", generalConfig.Logs.PostgreSQL.DSN)
		if err != nil {
			return nil, err
		}
		if generalConfig.Logs.PostgreSQL.MaxOpenConnections > int64(0) {
			db.DB.SetMaxOpenConns(int(generalConfig.Logs.PostgreSQL.MaxOpenConnections))
		}
		conf.TSLog = db
	}

	for clusterIDString, dsn := range generalConfig.Logs.PostgreSQL.DSNByClusterID {
		if strings.HasPrefix(dsn, "postgres") {
			clusterID, err := strconv.ParseInt(clusterIDString, 10, 64)
			if err != nil {
				return nil, err
			}

			db, err := sqlx.Connect("postgres", dsn)
			if err != nil {
				return nil, err
			}
			if generalConfig.Logs.PostgreSQL.MaxOpenConnections > int64(0) {
				db.DB.SetMaxOpenConns(int(generalConfig.Logs.PostgreSQL.MaxOpenConnections))
			}
			conf.TSLogByClusterID[clusterID] = db
		}
	}

	// ---------------------------------------------------------
	// ts_checks table
	//
	if strings.HasPrefix(generalConfig.Checks.PostgreSQL.DSN, "postgres") {
		db, err := sqlx.Connect("postgres", generalConfig.Checks.PostgreSQL.DSN)
		if err != nil {
			return nil, err
		}
		if generalConfig.Checks.PostgreSQL.MaxOpenConnections > int64(0) {
			db.DB.SetMaxOpenConns(int(generalConfig.Checks.PostgreSQL.MaxOpenConnections))
		}
		conf.TSCheck = db
	}

	for clusterIDString, dsn := range generalConfig.Checks.PostgreSQL.DSNByClusterID {
		if strings.HasPrefix(dsn, "postgres") {
			clusterID, err := strconv.ParseInt(clusterIDString, 10, 64)
			if err != nil {
				return nil, err
			}

			db, err := sqlx.Connect("postgres", dsn)
			if err != nil {
				return nil, err
			}
			if generalConfig.Checks.PostgreSQL.MaxOpenConnections > int64(0) {
				db.DB.SetMaxOpenConns(int(generalConfig.Checks.PostgreSQL.MaxOpenConnections))
			}
			conf.TSCheckByClusterID[clusterID] = db
		}
	}

	return conf, nil
}

// PGDBConfig stores all database configuration data.
type PGDBConfig struct {
	Core                       *sqlx.DB
	Host                       *sqlx.DB
	HostByClusterID            map[int64]*sqlx.DB
	TSMetric                   *sqlx.DB
	TSMetricByClusterID        map[int64]*sqlx.DB
	TSMetricAggr15m            *sqlx.DB
	TSMetricAggr15mByClusterID map[int64]*sqlx.DB
	TSEvent                    *sqlx.DB
	TSEventByClusterID         map[int64]*sqlx.DB
	TSLog                      *sqlx.DB
	TSLogByClusterID           map[int64]*sqlx.DB
	TSCheck                    *sqlx.DB
	TSCheckByClusterID         map[int64]*sqlx.DB
}

func (dbconf *PGDBConfig) GetHost(clusterID int64) *sqlx.DB {
	conn, ok := dbconf.HostByClusterID[clusterID]
	if !ok {
		conn = dbconf.Host
	}

	return conn
}

func (dbconf *PGDBConfig) GetTSMetric(clusterID int64) *sqlx.DB {
	conn, ok := dbconf.TSMetricByClusterID[clusterID]
	if !ok {
		conn = dbconf.TSMetric
	}

	return conn
}

func (dbconf *PGDBConfig) GetTSMetricAggr15m(clusterID int64) *sqlx.DB {
	conn, ok := dbconf.TSMetricAggr15mByClusterID[clusterID]
	if !ok {
		conn = dbconf.TSMetricAggr15m
	}

	return conn
}

func (dbconf *PGDBConfig) GetTSEvent(clusterID int64) *sqlx.DB {
	conn, ok := dbconf.TSEventByClusterID[clusterID]
	if !ok {
		conn = dbconf.TSEvent
	}

	return conn
}

func (dbconf *PGDBConfig) GetTSLog(clusterID int64) *sqlx.DB {
	conn, ok := dbconf.TSLogByClusterID[clusterID]
	if !ok {
		conn = dbconf.TSLog
	}

	return conn
}

func (dbconf *PGDBConfig) GetTSCheck(clusterID int64) *sqlx.DB {
	conn, ok := dbconf.TSCheckByClusterID[clusterID]
	if !ok {
		conn = dbconf.TSCheck
	}

	return conn
}
