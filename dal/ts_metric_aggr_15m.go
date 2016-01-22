package dal

import (
	"database/sql"
	"fmt"
	"time"

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
func (ts *TSMetricAggr15m) Upsert(tx *sqlx.Tx, clusterID, metricID int64, selectAggrRow *TSMetricSelectAggregateRow) error {
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

	something := make([]*TSMetricAggr15mRow, 0)
	err := ts.db.Select(&something, fmt.Sprintf("SELECT * from %v WHERE cluster_id=$1 AND metric_id=$2 AND created=$3 AND key=$4 LIMIT 1", ts.table), clusterID, metricID, created, selectAggrRow.Key)

	if err != nil {
		println(err.Error())
	}

	if err != nil || len(something) == 0 {
		_, err = ts.InsertIntoTable(tx, data)
	} else {
		_, err = ts.UpdateFromTable(tx, data, fmt.Sprintf("cluster_id=%v AND metric_id=%v AND created=%v AND key=%v", clusterID, metricID, created, selectAggrRow.Key))
	}

	return err
}
