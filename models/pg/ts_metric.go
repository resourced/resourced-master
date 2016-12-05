package pg

import (
	"context"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"

	"github.com/resourced/resourced-master/contexthelper"
	"github.com/resourced/resourced-master/models/shared"
)

func NewTSMetric(ctx context.Context, clusterID int64) *TSMetric {
	ts := &TSMetric{}
	ts.AppContext = ctx
	ts.table = "ts_metrics"
	ts.clusterID = clusterID
	ts.i = ts

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
	TSBase
}

func (ts *TSMetric) GetPGDB() (*sqlx.DB, error) {
	pgdbs, err := contexthelper.GetPGDBConfig(ts.AppContext)
	if err != nil {
		return nil, err
	}
	if pgdbs == nil {
		return nil, fmt.Errorf("Database handler went missing")
	}

	return pgdbs.GetTSMetric(ts.clusterID), nil
}

func (ts *TSMetric) metricRowsForHighchart(tx *sqlx.Tx, host string, tsMetricRows []*TSMetricRow) (*shared.TSMetricHighchartPayload, error) {
	hcPayload := &shared.TSMetricHighchartPayload{}
	hcPayload.Name = host
	hcPayload.Data = make([][]interface{}, len(tsMetricRows))

	for i, tsMetricRow := range tsMetricRows {
		row := make([]interface{}, 2)
		row[0] = tsMetricRow.Created.UnixNano() / 1000000
		row[1] = tsMetricRow.Value

		hcPayload.Data[i] = row
	}

	return hcPayload, nil
}

// Create a new record.
func (ts *TSMetric) Create(tx *sqlx.Tx, clusterID, metricID int64, host, key string, value float64, deletedFrom int64) error {
	insertData := make(map[string]interface{})
	insertData["cluster_id"] = clusterID
	insertData["metric_id"] = metricID
	insertData["key"] = key
	insertData["host"] = host
	insertData["value"] = value
	insertData["deleted"] = time.Unix(deletedFrom, 0).UTC()

	_, err := ts.InsertIntoTable(tx, insertData)
	return err
}

func (ts *TSMetric) CreateByHostRow(tx *sqlx.Tx, hostRow shared.IHostRow, metricsMap map[string]int64, deletedFrom int64) error {
	// Loop through every host's data and see if they are part of graph metrics.
	// If they are, insert a record in ts_metrics.
	// for path, data := range hostRow.DataAsFlatKeyValue() {
	// 	for dataKey, value := range data {
	// 		metricKey := path + "." + dataKey

	// 		if metricID, ok := metricsMap[metricKey]; ok {
	// 			// Deserialized JSON number -> interface{} always have float64 as type.
	// 			if trueValueFloat64, ok := value.(float64); ok {
	// 				// Ignore error for now, there's no need to break the entire loop when one insert fails.
	// 				err := ts.Create(tx, hostRow.GetClusterID(), metricID, hostRow.GetHostname(), metricKey, trueValueFloat64, deletedFrom)
	// 				if err != nil {
	// 					logrus.WithFields(logrus.Fields{
	// 						"Method":    "TSMetric.Create",
	// 						"ClusterID": hostRow.GetClusterID(),
	// 						"MetricID":  metricID,
	// 						"MetricKey": metricKey,
	// 						"Hostname":  hostRow.GetHostname(),
	// 						"Value":     trueValueFloat64,
	// 					}).Error(err)
	// 				}
	// 			}
	// 		}
	// 	}
	// }
	return nil
}

func (ts *TSMetric) AllByMetricIDHostAndRange(tx *sqlx.Tx, clusterID, metricID int64, host string, from, to, deletedFrom int64) ([]*TSMetricRow, error) {
	pgdb, err := ts.GetPGDB()
	if err != nil {
		return nil, err
	}

	rows := []*TSMetricRow{}
	query := fmt.Sprintf(`SELECT * FROM %v WHERE cluster_id=$1 AND metric_id=$2 AND host=$3 AND
created >= to_timestamp($4) at time zone 'utc' AND
created <= to_timestamp($5) at time zone 'utc' AND
deleted >= to_timestamp($6) at time zone 'utc'
ORDER BY cluster_id,metric_id,created ASC`, ts.table)

	err = pgdb.Select(&rows, query, clusterID, metricID, host, from, to, deletedFrom)

	if err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)

		logrus.WithFields(logrus.Fields{
			"Method":    "TSMetric.AllByMetricIDHostAndRange",
			"ClusterID": clusterID,
			"MetricID":  metricID,
			"Hostname":  host,
			"From":      from,
			"To":        to,
			"Deleted":   deletedFrom,
		}).Error(err)
	}
	return rows, err
}

func (ts *TSMetric) AllByMetricIDHostAndRangeForHighchart(tx *sqlx.Tx, clusterID, metricID int64, host string, from, to, deletedFrom int64) (*shared.TSMetricHighchartPayload, error) {
	tsMetricRows, err := ts.AllByMetricIDHostAndRange(tx, clusterID, metricID, host, from, to, deletedFrom)
	if err != nil {
		return nil, err
	}

	return ts.metricRowsForHighchart(tx, host, tsMetricRows)
}

func (ts *TSMetric) AllByMetricIDAndRange(tx *sqlx.Tx, clusterID, metricID, from, to, deletedFrom int64) ([]*TSMetricRow, error) {
	pgdb, err := ts.GetPGDB()
	if err != nil {
		return nil, err
	}

	rows := []*TSMetricRow{}
	query := fmt.Sprintf(`SELECT * FROM %v WHERE cluster_id=$1 AND metric_id=$2 AND
created >= to_timestamp($3) at time zone 'utc' AND
created <= to_timestamp($4) at time zone 'utc' AND
deleted >= to_timestamp($5) at time zone 'utc'
ORDER BY cluster_id,metric_id,created ASC`, ts.table)

	err = pgdb.Select(&rows, query, clusterID, metricID, from, to, deletedFrom)

	if err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
	}
	return rows, err
}

func (ts *TSMetric) AllByMetricIDAndRangeForHighchart(tx *sqlx.Tx, clusterID, metricID, from, to, deletedFrom int64) ([]*shared.TSMetricHighchartPayload, error) {
	tsMetricRows, err := ts.AllByMetricIDAndRange(tx, clusterID, metricID, from, to, deletedFrom)
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
	highChartPayloads := make([]*shared.TSMetricHighchartPayload, 0)

	for host, tsMetricRows := range mapHostsAndMetrics {
		highChartPayload, err := ts.metricRowsForHighchart(tx, host, tsMetricRows)
		if err != nil {
			return nil, err
		}
		highChartPayloads = append(highChartPayloads, highChartPayload)
	}

	return highChartPayloads, nil
}

func (ts *TSMetric) GetAggregateXMinutesByMetricIDAndHostname(tx *sqlx.Tx, clusterID, metricID int64, minutes int, hostname string) (*shared.TSMetricAggregateRow, error) {
	pgdb, err := ts.GetPGDB()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	from := now.Add(-1 * time.Duration(minutes) * time.Minute).UTC().Unix()

	row := &shared.TSMetricAggregateRow{}
	query := fmt.Sprintf("SELECT cluster_id, host, key, avg(value) as avg, max(value) as max, min(value) as min, sum(value) as sum FROM %v WHERE cluster_id=$1 AND metric_id=$2 AND created >= to_timestamp($2) at time zone 'utc' AND host=$3 GROUP BY cluster_id, metric_id, host", ts.table)
	err = pgdb.Get(row, query, clusterID, metricID, from, hostname)

	if err != nil {
		err = fmt.Errorf("%v. Query: %v, ClusterID: %v, Hostname: %v, MetricID: %v", err.Error(), query, clusterID, hostname, metricID)
	}
	return row, err
}
