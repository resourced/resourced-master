package dal

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
)

func NewTSMetric(db *sqlx.DB) *TSMetric {
	ts := &TSMetric{}
	ts.db = db
	ts.table = "ts_metrics"

	return ts
}

type TSMetricHighchartPayload struct {
	Name string          `json:"name"`
	Data [][]interface{} `json:"data"`
}

type TSMetricSelectAggregateRow struct {
	ClusterID   int64   `db:"cluster_id"`
	CreatedUnix int64   `db:"created_unix"`
	Key         string  `db:"key"`
	Host        string  `db:"host"`
	Avg         float64 `db:"avg"`
	Max         float64 `db:"max"`
	Min         float64 `db:"min"`
	Sum         float64 `db:"sum"`
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
	TSBase
}

func (ts *TSMetric) metricRowsForHighchart(tx *sqlx.Tx, host string, tsMetricRows []*TSMetricRow) (*TSMetricHighchartPayload, error) {
	hcPayload := &TSMetricHighchartPayload{}
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

func (ts *TSMetric) CreateByHostRow(tx *sqlx.Tx, hostRow *HostRow, metricsMap map[string]int64) error {
	tsAggr15m := NewTSMetricAggr15m(ts.db)

	// Loop through every host's data and see if they are part of graph metrics.
	// If they are, insert a record in ts_metrics.
	for path, data := range hostRow.DataAsFlatKeyValue() {
		for dataKey, value := range data {
			metricKey := path + "." + dataKey

			if metricID, ok := metricsMap[metricKey]; ok {
				// Deserialized JSON number -> interface{} always have float64 as type.
				if trueValueFloat64, ok := value.(float64); ok {
					// Ignore error for now, there's no need to break the entire loop when one insert fails.
					err := ts.Create(tx, hostRow.ClusterID, metricID, metricKey, hostRow.Hostname, trueValueFloat64)
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"Method":    "TSMetric.Create",
							"ClusterID": hostRow.ClusterID,
							"MetricID":  metricID,
							"MetricKey": metricKey,
							"Hostname":  hostRow.Hostname,
							"Value":     trueValueFloat64,
						}).Error(err)
					}

					// Aggregate avg,max,min,sum values per 15 minutes.
					go func() {
						selectAggrRows, err := ts.AggregateEvery(tx, hostRow.ClusterID, "15 minute")
						if err != nil {
							logrus.Error(err)
						} else {
							aggrTx, err := ts.db.Beginx()
							if err != nil {
								logrus.Error(err)
							}

							// Store those 15 minutes aggregation values
							for _, selectAggrRow := range selectAggrRows {
								err = tsAggr15m.InsertOrUpdate(aggrTx, hostRow.ClusterID, metricID, metricKey, selectAggrRow)
								if err != nil {
									logrus.WithFields(logrus.Fields{
										"Method":    "tsAggr15m.InsertOrUpdate",
										"ClusterID": hostRow.ClusterID,
										"MetricID":  metricID,
										"MetricKey": metricKey,
									}).Error(err)
								}
							}

							err = aggrTx.Commit()
							if err != nil {
								logrus.Error(err)
							}
						}
					}()

					// Aggregate avg,max,min,sum values per 15 minutes per host.
					go func() {
						selectAggrRows, err := ts.AggregateEveryPerHost(tx, hostRow.ClusterID, "15 minute")
						if err != nil {
							logrus.Error(err)
						} else {
							aggrTx, err := ts.db.Beginx()
							if err != nil {
								logrus.Error(err)
							}

							// Store those 15 minutes aggregation values per host
							for _, selectAggrRow := range selectAggrRows {
								err = tsAggr15m.InsertOrUpdate(aggrTx, hostRow.ClusterID, metricID, metricKey, selectAggrRow)
								if err != nil {
									logrus.WithFields(logrus.Fields{
										"Method":    "tsAggr15m.InsertOrUpdate",
										"ClusterID": hostRow.ClusterID,
										"MetricID":  metricID,
										"MetricKey": metricKey,
									}).Error(err)
								}
							}

							err = aggrTx.Commit()
							if err != nil {
								logrus.Error(err)
							}
						}
					}()
				}
			}
		}
	}
	return nil
}

func (ts *TSMetric) AggregateEvery(tx *sqlx.Tx, clusterID int64, interval string) ([]*TSMetricSelectAggregateRow, error) {
	if interval == "" {
		interval = "15 minute"
	}

	rows := []*TSMetricSelectAggregateRow{}
	query := fmt.Sprintf("SELECT cluster_id, cast(CEILING(extract('epoch' from created)/900)*900 as bigint) AS created_unix, key, avg(value) as avg, max(value) as max, min(value) as min, sum(value) as sum FROM %v WHERE cluster_id=$1 AND created >= (NOW() at time zone 'utc' - INTERVAL '%v') GROUP BY cluster_id, created_unix, key ORDER BY created_unix ASC", ts.table, interval)
	err := ts.db.Select(&rows, query, clusterID)

	if err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
	}
	return rows, err
}

func (ts *TSMetric) AggregateEveryPerHost(tx *sqlx.Tx, clusterID int64, interval string) ([]*TSMetricSelectAggregateRow, error) {
	if interval == "" {
		interval = "15 minute"
	}

	rows := []*TSMetricSelectAggregateRow{}
	query := fmt.Sprintf("SELECT cluster_id, cast(CEILING(extract('epoch' from created)/900)*900 as bigint) AS created_unix, host, key, avg(value) as avg, max(value) as max, min(value) as min, sum(value) as sum FROM %v WHERE cluster_id=$1 AND created >= (NOW() at time zone 'utc' - INTERVAL '%v') GROUP BY cluster_id, created_unix, host, key ORDER BY created_unix ASC", ts.table, interval)
	err := ts.db.Select(&rows, query, clusterID)

	if err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
	}
	return rows, err
}

func (ts *TSMetric) AllByMetricIDHostAndRange(tx *sqlx.Tx, clusterID, metricID int64, host string, from, to int64) ([]*TSMetricRow, error) {
	rows := []*TSMetricRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND metric_id=$2 AND host=$3 AND created >= to_timestamp($4) at time zone 'utc' AND created <= to_timestamp($5) at time zone 'utc' ORDER BY cluster_id,metric_id,created ASC", ts.table)
	err := ts.db.Select(&rows, query, clusterID, metricID, host, from, to)

	if err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
	}
	return rows, err
}

func (ts *TSMetric) AllByMetricIDHostAndRangeForHighchart(tx *sqlx.Tx, clusterID, metricID int64, host string, from, to int64) (*TSMetricHighchartPayload, error) {
	tsMetricRows, err := ts.AllByMetricIDHostAndRange(tx, clusterID, metricID, host, from, to)
	if err != nil {
		return nil, err
	}

	return ts.metricRowsForHighchart(tx, host, tsMetricRows)
}

func (ts *TSMetric) AllByMetricIDAndRange(tx *sqlx.Tx, clusterID, metricID int64, from, to int64) ([]*TSMetricRow, error) {
	rows := []*TSMetricRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND metric_id=$2 AND created >= to_timestamp($3) at time zone 'utc' AND created <= to_timestamp($4) at time zone 'utc' ORDER BY cluster_id,metric_id,created ASC", ts.table)
	err := ts.db.Select(&rows, query, clusterID, metricID, from, to)

	if err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
	}
	return rows, err
}

func (ts *TSMetric) AllByMetricIDAndRangeForHighchart(tx *sqlx.Tx, clusterID, metricID int64, from, to int64) ([]*TSMetricHighchartPayload, error) {
	tsMetricRows, err := ts.AllByMetricIDAndRange(tx, clusterID, metricID, from, to)
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
		highChartPayload, err := ts.metricRowsForHighchart(tx, host, tsMetricRows)
		if err != nil {
			return nil, err
		}
		highChartPayloads = append(highChartPayloads, highChartPayload)
	}

	return highChartPayloads, nil
}
