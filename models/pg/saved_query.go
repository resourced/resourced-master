package pg

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func NewSavedQuery(ctx context.Context) *SavedQuery {
	savedQuery := &SavedQuery{}
	savedQuery.AppContext = ctx
	savedQuery.table = "saved_queries"
	savedQuery.hasID = true
	savedQuery.i = savedQuery

	return savedQuery
}

type SavedQueryRowsWithError struct {
	SavedQueries []*SavedQueryRow
	Error        error
}

type SavedQueryRow struct {
	ID        int64  `db:"id"`
	UserID    int64  `db:"user_id"`
	ClusterID int64  `db:"cluster_id"`
	Type      string `db:"type"`
	Query     string `db:"query"`
}

type SavedQuery struct {
	Base
}

func (sq *SavedQuery) savedQueryRowFromSqlResult(tx *sqlx.Tx, sqlResult sql.Result) (*SavedQueryRow, error) {
	savedQueryId, err := sqlResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	return sq.GetByID(tx, savedQueryId)
}

// AllByClusterID returns all saved_query rows.
func (sq *SavedQuery) AllByClusterID(tx *sqlx.Tx, clusterID int64) ([]*SavedQueryRow, error) {
	pgdb, err := sq.GetPGDB()
	if err != nil {
		return nil, err
	}

	savedQueries := []*SavedQueryRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1", sq.table)
	err = pgdb.Select(&savedQueries, query, clusterID)

	return savedQueries, err
}

// AllByClusterIDAndType returns all saved_query rows.
func (sq *SavedQuery) AllByClusterIDAndType(tx *sqlx.Tx, clusterID int64, savedQueryType string) ([]*SavedQueryRow, error) {
	pgdb, err := sq.GetPGDB()
	if err != nil {
		return nil, err
	}

	savedQueries := []*SavedQueryRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND type=$2", sq.table)
	err = pgdb.Select(&savedQueries, query, clusterID, savedQueryType)

	return savedQueries, err
}

// GetByID returns record by id.
func (sq *SavedQuery) GetByID(tx *sqlx.Tx, id int64) (*SavedQueryRow, error) {
	pgdb, err := sq.GetPGDB()
	if err != nil {
		return nil, err
	}

	savedQueryRow := &SavedQueryRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=$1", sq.table)
	err = pgdb.Get(savedQueryRow, query, id)

	return savedQueryRow, err
}

// GetByAccessTokenAndQuery returns record by savedQuery.
func (sq *SavedQuery) GetByAccessTokenAndQuery(tx *sqlx.Tx, accessTokenRow *AccessTokenRow, savedQueryType, savedQuery string) (*SavedQueryRow, error) {
	pgdb, err := sq.GetPGDB()
	if err != nil {
		return nil, err
	}

	savedQueryRow := &SavedQueryRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND type=$2 AND query=$3", sq.table)
	err = pgdb.Get(savedQueryRow, query, accessTokenRow.ClusterID, savedQueryType, savedQuery)

	return savedQueryRow, err
}

// CreateOrUpdate performs insert/update for one savedQuery data.
func (sq *SavedQuery) CreateOrUpdate(tx *sqlx.Tx, accessTokenRow *AccessTokenRow, savedQueryType, savedQuery string) (*SavedQueryRow, error) {
	savedQueryRow, err := sq.GetByAccessTokenAndQuery(tx, accessTokenRow, savedQueryType, savedQuery)

	data := make(map[string]interface{})
	data["user_id"] = accessTokenRow.UserID
	data["cluster_id"] = accessTokenRow.ClusterID
	data["type"] = savedQueryType
	data["query"] = savedQuery

	// Perform INSERT
	if savedQueryRow == nil || err != nil {
		sqlResult, err := sq.InsertIntoTable(tx, data)
		if err != nil {
			return nil, err
		}

		return sq.savedQueryRowFromSqlResult(tx, sqlResult)
	}

	return savedQueryRow, nil
}
