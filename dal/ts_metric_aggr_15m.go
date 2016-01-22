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
	println("Am i in Upsert?")
	query := fmt.Sprintf(`INSERT INTO %v (cluster_id,metric_id,created,key,avg,max,min,sum) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT DO UPDATE SET avg=excluded.avg,max=excluded.max,min=excluded.min,sum=excluded.sum;`, ts.table)

	created := time.Unix(int64(selectAggrRow.CreatedUnix), 0)

	_, err := ts.db.Exec(query, clusterID, metricID, created, selectAggrRow.Key, selectAggrRow.Avg, selectAggrRow.Max, selectAggrRow.Min, selectAggrRow.Sum)

	return err
}
