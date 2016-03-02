package dal

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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

func (g *GraphRow) MetricsFromJSON() []*MetricRow {
	metricRows := make([]*MetricRow, 0)
	g.Metrics.Unmarshal(&metricRows)
	return metricRows
}

func (g *GraphRow) MetricsFromJSONGroupByN(n int) [][]*MetricRow {
	container := make([][]*MetricRow, 0)

	for metricIndex, metricRow := range g.MetricsFromJSON() {
		if metricIndex%n == 0 {
			container = append(container, make([]*MetricRow, 0))
		}

		container[len(container)-1] = append(container[len(container)-1], metricRow)
	}

	return container
}

func (g *GraphRow) MetricsFromJSONGroupByThree() [][]*MetricRow {
	return g.MetricsFromJSONGroupByN(3)
}

type Graph struct {
	Base
}

func (a *Graph) rowFromSqlResult(tx *sqlx.Tx, sqlResult sql.Result) (*GraphRow, error) {
	id, err := sqlResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	return a.GetById(tx, id)
}

func (a *Graph) BuildMetricsJSONForSave(tx *sqlx.Tx, clusterID int64, idsAndKeys []string) ([]byte, error) {
	idsAndKeysLen := 0
	if idsAndKeys != nil {
		idsAndKeysLen = len(idsAndKeys)
	}

	metrics := make([]*MetricRow, idsAndKeysLen)

	if idsAndKeys != nil {
		for i, idAndKey := range idsAndKeys {
			idAndKeySlice := strings.Split(idAndKey, "-")

			idInt64, err := strconv.ParseInt(idAndKeySlice[0], 10, 64)
			if err != nil {
				return nil, err
			}

			metricRow := &MetricRow{}
			metricRow.ID = idInt64
			metricRow.Key = strings.Join(idAndKeySlice[1:], "-")
			metricRow.ClusterID = clusterID

			metrics[i] = metricRow
		}
	}

	return json.Marshal(metrics)
}

// GetById returns one record by id.
func (a *Graph) GetById(tx *sqlx.Tx, id int64) (*GraphRow, error) {
	row := &GraphRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=$1", a.table)
	err := a.db.Get(row, query, id)

	return row, err
}

func (a *Graph) Create(tx *sqlx.Tx, clusterID int64, data map[string]interface{}) (*GraphRow, error) {
	data["cluster_id"] = clusterID

	metricsJSONBytes, err := a.BuildMetricsJSONForSave(tx, clusterID, nil)
	if err != nil {
		return nil, err
	}
	data["metrics"] = metricsJSONBytes

	sqlResult, err := a.InsertIntoTable(tx, data)
	if err != nil {
		return nil, err
	}

	return a.rowFromSqlResult(tx, sqlResult)
}

// AllByClusterID returns all rows by cluster_id.
func (a *Graph) AllByClusterID(tx *sqlx.Tx, clusterID int64) ([]*GraphRow, error) {
	rows := []*GraphRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1", a.table)
	err := a.db.Select(&rows, query, clusterID)

	return rows, err
}

// AllGraphs returns all rows.
func (a *Graph) AllGraphs(tx *sqlx.Tx) ([]*GraphRow, error) {
	rows := []*GraphRow{}
	query := fmt.Sprintf("SELECT * FROM %v", a.table)
	err := a.db.Select(&rows, query)

	return rows, err
}

// DeleteMetricFromGraphs deletes a particular metric from all rows by cluster_id.
func (a *Graph) DeleteMetricFromGraphs(tx *sqlx.Tx, clusterID, metricID int64) error {
	rows, err := a.AllByClusterID(tx, clusterID)
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

		_, err = a.UpdateByID(tx, data, row.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
