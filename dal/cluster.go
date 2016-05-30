package dal

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"
)

func NewCluster(db *sqlx.DB) *Cluster {
	c := &Cluster{}
	c.db = db
	c.table = "clusters"
	c.hasID = true

	return c
}

type ClusterRow struct {
	ID            int64               `db:"id"`
	Name          string              `db:"name"`
	DataRetention sqlx_types.JSONText `db:"data_retention"`
}

func (cr *ClusterRow) GetDataRetention() map[string]int {
	retentions := make(map[string]int)
	cr.DataRetention.Unmarshal(&retentions)

	return retentions
}

type Cluster struct {
	Base
}

func (c *Cluster) clusterRowFromSqlResult(tx *sqlx.Tx, sqlResult sql.Result) (*ClusterRow, error) {
	id, err := sqlResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	return c.GetByID(tx, id)
}

// GetByID returns one record by id.
func (c *Cluster) GetByID(tx *sqlx.Tx, id int64) (*ClusterRow, error) {
	row := &ClusterRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=$1", c.table)
	err := c.db.Get(row, query, id)

	return row, err
}

func (c *Cluster) Create(tx *sqlx.Tx, userId int64, name string) (*ClusterRow, error) {
	dataRetention := make(map[string]int)
	dataRetention["ts_checks"] = 1
	dataRetention["ts_metrics"] = 1
	dataRetention["ts_metrics_aggr_15m"] = 1
	dataRetention["ts_events"] = 1
	dataRetention["ts_executor_logs"] = 1
	dataRetention["ts_logs"] = 1

	dataRetentionJSON, err := json.Marshal(dataRetention)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	data["name"] = name
	data["data_retention"] = dataRetentionJSON

	sqlResult, err := c.InsertIntoTable(tx, data)
	if err != nil {
		return nil, err
	}

	clusterId, err := sqlResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	joinTableData := make(map[string]interface{})
	joinTableData["cluster_id"] = clusterId
	joinTableData["user_id"] = userId

	query := `INSERT INTO clusters_users (cluster_id,user_id) VALUES (:cluster_id,:user_id)`

	_, err = c.db.NamedExec(query, joinTableData)

	return c.clusterRowFromSqlResult(tx, sqlResult)
}

// AllByUserID returns all clusters rows by user ID.
func (c *Cluster) AllByUserID(tx *sqlx.Tx, userId int64) ([]*ClusterRow, error) {
	rows := []*ClusterRow{}
	query := fmt.Sprintf("SELECT id, name, data_retention FROM %v JOIN clusters_users ON %v.id = clusters_users.cluster_id WHERE user_id=$1", c.table, c.table)
	err := c.db.Select(&rows, query, userId)

	return rows, err
}

// All returns all clusters rows.
func (c *Cluster) All(tx *sqlx.Tx) ([]*ClusterRow, error) {
	rows := []*ClusterRow{}
	query := fmt.Sprintf("SELECT * FROM %v", c.table)
	err := c.db.Select(&rows, query)

	return rows, err
}
