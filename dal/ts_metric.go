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

func (ts *TSMetric) CreateByHostRow(tx *sqlx.Tx, hostRow *HostRow) error {
	metricsMap, err := NewMetric(ts.db).AllByClusterIDAsMap(tx, hostRow.ClusterID)
	if err != nil {
		return err
	}

	// Loop through every host's data and see if they are part of graph metrics.
	// If they are, insert a record in ts_metrics.
	for path, data := range hostRow.DataAsFlatKeyValue() {
		for dataKey, value := range data {
			metricKey := path + "." + dataKey

			if metricID, ok := metricsMap[metricKey]; ok {
				// Deserialized JSON number -> interface{} always have float64 as type.
				if trueValueFloat64, ok := value.(float64); ok {
					// What the...
					toInt64AndBack := float64(int64(trueValueFloat64))
					if toInt64AndBack == trueValueFloat64 {
						trueValueFloat64 = toInt64AndBack
					}

					// Ignore error for now, there's no need to break the entire loop when one insert fails.
					ts.Create(tx, hostRow.ClusterID, metricID, metricKey, hostRow.Name, trueValueFloat64)
				}
			}
		}
	}
	return nil
}
