package config

import (
	"github.com/gocql/gocql"
)

// NewCassandraDBConfig connects to all the databases and returns them in CassandraDBConfig instance.
func NewCassandraDBConfig(generalConfig GeneralConfig) (*CassandraDBConfig, error) {
	conf := &CassandraDBConfig{}

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

	return conf, nil
}

// CassandraDBConfig stores all database configuration data.
type CassandraDBConfig struct {
	TSMetric        *gocql.ClusterConfig
	TSMetricSession *gocql.Session
}
