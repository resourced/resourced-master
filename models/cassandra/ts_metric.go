package cassandra

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gocql/gocql"

	"github.com/resourced/resourced-master/models/shared"
)

func NewTSMetric(session *gocql.Session) *TSMetric {
	ts := &TSMetric{}
	ts.session = session
	ts.table = "ts_metrics"

	return ts
}

type TSMetricRow struct {
	ClusterID int64     `db:"cluster_id"`
	MetricID  int64     `db:"metric_id"`
	Created   time.Time `db:"created"`
	Deleted   time.Time `db:"deleted"`
	Key       string    `db:"key"`
	Host      string    `db:"host"`
	Value     float64   `db:"value"`
}

type TSMetric struct {
	Base
}

func (ts *TSMetric) CreateByHostRow(hostRow shared.IHostRow, metricsMap map[string]int64, ttl time.Duration) error {
	// Loop through every host's data and see if they are part of graph metrics.
	// If they are, insert a record in ts_metrics.
	for path, data := range hostRow.DataAsFlatKeyValue() {
		for dataKey, value := range data {
			metricKey := path + "." + dataKey

			if metricID, ok := metricsMap[metricKey]; ok {
				// Deserialized JSON number -> interface{} always have float64 as type.
				if trueValueFloat64, ok := value.(float64); ok {
					// Ignore error for now, there's no need to break the entire loop when one insert fails.
					err := ts.session.Query(
						fmt.Sprintf(`INSERT INTO %v (cluster_id, metric_id, key, host, value, created) VALUES (?, ?, ?, ?, ?, ?) USING TTL ?`, ts.table),
						hostRow.GetClusterID(),
						metricID,
						metricKey,
						hostRow.GetHostname(),
						trueValueFloat64,
						time.Now().UTC().Unix(),
						ttl,
					).Exec()

					if err != nil {
						logrus.WithFields(logrus.Fields{
							"Method":    "TSMetric.CreateByHostRow",
							"ClusterID": hostRow.GetClusterID(),
							"MetricID":  metricID,
							"MetricKey": metricKey,
							"Hostname":  hostRow.GetHostname(),
							"Value":     trueValueFloat64,
						}).Error(err)
					}
				}
			}
		}
	}
	return nil
}
