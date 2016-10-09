package shims

import (
	"fmt"
	"time"

	"github.com/resourced/resourced-master/models/cassandra"
	"github.com/resourced/resourced-master/models/pg"
	"github.com/resourced/resourced-master/models/shared"
)

func NewTSMetric(params Parameters) *TSMetric {
	ts := &TSMetric{}
	ts.Parameters = params
	return ts
}

type TSMetricHighchartPayload struct {
	Name string          `json:"name"`
	Data [][]interface{} `json:"data"`
}

type TSMetricRow struct {
	ClusterID int64   `db:"cluster_id"`
	MetricID  int64   `db:"metric_id"`
	Created   int64   `db:"created"`
	Key       string  `db:"key"`
	Host      string  `db:"host"`
	Value     float64 `db:"value"`
}

type TSMetric struct {
	Parameters Parameters
}

func (ts *TSMetric) CreateByHostRow(hostRow shared.IHostRow, metricsMap map[string]int64, deletedFrom int64, ttl time.Duration) error {
	if ts.Parameters.DBType == "pg" {
		return pg.NewTSMetric(ts.Parameters.PGDB).CreateByHostRow(ts.Parameters.PGTx, hostRow, metricsMap, deletedFrom)

	} else if ts.Parameters.DBType == "cassandra" {
		return cassandra.NewTSMetric(ts.Parameters.CassandraSession).CreateByHostRow(hostRow, metricsMap, ttl)
	}

	return fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

// func (ts *TSMetric) AllByMetricIDHostAndRange(clusterID, metricID int64, host string, from, to, deletedFrom int64) ([]interface{}, error) {
// 	if ts.Parameters.DBType == "pg" {
// 		return pg.NewTSMetric(ts.Parameters.PGDB).AllByMetricIDHostAndRange(ts.Parameters.PGTx, clusterID, metricID, host, from, to, deletedFrom)

// 	} else if ts.Parameters.DBType == "cassandra" {
// 		return cassandra.NewTSMetric(ts.Parameters.CassandraSession).AllByMetricIDHostAndRange(clusterID, metricID, host, from, to)
// 	}

// 	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
// }

func (ts *TSMetric) AllByMetricIDHostAndRangeForHighchart(clusterID, metricID int64, host string, from, to, deletedFrom int64) (interface{}, error) {
	if ts.Parameters.DBType == "pg" {
		return pg.NewTSMetric(ts.Parameters.PGDB).AllByMetricIDHostAndRangeForHighchart(ts.Parameters.PGTx, clusterID, metricID, host, from, to, deletedFrom)

	} else if ts.Parameters.DBType == "cassandra" {
		return cassandra.NewTSMetric(ts.Parameters.CassandraSession).AllByMetricIDHostAndRangeForHighchart(clusterID, metricID, host, from, to)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

// func (ts *TSMetric) AllByMetricIDAndRange(clusterID, metricID, from, to, deletedFrom int64) ([]interface{}, error) {
// 	if ts.Parameters.DBType == "pg" {
// 		return pg.NewTSMetric(ts.Parameters.PGDB).AllByMetricIDAndRange(ts.Parameters.PGTx, clusterID, metricID, from, to, deletedFrom)

// 	} else if ts.Parameters.DBType == "cassandra" {
// 		return cassandra.NewTSMetric(ts.Parameters.CassandraSession).AllByMetricIDAndRange(clusterID, metricID, from, to)
// 	}

// 	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
// }

func (ts *TSMetric) AllByMetricIDAndRangeForHighchart(clusterID, metricID, from, to, deletedFrom int64) (interface{}, error) {
	if ts.Parameters.DBType == "pg" {
		return pg.NewTSMetric(ts.Parameters.PGDB).AllByMetricIDAndRangeForHighchart(ts.Parameters.PGTx, clusterID, metricID, from, to, deletedFrom)

	} else if ts.Parameters.DBType == "cassandra" {
		return cassandra.NewTSMetric(ts.Parameters.CassandraSession).AllByMetricIDAndRangeForHighchart(clusterID, metricID, from, to)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}
