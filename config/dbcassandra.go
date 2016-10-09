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

	return conf, nil
}

// CassandraDBConfig stores all database configuration data.
type CassandraDBConfig struct {
	TSMetric        *gocql.ClusterConfig
	TSMetricSession *gocql.Session
}
