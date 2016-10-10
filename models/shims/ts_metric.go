package shims

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/resourced/resourced-master/contexthelper"
	"github.com/resourced/resourced-master/models/cassandra"
	"github.com/resourced/resourced-master/models/pg"
	"github.com/resourced/resourced-master/models/shared"
)

func NewTSMetric(ctx context.Context, clusterID int64) *TSMetric {
	ts := &TSMetric{}
	ts.AppContext = ctx
	ts.ClusterID = clusterID
	return ts
}

type TSMetric struct {
	Base
	ClusterID int64
}

func (ts *TSMetric) GetDBType() string {
	generalConfig, err := contexthelper.GetGeneralConfig(ts.AppContext)
	if err != nil {
		return ""
	}

	return generalConfig.GetMetricsDB()
}

func (ts *TSMetric) GetPGDB() (*sqlx.DB, error) {
	pgdbs, err := contexthelper.GetPGDBConfig(ts.AppContext)
	if err != nil {
		return nil, err
	}

	return pgdbs.GetTSMetric(ts.ClusterID), nil
}

func (ts *TSMetric) CreateByHostRow(hostRow shared.IHostRow, metricsMap map[string]int64, deletedFrom int64, ttl time.Duration) error {
	if ts.GetDBType() == "pg" {
		pgdb, err := ts.GetPGDB()
		if err != nil {
			return err
		}
		return pg.NewTSMetric(pgdb).CreateByHostRow(nil, hostRow, metricsMap, deletedFrom)

	} else if ts.GetDBType() == "cassandra" {
		return cassandra.NewTSMetric(ts.AppContext).CreateByHostRow(hostRow, metricsMap, ttl)
	}

	return fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

func (ts *TSMetric) AllByMetricIDHostAndRangeForHighchart(clusterID, metricID int64, host string, from, to, deletedFrom int64) (*shared.TSMetricHighchartPayload, error) {
	if ts.GetDBType() == "pg" {
		pgdb, err := ts.GetPGDB()
		if err != nil {
			return nil, err
		}
		return pg.NewTSMetric(pgdb).AllByMetricIDHostAndRangeForHighchart(nil, clusterID, metricID, host, from, to, deletedFrom)

	} else if ts.GetDBType() == "cassandra" {
		return cassandra.NewTSMetric(ts.AppContext).AllByMetricIDHostAndRangeForHighchart(clusterID, metricID, host, from, to)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

func (ts *TSMetric) AllByMetricIDAndRangeForHighchart(clusterID, metricID, from, to, deletedFrom int64) ([]*shared.TSMetricHighchartPayload, error) {
	if ts.GetDBType() == "pg" {
		pgdb, err := ts.GetPGDB()
		if err != nil {
			return nil, err
		}
		return pg.NewTSMetric(pgdb).AllByMetricIDAndRangeForHighchart(nil, clusterID, metricID, from, to, deletedFrom)

	} else if ts.GetDBType() == "cassandra" {
		return cassandra.NewTSMetric(ts.AppContext).AllByMetricIDAndRangeForHighchart(clusterID, metricID, from, to)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

func (ts *TSMetric) GetAggregateXMinutesByMetricIDAndHostname(clusterID, metricID int64, minutes int, hostname string) (*shared.TSMetricAggregateRow, error) {
	if ts.GetDBType() == "pg" {
		pgdb, err := ts.GetPGDB()
		if err != nil {
			return nil, err
		}
		return pg.NewTSMetric(pgdb).GetAggregateXMinutesByMetricIDAndHostname(nil, clusterID, metricID, minutes, hostname)

	} else if ts.GetDBType() == "cassandra" {
		return cassandra.NewTSMetric(ts.AppContext).GetAggregateXMinutesByMetricIDAndHostname(clusterID, metricID, minutes, hostname)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}
