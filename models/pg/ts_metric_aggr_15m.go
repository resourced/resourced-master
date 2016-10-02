package pg

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
)

func NewTSMetricAggr15m(db *sqlx.DB) *TSMetricAggr15m {
	ts := &TSMetricAggr15m{}
	ts.db = db
	ts.table = "ts_metrics_aggr_15m"

	return ts
}

type TSMetricAggr15mRow struct {
	ClusterID int64          `db:"cluster_id"`
	MetricID  int64          `db:"metric_id"`
	Created   time.Time      `db:"created"`
	Deleted   time.Time      `db:"deleted"`
	Key       string         `db:"key"`
	Host      sql.NullString `db:"host"`
	Avg       float64        `db:"avg"`
	Max       float64        `db:"max"`
	Min       float64        `db:"min"`
	Sum       float64        `db:"sum"`
}

type TSMetricAggr15m struct {
	TSBase
}

func (ts *TSMetricAggr15m) metricRowsForHighchart(tx *sqlx.Tx, host string, tsMetricAggrRows []*TSMetricAggr15mRow, aggr string) (*TSMetricHighchartPayload, error) {
	hcPayload := &TSMetricHighchartPayload{}
	hcPayload.Name = host
	hcPayload.Data = make([][]interface{}, len(tsMetricAggrRows))

	for i, tsMetricAggrRow := range tsMetricAggrRows {
		row := make([]interface{}, 2)
		row[0] = tsMetricAggrRow.Created.UnixNano() / 1000000 // in seconds

		if aggr == "max" || aggr == "Max" {
			row[1] = tsMetricAggrRow.Max
		} else if aggr == "min" || aggr == "Min" {
			row[1] = tsMetricAggrRow.Min
		} else if aggr == "sum" || aggr == "Sum" {
			row[1] = tsMetricAggrRow.Sum
		} else if aggr == "avg" || aggr == "Avg" {
			row[1] = tsMetricAggrRow.Avg
		} else {
			all := make(map[string]interface{})
			all["max"] = tsMetricAggrRow.Max
			all["min"] = tsMetricAggrRow.Min
			all["sum"] = tsMetricAggrRow.Sum
			all["avg"] = tsMetricAggrRow.Avg

			row[1] = all
		}

		hcPayload.Data[i] = row
	}

	return hcPayload, nil
}

// CreateByHostRow creates new rows given host data.
func (ts *TSMetricAggr15m) CreateByHostRow(tx *sqlx.Tx, hostRow *HostRow, metricsMap map[string]int64, selectAggrRows []*TSMetricSelectAggregateRow, deletedFrom int64) error {
	// Loop through every host's data and see if they are part of graph metrics.
	// If they are, insert a record in ts_metrics_aggr_15m.
	for path, data := range hostRow.DataAsFlatKeyValue() {
		for dataKey, _ := range data {
			metricKey := path + "." + dataKey

			if metricID, ok := metricsMap[metricKey]; ok {
				aggrTx, err := ts.db.Beginx()
				if err != nil {
					return err
				}

				err = ts.InsertOrUpdateMany(aggrTx, hostRow.ClusterID, metricID, metricKey, selectAggrRows, deletedFrom)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"Method":    "TSMetricAggr15m.InsertOrUpdate",
						"ClusterID": hostRow.ClusterID,
						"MetricID":  metricID,
						"MetricKey": metricKey,
					}).Error(err)
				}

				err = aggrTx.Commit()
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// CreateByHostRowPerHost creates new rows given host data per host.
func (ts *TSMetricAggr15m) CreateByHostRowPerHost(tx *sqlx.Tx, hostRow *HostRow, metricsMap map[string]int64, selectAggrRows []*TSMetricSelectAggregateRow, deletedFrom int64) error {
	// Loop through every host's data and see if they are part of graph metrics.
	// If they are, insert a record in ts_metrics_aggr_15m.
	for path, data := range hostRow.DataAsFlatKeyValue() {
		for dataKey, _ := range data {
			metricKey := path + "." + dataKey

			if metricID, ok := metricsMap[metricKey]; ok {
				aggrTx, err := ts.db.Beginx()
				if err != nil {
					logrus.Error(err)
				}

				// Store those 15 minutes aggregation values per host
				err = ts.InsertOrUpdateMany(aggrTx, hostRow.ClusterID, metricID, metricKey, selectAggrRows, deletedFrom)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"Method":    "TSMetricAggr15m.InsertOrUpdate",
						"ClusterID": hostRow.ClusterID,
						"MetricID":  metricID,
						"MetricKey": metricKey,
					}).Error(err)
				}

				err = aggrTx.Commit()
				if err != nil {
					logrus.Error(err)
				}
			}
		}
	}
	return nil
}

// InsertOrUpdateMany multiple records.
func (ts *TSMetricAggr15m) InsertOrUpdateMany(tx *sqlx.Tx, clusterID, metricID int64, metricKey string, selectAggrRows []*TSMetricSelectAggregateRow, deletedFrom int64) (err error) {
	for _, selectAggrRow := range selectAggrRows {
		err = ts.InsertOrUpdate(tx, clusterID, metricID, metricKey, selectAggrRow, deletedFrom)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Method":    "TSMetricAggr15m.InsertOrUpdate",
				"ClusterID": clusterID,
				"MetricID":  metricID,
				"MetricKey": metricKey,
			}).Error(err)

			return err
		}
	}

	return nil
}

// InsertOrUpdate a new record.
func (ts *TSMetricAggr15m) InsertOrUpdate(tx *sqlx.Tx, clusterID, metricID int64, metricKey string, selectAggrRow *TSMetricSelectAggregateRow, deletedFrom int64) (err error) {
	// Check if metricKey is correct, if not don't do anything
	if metricKey != selectAggrRow.Key {
		return nil
	}
	if selectAggrRow == nil {
		return errors.New("Aggregate row cannot be nil")
	}

	created := time.Unix(int64(selectAggrRow.CreatedUnix), 0).UTC()

	data := make(map[string]interface{})
	data["cluster_id"] = clusterID
	data["metric_id"] = metricID
	data["created"] = created
	data["key"] = selectAggrRow.Key
	data["avg"] = selectAggrRow.Avg
	data["max"] = selectAggrRow.Max
	data["min"] = selectAggrRow.Min
	data["sum"] = selectAggrRow.Sum
	data["deleted"] = time.Unix(deletedFrom, 0).UTC()

	if selectAggrRow.Host != "" {
		data["host"] = selectAggrRow.Host
	}

	aggrSelectRows := make([]*TSMetricAggr15mRow, 0)
	var query string

	if selectAggrRow.Host != "" {
		query = fmt.Sprintf("SELECT * from %v WHERE cluster_id=$1 AND created=$2 AND key=$3 AND host=$4 LIMIT 1", ts.table)
		err = ts.db.Select(&aggrSelectRows, query, clusterID, created, selectAggrRow.Key, selectAggrRow.Host)

	} else {
		query = fmt.Sprintf("SELECT * from %v WHERE cluster_id=$1 AND created=$2 AND key=$3 LIMIT 1", ts.table)
		err = ts.db.Select(&aggrSelectRows, query, clusterID, created, selectAggrRow.Key)
	}

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Method":    "TSMetricAggr15m.InsertOrUpdate.Select",
			"Created":   created,
			"ClusterID": clusterID,
			"MetricID":  metricID,
			"MetricKey": metricKey,
			"Query":     query,
		}).Error(err)

		return err
	}

	if err != nil || len(aggrSelectRows) == 0 {
		_, err = ts.InsertIntoTable(tx, data)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Method":    "TSMetricAggr15m.InsertOrUpdate.Insert",
				"Created":   created,
				"ClusterID": clusterID,
				"MetricID":  metricID,
				"MetricKey": metricKey,
			}).Error(err)
		}

	} else if selectAggrRow.Host != "" {
		query := fmt.Sprintf(`UPDATE %v SET avg=$1,max=$2,min=$3,sum=$4 WHERE cluster_id=$5 AND created=$6 AND key=$7 AND host=$8`, ts.table)
		_, err = tx.Exec(query, selectAggrRow.Avg, selectAggrRow.Max, selectAggrRow.Min, selectAggrRow.Sum, clusterID, created, selectAggrRow.Key, selectAggrRow.Host)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Method":    "TSMetricAggr15m.InsertOrUpdate.Update",
				"Created":   created,
				"ClusterID": clusterID,
				"MetricID":  metricID,
				"MetricKey": metricKey,
				"Host":      selectAggrRow.Host,
				"Query":     query,
			}).Error(err)
		}

	} else {
		query := fmt.Sprintf(`UPDATE %v SET avg=$1,max=$2,min=$3,sum=$4 WHERE cluster_id=$5 AND created=$6 AND key=$7`, ts.table)
		_, err = tx.Exec(query, selectAggrRow.Avg, selectAggrRow.Max, selectAggrRow.Min, selectAggrRow.Sum, clusterID, created, selectAggrRow.Key)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Method":    "TSMetricAggr15m.InsertOrUpdate.Update",
				"Created":   created,
				"ClusterID": clusterID,
				"MetricID":  metricID,
				"MetricKey": metricKey,
				"Query":     query,
			}).Error(err)
		}
	}

	return err
}

func (ts *TSMetricAggr15m) AllByMetricIDHostAndRange(tx *sqlx.Tx, clusterID, metricID int64, host string, from, to, deletedFrom int64) ([]*TSMetricAggr15mRow, error) {
	rows := []*TSMetricAggr15mRow{}
	query := fmt.Sprintf(`SELECT * FROM %v WHERE cluster_id=$1 AND metric_id=$2 AND host=$3 AND
created >= to_timestamp($4) at time zone 'utc' AND
created <= to_timestamp($5) at time zone 'utc' AND
deleted >= to_timestamp($6) at time zone 'utc' AND
host <> ''
ORDER BY cluster_id,metric_id,created ASC`, ts.table)

	err := ts.db.Select(&rows, query, clusterID, metricID, host, from, to, deletedFrom)

	return rows, err
}

func (ts *TSMetricAggr15m) AllByMetricIDAndRange(tx *sqlx.Tx, clusterID, metricID, from, to, deletedFrom int64) ([]*TSMetricAggr15mRow, error) {
	rows := []*TSMetricAggr15mRow{}
	query := fmt.Sprintf(`SELECT * FROM %v WHERE cluster_id=$1 AND metric_id=$2 AND
created >= to_timestamp($3) at time zone 'utc' AND
created <= to_timestamp($4) at time zone 'utc' AND
deleted >= to_timestamp($5) at time zone 'utc' AND
host <> ''
ORDER BY cluster_id,metric_id,created ASC`, ts.table)

	err := ts.db.Select(&rows, query, clusterID, metricID, from, to, deletedFrom)

	return rows, err
}

func (ts *TSMetricAggr15m) TransformForHighchart(tx *sqlx.Tx, tsMetricRows []*TSMetricAggr15mRow, aggr string) ([]*TSMetricHighchartPayload, error) {
	// Group all TSMetricAggr15mRows per host
	mapHostsAndMetrics := make(map[string][]*TSMetricAggr15mRow)

	for _, tsMetricRow := range tsMetricRows {
		host := tsMetricRow.Host.String

		if _, ok := mapHostsAndMetrics[host]; !ok {
			mapHostsAndMetrics[host] = make([]*TSMetricAggr15mRow, 0)
		}

		mapHostsAndMetrics[host] = append(mapHostsAndMetrics[host], tsMetricRow)
	}

	// Then generate multiple Highchart payloads per all these hosts.
	highChartPayloads := make([]*TSMetricHighchartPayload, 0)

	for host, tsMetricRows := range mapHostsAndMetrics {
		highChartPayload, err := ts.metricRowsForHighchart(tx, host, tsMetricRows, aggr)
		if err != nil {
			return nil, err
		}
		highChartPayloads = append(highChartPayloads, highChartPayload)
	}

	return highChartPayloads, nil
}

func (ts *TSMetricAggr15m) AllByMetricIDHostAndRangeForHighchart(tx *sqlx.Tx, clusterID, metricID int64, host string, from, to, deletedFrom int64, aggr string) ([]*TSMetricHighchartPayload, error) {
	tsMetricRows, err := ts.AllByMetricIDHostAndRange(tx, clusterID, metricID, host, from, to, deletedFrom)
	if err != nil {
		return nil, err
	}

	return ts.TransformForHighchart(tx, tsMetricRows, aggr)
}

func (ts *TSMetricAggr15m) AllByMetricIDAndRangeForHighchart(tx *sqlx.Tx, clusterID, metricID, from, to, deletedFrom int64, aggr string) ([]*TSMetricHighchartPayload, error) {
	if aggr == "" {
		aggr = "avg"
	}
	tsMetricRows, err := ts.AllByMetricIDAndRange(tx, clusterID, metricID, from, to, deletedFrom)
	if err != nil {
		return nil, err
	}

	return ts.TransformForHighchart(tx, tsMetricRows, aggr)
}
