package cassandra

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gocql/gocql"

	"github.com/resourced/resourced-master/contexthelper"
	"github.com/resourced/resourced-master/models/shared"
)

func NewTSMetric(ctx context.Context) *TSMetric {
	ts := &TSMetric{}
	ts.AppContext = ctx
	ts.table = "ts_metrics"

	return ts
}

type TSMetric struct {
	Base
}

func (ts *TSMetric) GetCassandraSession() (*gocql.Session, error) {
	cassandradbs, err := contexthelper.GetCassandraDBConfig(ts.AppContext)
	if err != nil {
		return nil, err
	}

	return cassandradbs.TSMetricSession, nil
}

func (ts *TSMetric) CreateByHostRow(hostRow shared.IHostRow, metricsMap map[string]int64, ttl time.Duration) error {
	session, err := ts.GetCassandraSession()
	if err != nil {
		return err
	}

	// Loop through every host's data and see if they are part of graph metrics.
	// If they are, insert a record in ts_metrics.
	for metricKey, value := range hostRow.GetData() {
		if metricID, ok := metricsMap[metricKey]; ok {

			// Unfortunately, value is always string because Cassandra map does not support mixed types.
			value64, err := strconv.ParseFloat(value, 64)
			if err == nil {
				// Ignore error for now, there's no need to break the entire loop when one insert fails.
				err := session.Query(
					fmt.Sprintf(`INSERT INTO %v (cluster_id, metric_id, key, host, value, created) VALUES (?, ?, ?, ?, ?, ?) USING TTL ?`, ts.table),
					hostRow.GetClusterID(),
					metricID,
					metricKey,
					hostRow.GetHostname(),
					value64,
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
						"Value":     value64,
					}).Error(err)
				}
			}
		}
	}
	return nil
}

func (ts *TSMetric) metricRowsForHighchart(host string, tsMetricRows []*shared.TSMetricRow) (*shared.TSMetricHighchartPayload, error) {
	hcPayload := &shared.TSMetricHighchartPayload{}
	hcPayload.Name = host
	hcPayload.Data = make([][]interface{}, len(tsMetricRows))

	for i, tsMetricRow := range tsMetricRows {
		row := make([]interface{}, 2)
		row[0] = tsMetricRow.Created * 1000
		row[1] = tsMetricRow.Value

		hcPayload.Data[i] = row
	}

	return hcPayload, nil
}

func (ts *TSMetric) AllByMetricIDHostAndRange(clusterID, metricID int64, host string, from, to int64) ([]*shared.TSMetricRow, error) {
	session, err := ts.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	rows := []*shared.TSMetricRow{}
	query := fmt.Sprintf(`SELECT cluster_id, metric_id, created, key, host, value FROM %v WHERE cluster_id=? AND metric_id=? AND host=? AND created >= ? AND created <= ? ORDER BY created ASC ALLOW FILTERING`, ts.table)

	var scannedClusterID, scannedMetricID, scannedCreated int64
	var scannedKey, scannedHost string
	var scannedValue float64

	iter := session.Query(query, clusterID, metricID, host, from, to).Iter()
	for iter.Scan(&scannedClusterID, &scannedMetricID, &scannedCreated, &scannedKey, &scannedHost, &scannedValue) {
		rows = append(rows, &shared.TSMetricRow{
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

func (ts *TSMetric) AllByMetricIDHostAndRangeForHighchart(clusterID, metricID int64, host string, from, to int64) (*shared.TSMetricHighchartPayload, error) {
	tsMetricRows, err := ts.AllByMetricIDHostAndRange(clusterID, metricID, host, from, to)
	if err != nil {
		return nil, err
	}

	return ts.metricRowsForHighchart(host, tsMetricRows)
}

func (ts *TSMetric) AllByMetricIDAndRange(clusterID, metricID int64, from, to int64) ([]*shared.TSMetricRow, error) {
	session, err := ts.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	rows := []*shared.TSMetricRow{}
	query := fmt.Sprintf(`SELECT * FROM %v WHERE cluster_id=? AND metric_id=? AND created >= ? AND created <= ? ORDER BY created ASC`, ts.table)

	var scannedClusterID, scannedMetricID, scannedCreated int64
	var scannedKey, scannedHost string
	var scannedValue float64

	iter := session.Query(query, clusterID, metricID, from, to).Iter()
	for iter.Scan(&scannedClusterID, &scannedMetricID, &scannedCreated, &scannedKey, &scannedHost, &scannedValue) {
		rows = append(rows, &shared.TSMetricRow{
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
			"Method":    "TSMetric.AllByMetricIDAndRange",
			"ClusterID": clusterID,
			"MetricID":  metricID,
			"From":      from,
			"To":        to,
		}).Error(err)

		return nil, err
	}

	return rows, nil
}

func (ts *TSMetric) AllByMetricIDAndRangeForHighchart(clusterID, metricID, from, to int64) ([]*shared.TSMetricHighchartPayload, error) {
	tsMetricRows, err := ts.AllByMetricIDAndRange(clusterID, metricID, from, to)
	if err != nil {
		return nil, err
	}

	// Group all shared.TSMetricRows per host
	mapHostsAndMetrics := make(map[string][]*shared.TSMetricRow)

	for _, tsMetricRow := range tsMetricRows {
		host := tsMetricRow.Host

		if _, ok := mapHostsAndMetrics[host]; !ok {
			mapHostsAndMetrics[host] = make([]*shared.TSMetricRow, 0)
		}

		mapHostsAndMetrics[host] = append(mapHostsAndMetrics[host], tsMetricRow)
	}

	// Then generate multiple Highchart payloads per all these hosts.
	highChartPayloads := make([]*shared.TSMetricHighchartPayload, 0)

	for host, tsMetricRows := range mapHostsAndMetrics {
		highChartPayload, err := ts.metricRowsForHighchart(host, tsMetricRows)
		if err != nil {
			return nil, err
		}
		highChartPayloads = append(highChartPayloads, highChartPayload)
	}

	return highChartPayloads, nil
}

func (ts *TSMetric) GetAggregateXMinutesByMetricIDAndHostname(clusterID, metricID int64, minutes int, hostname string) (*shared.TSMetricAggregateRow, error) {
	session, err := ts.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	from := now.Add(-1 * time.Duration(minutes) * time.Minute).UTC().Unix()

	var row *shared.TSMetricAggregateRow
	query := fmt.Sprintf("SELECT cluster_id, host, key, avg(value) as avg, max(value) as max, min(value) as min, sum(value) as sum FROM %v WHERE cluster_id=? AND metric_id=? AND created >= ? AND host=? ALLOW FILTERING", ts.table)

	var scannedClusterID int64
	var scannedAvg, scannedMax, scannedMin, scannedSum float64
	var scannedKey, scannedHost string

	iter := session.Query(query, clusterID, metricID, from, hostname).Iter()
	for iter.Scan(&scannedClusterID, &scannedHost, &scannedKey, &scannedAvg, &scannedMax, &scannedMin, &scannedSum) {
		row = &shared.TSMetricAggregateRow{
			ClusterID: scannedClusterID,
			Key:       scannedKey,
			Host:      scannedHost,
			Avg:       scannedAvg,
			Max:       scannedMax,
			Min:       scannedMin,
			Sum:       scannedSum,
		}
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{
			"Method":    "TSMetric.GetAggregateXMinutesByMetricIDAndHostname",
			"ClusterID": clusterID,
			"MetricID":  metricID,
			"From":      from,
			"Hostname":  hostname,
		}).Error(err)

		return nil, err
	}

	return row, nil
}
