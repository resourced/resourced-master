package config

import (
	"github.com/gocql/gocql"
)

// NewCassandraDBConfig connects to all the databases and returns them in CassandraDBConfig instance.
func NewCassandraDBConfig(generalConfig GeneralConfig) (*CassandraDBConfig, error) {
	conf := &CassandraDBConfig{}

	// ---------------------------------------------------------
	// core table
	//
	if len(generalConfig.Cassandra.Hosts) > 0 {
		cluster := gocql.NewCluster(generalConfig.Cassandra.Hosts...)
		cluster.Keyspace = generalConfig.Cassandra.Keyspace

		if generalConfig.Cassandra.ProtoVersion > 0 {
			cluster.ProtoVersion = generalConfig.Cassandra.ProtoVersion
		} else {
			cluster.ProtoVersion = 4
		}

		if generalConfig.Cassandra.Port > 0 {
			cluster.Port = generalConfig.Cassandra.Port
		} else {
			cluster.Port = 9042
		}

		if generalConfig.Cassandra.NumConns > 0 {
			cluster.NumConns = generalConfig.Cassandra.NumConns
		} else {
			cluster.NumConns = 2
		}

		if generalConfig.Cassandra.MaxPreparedStmts > 0 {
			cluster.MaxPreparedStmts = generalConfig.Cassandra.MaxPreparedStmts
		} else {
			cluster.MaxPreparedStmts = 1000
		}

		if generalConfig.Cassandra.MaxRoutingKeyInfo > 0 {
			cluster.MaxRoutingKeyInfo = generalConfig.Cassandra.MaxRoutingKeyInfo
		} else {
			cluster.MaxRoutingKeyInfo = 1000
		}

		if generalConfig.Cassandra.PageSize > 0 {
			cluster.PageSize = generalConfig.Cassandra.PageSize
		} else {
			cluster.PageSize = 5000
		}

		if generalConfig.Cassandra.Consistency == "one" {
			cluster.Consistency = gocql.One
		} else if generalConfig.Cassandra.Consistency == "quorum" {
			cluster.Consistency = gocql.Quorum
		} else if generalConfig.Cassandra.Consistency == "" {
			cluster.Consistency = gocql.One
		}

		session, err := cluster.CreateSession()
		if err != nil {
			return nil, err
		}

		conf.Core = cluster
		conf.CoreSession = session
	}

	// ---------------------------------------------------------
	// core table
	//
	if len(generalConfig.Cassandra.Hosts) > 0 {
		cluster := gocql.NewCluster(generalConfig.Cassandra.Hosts...)
		cluster.Keyspace = generalConfig.Cassandra.Keyspace

		if generalConfig.Cassandra.ProtoVersion > 0 {
			cluster.ProtoVersion = generalConfig.Cassandra.ProtoVersion
		} else {
			cluster.ProtoVersion = 4
		}

		if generalConfig.Cassandra.Port > 0 {
			cluster.Port = generalConfig.Cassandra.Port
		} else {
			cluster.Port = 9042
		}

		if generalConfig.Cassandra.NumConns > 0 {
			cluster.NumConns = generalConfig.Cassandra.NumConns
		} else {
			cluster.NumConns = 2
		}

		if generalConfig.Cassandra.MaxPreparedStmts > 0 {
			cluster.MaxPreparedStmts = generalConfig.Cassandra.MaxPreparedStmts
		} else {
			cluster.MaxPreparedStmts = 1000
		}

		if generalConfig.Cassandra.MaxRoutingKeyInfo > 0 {
			cluster.MaxRoutingKeyInfo = generalConfig.Cassandra.MaxRoutingKeyInfo
		} else {
			cluster.MaxRoutingKeyInfo = 1000
		}

		if generalConfig.Cassandra.PageSize > 0 {
			cluster.PageSize = generalConfig.Cassandra.PageSize
		} else {
			cluster.PageSize = 5000
		}

		if generalConfig.Cassandra.Consistency == "one" {
			cluster.Consistency = gocql.One
		} else if generalConfig.Cassandra.Consistency == "quorum" {
			cluster.Consistency = gocql.Quorum
		} else if generalConfig.Cassandra.Consistency == "" {
			cluster.Consistency = gocql.One
		}

		session, err := cluster.CreateSession()
		if err != nil {
			return nil, err
		}

		conf.Core = cluster
		conf.CoreSession = session
	}

	// ---------------------------------------------------------
	// hosts table
	//
	if len(generalConfig.Hosts.Cassandra.Hosts) > 0 {
		cluster := gocql.NewCluster(generalConfig.Hosts.Cassandra.Hosts...)
		cluster.Keyspace = generalConfig.Hosts.Cassandra.Keyspace

		if generalConfig.Hosts.Cassandra.ProtoVersion > 0 {
			cluster.ProtoVersion = generalConfig.Hosts.Cassandra.ProtoVersion
		} else {
			cluster.ProtoVersion = 4
		}

		if generalConfig.Hosts.Cassandra.Port > 0 {
			cluster.Port = generalConfig.Hosts.Cassandra.Port
		} else {
			cluster.Port = 9042
		}

		if generalConfig.Hosts.Cassandra.NumConns > 0 {
			cluster.NumConns = generalConfig.Hosts.Cassandra.NumConns
		} else {
			cluster.NumConns = 2
		}

		if generalConfig.Hosts.Cassandra.MaxPreparedStmts > 0 {
			cluster.MaxPreparedStmts = generalConfig.Hosts.Cassandra.MaxPreparedStmts
		} else {
			cluster.MaxPreparedStmts = 1000
		}

		if generalConfig.Hosts.Cassandra.MaxRoutingKeyInfo > 0 {
			cluster.MaxRoutingKeyInfo = generalConfig.Hosts.Cassandra.MaxRoutingKeyInfo
		} else {
			cluster.MaxRoutingKeyInfo = 1000
		}

		if generalConfig.Hosts.Cassandra.PageSize > 0 {
			cluster.PageSize = generalConfig.Hosts.Cassandra.PageSize
		} else {
			cluster.PageSize = 5000
		}

		if generalConfig.Hosts.Cassandra.Consistency == "one" {
			cluster.Consistency = gocql.One
		} else if generalConfig.Hosts.Cassandra.Consistency == "quorum" {
			cluster.Consistency = gocql.Quorum
		} else if generalConfig.Hosts.Cassandra.Consistency == "" {
			cluster.Consistency = gocql.One
		}

		session, err := cluster.CreateSession()
		if err != nil {
			return nil, err
		}

		conf.Host = cluster
		conf.HostSession = session
	}

	// ---------------------------------------------------------
	// ts_metrics table
	//
	if len(generalConfig.Metrics.Cassandra.Hosts) > 0 {
		cluster := gocql.NewCluster(generalConfig.Metrics.Cassandra.Hosts...)
		cluster.Keyspace = generalConfig.Metrics.Cassandra.Keyspace

		if generalConfig.Metrics.Cassandra.ProtoVersion > 0 {
			cluster.ProtoVersion = generalConfig.Metrics.Cassandra.ProtoVersion
		} else {
			cluster.ProtoVersion = 4
		}

		if generalConfig.Metrics.Cassandra.Port > 0 {
			cluster.Port = generalConfig.Metrics.Cassandra.Port
		} else {
			cluster.Port = 9042
		}

		if generalConfig.Metrics.Cassandra.NumConns > 0 {
			cluster.NumConns = generalConfig.Metrics.Cassandra.NumConns
		} else {
			cluster.NumConns = 2
		}

		if generalConfig.Metrics.Cassandra.MaxPreparedStmts > 0 {
			cluster.MaxPreparedStmts = generalConfig.Metrics.Cassandra.MaxPreparedStmts
		} else {
			cluster.MaxPreparedStmts = 1000
		}

		if generalConfig.Metrics.Cassandra.MaxRoutingKeyInfo > 0 {
			cluster.MaxRoutingKeyInfo = generalConfig.Metrics.Cassandra.MaxRoutingKeyInfo
		} else {
			cluster.MaxRoutingKeyInfo = 1000
		}

		if generalConfig.Metrics.Cassandra.PageSize > 0 {
			cluster.PageSize = generalConfig.Metrics.Cassandra.PageSize
		} else {
			cluster.PageSize = 5000
		}

		if generalConfig.Metrics.Cassandra.Consistency == "one" {
			cluster.Consistency = gocql.One
		} else if generalConfig.Metrics.Cassandra.Consistency == "quorum" {
			cluster.Consistency = gocql.Quorum
		} else if generalConfig.Metrics.Cassandra.Consistency == "" {
			cluster.Consistency = gocql.One
		}

		session, err := cluster.CreateSession()
		if err != nil {
			return nil, err
		}

		conf.TSMetric = cluster
		conf.TSMetricSession = session
	}

	// ---------------------------------------------------------
	// ts_logs table
	//
	if len(generalConfig.Logs.Cassandra.Hosts) > 0 {
		cluster := gocql.NewCluster(generalConfig.Logs.Cassandra.Hosts...)
		cluster.Keyspace = generalConfig.Logs.Cassandra.Keyspace

		if generalConfig.Logs.Cassandra.ProtoVersion > 0 {
			cluster.ProtoVersion = generalConfig.Logs.Cassandra.ProtoVersion
		} else {
			cluster.ProtoVersion = 4
		}

		if generalConfig.Logs.Cassandra.Port > 0 {
			cluster.Port = generalConfig.Logs.Cassandra.Port
		} else {
			cluster.Port = 9042
		}

		if generalConfig.Logs.Cassandra.NumConns > 0 {
			cluster.NumConns = generalConfig.Logs.Cassandra.NumConns
		} else {
			cluster.NumConns = 2
		}

		if generalConfig.Logs.Cassandra.MaxPreparedStmts > 0 {
			cluster.MaxPreparedStmts = generalConfig.Logs.Cassandra.MaxPreparedStmts
		} else {
			cluster.MaxPreparedStmts = 1000
		}

		if generalConfig.Logs.Cassandra.MaxRoutingKeyInfo > 0 {
			cluster.MaxRoutingKeyInfo = generalConfig.Logs.Cassandra.MaxRoutingKeyInfo
		} else {
			cluster.MaxRoutingKeyInfo = 1000
		}

		if generalConfig.Logs.Cassandra.PageSize > 0 {
			cluster.PageSize = generalConfig.Logs.Cassandra.PageSize
		} else {
			cluster.PageSize = 5000
		}

		if generalConfig.Logs.Cassandra.Consistency == "one" {
			cluster.Consistency = gocql.One
		} else if generalConfig.Logs.Cassandra.Consistency == "quorum" {
			cluster.Consistency = gocql.Quorum
		} else if generalConfig.Logs.Cassandra.Consistency == "" {
			cluster.Consistency = gocql.One
		}

		session, err := cluster.CreateSession()
		if err != nil {
			return nil, err
		}

		conf.TSLog = cluster
		conf.TSLogSession = session
	}

	// ---------------------------------------------------------
	// ts_events table
	//
	if len(generalConfig.Events.Cassandra.Hosts) > 0 {
		cluster := gocql.NewCluster(generalConfig.Events.Cassandra.Hosts...)
		cluster.Keyspace = generalConfig.Events.Cassandra.Keyspace

		if generalConfig.Events.Cassandra.ProtoVersion > 0 {
			cluster.ProtoVersion = generalConfig.Events.Cassandra.ProtoVersion
		} else {
			cluster.ProtoVersion = 4
		}

		if generalConfig.Events.Cassandra.Port > 0 {
			cluster.Port = generalConfig.Events.Cassandra.Port
		} else {
			cluster.Port = 9042
		}

		if generalConfig.Events.Cassandra.NumConns > 0 {
			cluster.NumConns = generalConfig.Events.Cassandra.NumConns
		} else {
			cluster.NumConns = 2
		}

		if generalConfig.Events.Cassandra.MaxPreparedStmts > 0 {
			cluster.MaxPreparedStmts = generalConfig.Events.Cassandra.MaxPreparedStmts
		} else {
			cluster.MaxPreparedStmts = 1000
		}

		if generalConfig.Events.Cassandra.MaxRoutingKeyInfo > 0 {
			cluster.MaxRoutingKeyInfo = generalConfig.Events.Cassandra.MaxRoutingKeyInfo
		} else {
			cluster.MaxRoutingKeyInfo = 1000
		}

		if generalConfig.Events.Cassandra.PageSize > 0 {
			cluster.PageSize = generalConfig.Events.Cassandra.PageSize
		} else {
			cluster.PageSize = 5000
		}

		if generalConfig.Events.Cassandra.Consistency == "one" {
			cluster.Consistency = gocql.One
		} else if generalConfig.Events.Cassandra.Consistency == "quorum" {
			cluster.Consistency = gocql.Quorum
		} else if generalConfig.Events.Cassandra.Consistency == "" {
			cluster.Consistency = gocql.One
		}

		session, err := cluster.CreateSession()
		if err != nil {
			return nil, err
		}

		conf.TSEvent = cluster
		conf.TSEventSession = session
	}
	return conf, nil
}

// CassandraDBConfig stores all database configuration data.
type CassandraDBConfig struct {
	Core        *gocql.ClusterConfig
	CoreSession *gocql.Session

	Host        *gocql.ClusterConfig
	HostSession *gocql.Session

	TSMetric        *gocql.ClusterConfig
	TSMetricSession *gocql.Session

	TSLog        *gocql.ClusterConfig
	TSLogSession *gocql.Session

	TSEvent        *gocql.ClusterConfig
	TSEventSession *gocql.Session
}
