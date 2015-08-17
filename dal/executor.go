package dal

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"
)

func NewExecutor(db *sqlx.DB) *Executor {
	e := &Executor{}
	e.db = db
	e.table = "executors"

	return e
}

type ExecutorRow struct {
	ClusterID int64               `db:"cluster_id" json:"-"`
	Hostname  string              `db:"hostname"`
	Data      sqlx_types.JsonText `db:"data"`
}

func (executorRow *ExecutorRow) DataString() string {
	return string(executorRow.Data)
}

type Executor struct {
	Base
}

// AllByClusterID returns all executor rows by cluster id.
func (e *Executor) AllByClusterID(tx *sqlx.Tx, clusterID int64) ([]*ExecutorRow, error) {
	executorRows := []*ExecutorRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 ORDER BY hostname ASC", e.table)
	err := e.db.Select(&executorRows, query, clusterID)

	return executorRows, err
}

// GetByClusterIDAndHostname returns record by cluster_id and hostname.
func (e *Executor) GetByClusterIDAndHostname(tx *sqlx.Tx, clusterID int64, hostname string) (*ExecutorRow, error) {
	executorRow := &ExecutorRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND hostname=$2", e.table)
	err := e.db.Get(executorRow, query, clusterID, hostname)

	return executorRow, err
}

// UpdateByClusterIDAndHostname updates record by cluster_id.
func (e *Executor) UpdateByClusterIDAndHostname(tx *sqlx.Tx, clusterID int64, hostname string, data []byte) (*ExecutorRow, error) {
	query := fmt.Sprintf("UPDATE %v SET data=$3 WHERE cluster_id=$1 AND hostname=$2", e.table)

	_, err := e.db.Exec(query, clusterID, hostname, data)
	if err != nil {
		return nil, err
	}

	return &ExecutorRow{clusterID, hostname, data}, nil
}

// CreateOrUpdate performs insert/update for one executor data.
func (e *Executor) CreateOrUpdate(tx *sqlx.Tx, clusterID int64, hostname string, data []byte) (*ExecutorRow, error) {
	executorRow, err := e.GetByClusterIDAndHostname(tx, clusterID, hostname)

	// Perform INSERT
	if executorRow == nil || err != nil {
		saveData := make(map[string]interface{})
		saveData["cluster_id"] = clusterID
		saveData["hostname"] = hostname
		saveData["data"] = data

		_, err := e.InsertIntoTable(tx, saveData)
		if err != nil {
			return nil, err
		}

		return &ExecutorRow{clusterID, hostname, data}, nil
	}

	// Perform UPDATE
	return e.UpdateByClusterIDAndHostname(tx, clusterID, hostname, data)
}
