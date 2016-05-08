package dal

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"
)

func NewGraph(db *sqlx.DB) *Graph {
	g := &Graph{}
	g.db = db
	g.table = "graphs"
	g.hasID = true

	return g
}

type GraphRowsWithError struct {
	Graphs []*GraphRow
	Error  error
}

type GraphRow struct {
	ID          int64               `db:"id"`
	ClusterID   int64               `db:"cluster_id"`
	Name        string              `db:"name"`
	Description string              `db:"description"`
	Range       string              `db:"range"`
	Metrics     sqlx_types.JSONText `db:"metrics"`
}

func (gr *GraphRow) MetricsFromJSON() []*MetricRow {
	metricRows := make([]*MetricRow, 0)
	gr.Metrics.Unmarshal(&metricRows)
	return metricRows
}

func (gr *GraphRow) MetricsFromJSONGroupByN(n int) [][]*MetricRow {
	container := make([][]*MetricRow, 0)

	for metricIndex, metricRow := range gr.MetricsFromJSON() {
		if metricIndex%n == 0 {
			container = append(container, make([]*MetricRow, 0))
		}

		container[len(container)-1] = append(container[len(container)-1], metricRow)
	}

	return container
}

func (gr *GraphRow) MetricsFromJSONGroupByThree() [][]*MetricRow {
	return gr.MetricsFromJSONGroupByN(3)
}

type Graph struct {
	Base
}

func (g *Graph) rowFromSqlResult(tx *sqlx.Tx, sqlResult sql.Result) (*GraphRow, error) {
	id, err := sqlResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	return g.GetByID(tx, id)
}

// GetByClusterIDAndID returns one record by id.
func (g *Graph) GetByClusterIDAndID(tx *sqlx.Tx, clusterID, id int64) (*GraphRow, error) {
	row := &GraphRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND id=$2", g.table)
	err := g.db.Get(row, query, clusterID, id)

	return row, err
}

// GetByID returns one record by id.
func (g *Graph) GetByID(tx *sqlx.Tx, id int64) (*GraphRow, error) {
	row := &GraphRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=$1", g.table)
	err := g.db.Get(row, query, id)

	return row, err
}

func (g *Graph) Create(tx *sqlx.Tx, clusterID int64, data map[string]interface{}) (*GraphRow, error) {
	data["cluster_id"] = clusterID
	data["metrics"] = []byte("[]")

	sqlResult, err := g.InsertIntoTable(tx, data)
	if err != nil {
		return nil, err
	}

	return g.rowFromSqlResult(tx, sqlResult)
}

// AllByClusterID returns all rows by cluster_id.
func (g *Graph) AllByClusterID(tx *sqlx.Tx, clusterID int64) ([]*GraphRow, error) {
	rows := []*GraphRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1", g.table)
	err := g.db.Select(&rows, query, clusterID)

	return rows, err
}

// AllGraphs returns all rows.
func (g *Graph) AllGraphs(tx *sqlx.Tx) ([]*GraphRow, error) {
	rows := []*GraphRow{}
	query := fmt.Sprintf("SELECT * FROM %v", g.table)
	err := g.db.Select(&rows, query)

	return rows, err
}

// DeleteMetricFromGraphs deletes a particular metric from all rows by cluster_id.
func (g *Graph) DeleteMetricFromGraphs(tx *sqlx.Tx, clusterID, metricID int64) error {
	rows, err := g.AllByClusterID(tx, clusterID)
	if err != nil {
		return err
	}

	for _, row := range rows {
		metrics := row.MetricsFromJSON()
		newMetrics := make([]*MetricRow, 0)

		for _, metric := range metrics {
			if metric.ID != metricID {
				newMetrics = append(newMetrics, metric)
			}
		}

		newMetricsJSONBytes, err := json.Marshal(newMetrics)
		if err != nil {
			return err
		}

		data := make(map[string]interface{})
		data["metrics"] = newMetricsJSONBytes

		_, err = g.UpdateByID(tx, data, row.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateMetricsByClusterIDAndID updates metrics JSON and then returns record by id.
func (g *Graph) UpdateMetricsByClusterIDAndID(tx *sqlx.Tx, clusterID, id int64, metricsJSON []byte) (*GraphRow, error) {
	row, err := g.GetByClusterIDAndID(tx, clusterID, id)
	if err != nil {
		println(err.Error())
		return nil, err
	}

	data := make(map[string]interface{})
	data["metrics"] = metricsJSON

	_, err = g.UpdateByID(tx, data, row.ID)
	if err != nil {
		println(err.Error())
		return nil, err
	}

	return g.GetByClusterIDAndID(tx, clusterID, id)
}
