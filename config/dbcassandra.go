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
		cluster.ProtoVersion = generalConfig.Metrics.Cassandra.ProtoVersion
		cluster.Port = generalConfig.Metrics.Cassandra.Port
		cluster.Keyspace = generalConfig.Metrics.Cassandra.Keyspace
		cluster.NumConns = generalConfig.Metrics.Cassandra.NumConns
		cluster.MaxPreparedStmts = generalConfig.Metrics.Cassandra.MaxPreparedStmts
		cluster.MaxRoutingKeyInfo = generalConfig.Metrics.Cassandra.MaxRoutingKeyInfo
		cluster.PageSize = generalConfig.Metrics.Cassandra.PageSize

		if generalConfig.Metrics.Cassandra.Consistency == "one" {
			cluster.Consistency = gocql.One
		} else if generalConfig.Metrics.Cassandra.Consistency == "quorum" {
			cluster.Consistency = gocql.Quorum
		}

		session, err := cluster.CreateSession()
		if err != nil {
			return nil, err
		}

		conf.TSMetric = cluster
		conf.TSMetricSession = session
	}

	// ---------------------------------------------------------
	// ts_metrics_aggr_15m table
	//
	if len(generalConfig.MetricsAggr15m.Cassandra.Hosts) > 0 {
		cluster := gocql.NewCluster(generalConfig.MetricsAggr15m.Cassandra.Hosts...)
		cluster.ProtoVersion = generalConfig.MetricsAggr15m.Cassandra.ProtoVersion
		cluster.Port = generalConfig.MetricsAggr15m.Cassandra.Port
		cluster.Keyspace = generalConfig.MetricsAggr15m.Cassandra.Keyspace
		cluster.NumConns = generalConfig.MetricsAggr15m.Cassandra.NumConns
		cluster.MaxPreparedStmts = generalConfig.MetricsAggr15m.Cassandra.MaxPreparedStmts
		cluster.MaxRoutingKeyInfo = generalConfig.MetricsAggr15m.Cassandra.MaxRoutingKeyInfo
		cluster.PageSize = generalConfig.MetricsAggr15m.Cassandra.PageSize

		if generalConfig.MetricsAggr15m.Cassandra.Consistency == "one" {
			cluster.Consistency = gocql.One
		} else if generalConfig.MetricsAggr15m.Cassandra.Consistency == "quorum" {
			cluster.Consistency = gocql.Quorum
		}

		session, err := cluster.CreateSession()
		if err != nil {
			return nil, err
		}

		conf.TSMetricAggr15m = cluster
		conf.TSMetricAggr15mSession = session
	}

	return conf, nil
}

// CassandraDBConfig stores all database configuration data.
type CassandraDBConfig struct {
	TSMetric               *gocql.ClusterConfig
	TSMetricAggr15m        *gocql.ClusterConfig
	TSMetricSession        *gocql.Session
	TSMetricAggr15mSession *gocql.Session
}
