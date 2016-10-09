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

func (ts *TSMetric) AllByMetricIDHostAndRangeForHighchart(clusterID, metricID int64, host string, from, to, deletedFrom int64) (*shared.TSMetricHighchartPayload, error) {
	if ts.Parameters.DBType == "pg" {
		return pg.NewTSMetric(ts.Parameters.PGDB).AllByMetricIDHostAndRangeForHighchart(ts.Parameters.PGTx, clusterID, metricID, host, from, to, deletedFrom)

	} else if ts.Parameters.DBType == "cassandra" {
		return cassandra.NewTSMetric(ts.Parameters.CassandraSession).AllByMetricIDHostAndRangeForHighchart(clusterID, metricID, host, from, to)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

func (ts *TSMetric) AllByMetricIDAndRangeForHighchart(clusterID, metricID, from, to, deletedFrom int64) ([]*shared.TSMetricHighchartPayload, error) {
	if ts.Parameters.DBType == "pg" {
		return pg.NewTSMetric(ts.Parameters.PGDB).AllByMetricIDAndRangeForHighchart(ts.Parameters.PGTx, clusterID, metricID, from, to, deletedFrom)

	} else if ts.Parameters.DBType == "cassandra" {
		return cassandra.NewTSMetric(ts.Parameters.CassandraSession).AllByMetricIDAndRangeForHighchart(clusterID, metricID, from, to)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

func (ts *TSMetric) GetAggregateXMinutesByMetricIDAndHostname(clusterID, metricID int64, minutes int, hostname string) (*shared.TSMetricAggregateRow, error) {
	if ts.Parameters.DBType == "pg" {
		return pg.NewTSMetric(ts.Parameters.PGDB).GetAggregateXMinutesByMetricIDAndHostname(ts.Parameters.PGTx, clusterID, metricID, minutes, hostname)

	} else if ts.Parameters.DBType == "cassandra" {
		return cassandra.NewTSMetric(ts.Parameters.CassandraSession).GetAggregateXMinutesByMetricIDAndHostname(clusterID, metricID, minutes, hostname)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}
