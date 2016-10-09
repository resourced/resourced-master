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

type TSMetricHighchartPayload struct {
	Name string          `json:"name"`
	Data [][]interface{} `json:"data"`
}

func (hcPayload *TSMetricHighchartPayload) GetName() string {
	return hcPayload.Name
}
func (hcPayload *TSMetricHighchartPayload) GetData() [][]interface{} {
	return hcPayload.Data
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

func (ts *TSMetric) metricRowsForHighchart(host string, tsMetricRows []*TSMetricRow) (*TSMetricHighchartPayload, error) {
	hcPayload := &TSMetricHighchartPayload{}
	hcPayload.Name = host
	hcPayload.Data = make([][]interface{}, len(tsMetricRows))

	for i, tsMetricRow := range tsMetricRows {
		row := make([]interface{}, 2)
		row[0] = tsMetricRow.Created / 1000000
		row[1] = tsMetricRow.Value

		hcPayload.Data[i] = row
	}

	return hcPayload, nil
}

func (ts *TSMetric) AllByMetricIDHostAndRange(clusterID, metricID int64, host string, from, to int64) ([]*TSMetricRow, error) {
	rows := []*TSMetricRow{}
	query := fmt.Sprintf(`SELECT cluster_id, metric_id, created, key, host, value FROM %v WHERE cluster_id=? AND metric_id=? AND host=? AND created >= ? AND created <= ? ORDER BY cluster_id,metric_id,created ASC`, ts.table)

	var scannedClusterID, scannedMetricID, scannedCreated int64
	var scannedKey, scannedHost string
	var scannedValue float64

	iter := ts.session.Query(query, clusterID, metricID, host, from, to).Iter()
	for iter.Scan(&scannedClusterID, &scannedMetricID, &scannedCreated, &scannedKey, &scannedHost, &scannedValue) {
		rows = append(rows, &TSMetricRow{
			ClusterID: scannedClusterID,
			MetricID:  scannedMetricID,
			Created:   scannedCreated,
			Key:       scannedKey,
			Host:      scannedHost,
			Value:     scannedValue,
		})
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{
			"Method":    "TSMetric.AllByMetricIDHostAndRange",
			"ClusterID": clusterID,
			"MetricID":  metricID,
			"Hostname":  host,
			"From":      from,
			"To":        to,
		}).Error(err)

		return nil, err
	}

	return rows, nil
}

func (ts *TSMetric) AllByMetricIDHostAndRangeForHighchart(clusterID, metricID int64, host string, from, to int64) (*TSMetricHighchartPayload, error) {
	tsMetricRows, err := ts.AllByMetricIDHostAndRange(clusterID, metricID, host, from, to)
	if err != nil {
		return nil, err
	}

	return ts.metricRowsForHighchart(host, tsMetricRows)
}

func (ts *TSMetric) AllByMetricIDAndRange(clusterID, metricID int64, from, to int64) ([]*TSMetricRow, error) {
	rows := []*TSMetricRow{}
	query := fmt.Sprintf(`SELECT * FROM %v WHERE cluster_id=? AND metric_id=? AND created >= ? AND created <= ? ORDER BY cluster_id,metric_id,created ASC`, ts.table)

	var scannedClusterID, scannedMetricID, scannedCreated int64
	var scannedKey, scannedHost string
	var scannedValue float64

	iter := ts.session.Query(query, clusterID, metricID, from, to).Iter()
	for iter.Scan(&scannedClusterID, &scannedMetricID, &scannedCreated, &scannedKey, &scannedHost, &scannedValue) {
		rows = append(rows, &TSMetricRow{
			ClusterID: scannedClusterID,
			MetricID:  scannedMetricID,
			Created:   scannedCreated,
			Key:       scannedKey,
			Host:      scannedHost,
			Value:     scannedValue,
		})
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{
			"Method":    "TSMetric.AllByMetricIDHostAndRange",
			"ClusterID": clusterID,
			"MetricID":  metricID,
			"From":      from,
			"To":        to,
		}).Error(err)

		return nil, err
	}

	return rows, nil
}

func (ts *TSMetric) AllByMetricIDAndRangeForHighchart(clusterID, metricID, from, to int64) ([]*TSMetricHighchartPayload, error) {
	tsMetricRows, err := ts.AllByMetricIDAndRange(clusterID, metricID, from, to)
	if err != nil {
		return nil, err
	}

	// Group all TSMetricRows per host
	mapHostsAndMetrics := make(map[string][]*TSMetricRow)

	for _, tsMetricRow := range tsMetricRows {
		host := tsMetricRow.Host

		if _, ok := mapHostsAndMetrics[host]; !ok {
			mapHostsAndMetrics[host] = make([]*TSMetricRow, 0)
		}

		mapHostsAndMetrics[host] = append(mapHostsAndMetrics[host], tsMetricRow)
	}

	// Then generate multiple Highchart payloads per all these hosts.
	highChartPayloads := make([]*TSMetricHighchartPayload, 0)

	for host, tsMetricRows := range mapHostsAndMetrics {
		highChartPayload, err := ts.metricRowsForHighchart(host, tsMetricRows)
		if err != nil {
			return nil, err
		}
		highChartPayloads = append(highChartPayloads, highChartPayload)
	}

	return highChartPayloads, nil
}
