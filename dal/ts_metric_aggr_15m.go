package dal

import (
	"database/sql"
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
	Key       string         `db:"key"`
	Host      sql.NullString `db:"host"`
	Avg       float64        `db:"avg"`
	Max       float64        `db:"max"`
	Min       float64        `db:"min"`
	Sum       float64        `db:"sum"`
}

type TSMetricAggr15m struct {
	Base
}

// Upsert a new record.
func (ts *TSMetricAggr15m) Upsert(tx *sqlx.Tx, clusterID, metricID int64, metricKey string, selectAggrRow *TSMetricSelectAggregateRow) error {
	// Check if metricKey is correct, if not don't do anything
	if metricKey != selectAggrRow.Key {
		return nil
	}

	created := time.Unix(int64(selectAggrRow.CreatedUnix), 0)

	data := make(map[string]interface{})
	data["cluster_id"] = clusterID
	data["metric_id"] = metricID
	data["created"] = created
	data["key"] = selectAggrRow.Key
	data["avg"] = selectAggrRow.Avg
	data["max"] = selectAggrRow.Max
	data["min"] = selectAggrRow.Min
	data["sum"] = selectAggrRow.Sum

	query := fmt.Sprintf("SELECT * from %v WHERE cluster_id=$1 AND created=$2 AND key=$3 LIMIT 1", ts.table)
	aggrSelectRows := make([]*TSMetricAggr15mRow, 0)
	err := ts.db.Select(&aggrSelectRows, query, clusterID, created, selectAggrRow.Key)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Method":    "TSMetricAggr15m.Upsert",
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
				"Method":    "TSMetricAggr15m.Upsert.InsertIntoTable",
				"ClusterID": clusterID,
				"MetricID":  metricID,
				"MetricKey": metricKey,
			}).Error(err)
		}

	} else {
		query := fmt.Sprintf(`UPDATE %v SET avg=$1,max=$2,min=$3,sum=$4 WHERE cluster_id=$5 AND created=$6 AND key=$7`, ts.table)
		_, err = tx.Exec(query, selectAggrRow.Avg, selectAggrRow.Max, selectAggrRow.Min, selectAggrRow.Sum, clusterID, created, selectAggrRow.Key)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Method":    "TSMetricAggr15m.Upsert.UpdateFromTable",
				"ClusterID": clusterID,
				"MetricID":  metricID,
				"MetricKey": metricKey,
				"Query":     query,
			}).Error(err)
		}
	}

	return err
}
