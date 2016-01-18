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

type GraphRow struct {
	ID          int64               `db:"id"`
	ClusterID   int64               `db:"cluster_id"`
	Name        string              `db:"name"`
	Description string              `db:"description"`
	Metrics     sqlx_types.JSONText `db:"metrics"`
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

func (a *Graph) BuildMetricsJSONForSave(tx *sqlx.Tx, idsAndKeys []string) ([]byte, error) {
	idsAndKeysLen := 0
	if idsAndKeys != nil {
		idsAndKeysLen = len(idsAndKeys)
	}

	ids := make([]int, idsAndKeysLen)
	keys := make([]string, idsAndKeysLen)

	for i, idAndKey := range idsAndKeys {
		idAndKeySlice := strings.Split(idAndKey, "-")

		idInt, err := strconv.Atoi(idAndKeySlice[0])
		if err != nil {
			return nil, err
		}

		ids[i] = idInt
		keys[i] = idAndKeySlice[1]
	}

	metrics := make(map[string]interface{})
	metrics["ids"] = ids
	metrics["keys"] = keys

	return json.Marshal(metrics)
}

// GetById returns one record by id.
func (a *Graph) GetById(tx *sqlx.Tx, id int64) (*GraphRow, error) {
	row := &GraphRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=$1", a.table)
	err := a.db.Get(row, query, id)

	return row, err
}

func (a *Graph) Create(tx *sqlx.Tx, clusterID int64, name, description string) (*GraphRow, error) {
	data := make(map[string]interface{})
	data["cluster_id"] = clusterID
	data["name"] = name
	data["description"] = description

	metricsJSONBytes, err := a.BuildMetricsJSONForSave(tx, nil)
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
