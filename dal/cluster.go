package dal

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func NewCluster(db *sqlx.DB) *Cluster {
	application := &Cluster{}
	application.db = db
	application.table = "clusters"
	application.hasID = true

	return application
}

type ClusterRow struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

type Cluster struct {
	Base
}

func (a *Cluster) clusterRowFromSqlResult(tx *sqlx.Tx, sqlResult sql.Result) (*ClusterRow, error) {
	id, err := sqlResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	return a.GetById(tx, id)
}

// GetById returns one record by id.
func (a *Cluster) GetById(tx *sqlx.Tx, id int64) (*ClusterRow, error) {
	applicationRow := &ClusterRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=$1", a.table)
	err := a.db.Get(applicationRow, query, id)

	return applicationRow, err
}

func (a *Cluster) Create(tx *sqlx.Tx, userId int64, name string) (*ClusterRow, error) {
	data := make(map[string]interface{})
	data["name"] = name

	sqlResult, err := a.InsertIntoTable(tx, data)
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

	_, err = a.db.NamedExec(query, joinTableData)

	return a.clusterRowFromSqlResult(tx, sqlResult)
}

// AllClustersByUserID returns all clusters rows.
func (a *Cluster) AllClustersByUserID(tx *sqlx.Tx, userId int64) ([]*ClusterRow, error) {
	accessTokens := []*ClusterRow{}
	query := fmt.Sprintf("SELECT id, name FROM %v JOIN clusters_users ON %v.id = clusters_users.cluster_id WHERE user_id=$1", a.table, a.table)
	err := a.db.Select(&accessTokens, query, userId)

	return accessTokens, err
}
