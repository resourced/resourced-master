package dal

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func NewMetric(db *sqlx.DB) *Metric {
	m := &Metric{}
	m.db = db
	m.table = "metrics"
	m.hasID = true

	return m
}

type MetricRowsWithError struct {
	Metrics []*MetricRow
	Error   error
}

type MetricsMapWithError struct {
	MetricsMap map[string]int64
	Error      error
}

type MetricRow struct {
	ID        int64  `db:"id"`
	ClusterID int64  `db:"cluster_id"`
	Key       string `db:"key"`
}

type Metric struct {
	Base
}

func (d *Metric) metricRowFromSqlResult(tx *sqlx.Tx, sqlResult sql.Result) (*MetricRow, error) {
	id, err := sqlResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	return d.GetById(tx, id)
}

// GetById returns one record by id.
func (d *Metric) GetById(tx *sqlx.Tx, id int64) (*MetricRow, error) {
	row := &MetricRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=$1", d.table)
	err := d.db.Get(row, query, id)

	return row, err
}

// GetByClusterIDAndKey returns one record by cluster_id and key.
func (d *Metric) GetByClusterIDAndKey(tx *sqlx.Tx, clusterID int64, key string) (*MetricRow, error) {
	row := &MetricRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND key=$2", d.table)
	err := d.db.Get(row, query, clusterID, key)

	return row, err
}

func (d *Metric) CreateOrUpdate(tx *sqlx.Tx, clusterID int64, key string) (*MetricRow, error) {
	metricRow, err := d.GetByClusterIDAndKey(tx, clusterID, key)

	data := make(map[string]interface{})
	data["cluster_id"] = clusterID
	data["key"] = key

	// Perform INSERT
	if metricRow == nil || err != nil {
		sqlResult, err := d.InsertIntoTable(tx, data)
		if err != nil {
			return nil, err
		}

		return d.metricRowFromSqlResult(tx, sqlResult)
	}

	// Perform UPDATE
	_, err = d.UpdateFromTable(tx, data, fmt.Sprintf("cluster_id=%v AND key=%v", clusterID, key))
	if err != nil {
		return nil, err
	}

	return metricRow, nil
}

// AllByClusterID returns all rows.
func (d *Metric) AllByClusterID(tx *sqlx.Tx, clusterID int64) ([]*MetricRow, error) {
	rows := []*MetricRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1", d.table)
	err := d.db.Select(&rows, query, clusterID)

	return rows, err
}

// AllByClusterIDAsMap returns all rows.
func (d *Metric) AllByClusterIDAsMap(tx *sqlx.Tx, clusterID int64) (map[string]int64, error) {
	result := make(map[string]int64)

	rows, err := d.AllByClusterID(tx, clusterID)
	if err != nil {
		return result, err
	}

	for _, row := range rows {
		result[row.Key] = row.ID
	}

	return result, nil
}
