package dal

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"
)

func NewCheck(db *sqlx.DB) *Check {
	g := &Check{}
	g.db = db
	g.table = "checks"
	g.hasID = true

	return g
}

type CheckRowsWithError struct {
	Checks []*CheckRow
	Error  error
}

type CheckRow struct {
	ID                    int64               `db:"id"`
	ClusterID             int64               `db:"cluster_id"`
	Name                  string              `db:"name"`
	Interval              string              `db:"interval"`
	HostsQuery            string              `db:"hosts_query"`
	HostsList             sqlx_types.JSONText `db:"hosts_list"`
	Expressions           sqlx_types.JSONText `db:"expressions"`
	Triggers              sqlx_types.JSONText `db:"triggers"`
	LastResultHosts       sqlx_types.JSONText `db:"last_result_hosts"`
	LastResultExpressions sqlx_types.JSONText `db:"last_result_expressions"`
}

type Check struct {
	Base
}

func (a *Check) rowFromSqlResult(tx *sqlx.Tx, sqlResult sql.Result) (*CheckRow, error) {
	id, err := sqlResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	return a.GetByID(tx, id)
}

// GetByID returns one record by id.
func (a *Check) GetByID(tx *sqlx.Tx, id int64) (*CheckRow, error) {
	row := &CheckRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=$1", a.table)
	err := a.db.Get(row, query, id)

	return row, err
}

func (a *Check) Create(tx *sqlx.Tx, clusterID int64, data map[string]interface{}) (*CheckRow, error) {
	data["cluster_id"] = clusterID

	sqlResult, err := a.InsertIntoTable(tx, data)
	if err != nil {
		return nil, err
	}

	return a.rowFromSqlResult(tx, sqlResult)
}

// AllByClusterID returns all rows by cluster_id.
func (a *Check) AllByClusterID(tx *sqlx.Tx, clusterID int64) ([]*CheckRow, error) {
	rows := []*CheckRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1", a.table)
	err := a.db.Select(&rows, query, clusterID)

	return rows, err
}

// AllChecks returns all rows.
func (a *Check) AllChecks(tx *sqlx.Tx) ([]*CheckRow, error) {
	rows := []*CheckRow{}
	query := fmt.Sprintf("SELECT * FROM %v", a.table)
	err := a.db.Select(&rows, query)

	return rows, err
}
