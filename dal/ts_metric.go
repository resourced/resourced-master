package dal

import (
	"github.com/jmoiron/sqlx"
	"time"
)

func NewTSMetric(db *sqlx.DB) *TSMetric {
	ts := &TSMetric{}
	ts.db = db
	ts.table = "ts_metrics"

	return ts
}

type TSMetricRow struct {
	ClusterID int64     `db:"cluster_id"`
	MetricID  int64     `db:"metric_id"`
	Created   time.Time `db:"created"`
	Key       string    `db:"key"`
	Host      string    `db:"host"`
	Value     float64   `db:"value"`
}

type TSMetric struct {
	Base
}

// Create a new record.
func (ts *TSMetric) Create(tx *sqlx.Tx, clusterID, metricID int64, key, host string, value float64) error {
	insertData := make(map[string]interface{})
	insertData["cluster_id"] = clusterID
	insertData["metric_id"] = metricID
	insertData["key"] = key
	insertData["host"] = host
	insertData["value"] = value

	_, err := ts.InsertIntoTable(tx, insertData)
	return err
}
