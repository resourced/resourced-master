package cassandra

import (
	"context"
	"fmt"

	"github.com/Sirupsen/logrus"
)

func NewSavedQuery(ctx context.Context) *SavedQuery {
	savedQuery := &SavedQuery{}
	savedQuery.AppContext = ctx
	savedQuery.table = "saved_queries"

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

// GetByID returns record by id.
func (sq *SavedQuery) GetByID(id int64) (*SavedQueryRow, error) {
	session, err := sq.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT id, name, creator_id, creator_email, data_retention, members FROM %v WHERE id=?", sq.table)

	var scannedID, scannedUserID, scannedClusterID int64
	var scannedType, scannedQuery string

	err = session.Query(query, id).Scan(&scannedID, &scannedUserID, &scannedClusterID, &scannedType, &scannedQuery)
	if err != nil {
		return nil, err
	}

	row := &SavedQueryRow{
		ID:        scannedID,
		UserID:    scannedUserID,
		ClusterID: scannedClusterID,
		Type:      scannedType,
		Query:     scannedQuery,
	}

	return row, err
}

// AllByClusterID returns all saved_query rows.
func (sq *SavedQuery) AllByClusterID(clusterID int64) ([]*SavedQueryRow, error) {
	session, err := sq.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	rows := []*SavedQueryRow{}

	query := fmt.Sprintf(`SELECT id, user_id, cluster_id, type, query FROM %v WHERE cluster_id=? ALLOW FILTERING`, sq.table)

	var scannedID, scannedUserID, scannedClusterID int64
	var scannedType, scannedQuery string

	iter := session.Query(query, clusterID).Iter()
	for iter.Scan(&scannedID, &scannedUserID, &scannedClusterID, &scannedType, &scannedQuery) {
		rows = append(rows, &SavedQueryRow{
			ID:        scannedID,
			UserID:    scannedUserID,
			ClusterID: scannedClusterID,
			Type:      scannedType,
			Query:     scannedQuery,
		})
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{"Method": "SavedQuery.AllByClusterID"}).Error(err)

		return nil, err
	}
	return rows, err
}

// AllByClusterIDAndType returns all saved_query rows.
func (sq *SavedQuery) AllByClusterIDAndType(clusterID int64, savedQueryType string) ([]*SavedQueryRow, error) {
	session, err := sq.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	rows := []*SavedQueryRow{}

	query := fmt.Sprintf(`SELECT id, user_id, cluster_id, type, query FROM %v WHERE cluster_id=? AND type=? ALLOW FILTERING`, sq.table)

	var scannedID, scannedUserID, scannedClusterID int64
	var scannedType, scannedQuery string

	iter := session.Query(query, clusterID, savedQueryType).Iter()
	for iter.Scan(&scannedID, &scannedUserID, &scannedClusterID, &scannedType, &scannedQuery) {
		rows = append(rows, &SavedQueryRow{
			ID:        scannedID,
			UserID:    scannedUserID,
			ClusterID: scannedClusterID,
			Type:      scannedType,
			Query:     scannedQuery,
		})
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{"Method": "SavedQuery.AllByClusterIDAndType"}).Error(err)

		return nil, err
	}
	return rows, err
}

// GetByAccessTokenAndQuery returns record by clusterID, savedQueryType, savedQuery.
func (sq *SavedQuery) GetByAccessTokenAndQuery(clusterID int64, savedQueryType, savedQuery string) (*SavedQueryRow, error) {
	session, err := sq.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT id, user_id, cluster_id, type, query FROM %v WHERE cluster_id=? AND type=? AND query=? ALLOW FILTERING", sq.table)

	var scannedID, scannedUserID, scannedClusterID int64
	var scannedType, scannedQuery string

	err = session.Query(query, clusterID, savedQueryType, savedQuery).Scan(&scannedID, &scannedUserID, &scannedClusterID, &scannedType, &scannedQuery)
	if err != nil {
		return nil, err
	}

	row := &SavedQueryRow{
		ID:        scannedID,
		UserID:    scannedUserID,
		ClusterID: scannedClusterID,
		Type:      scannedType,
		Query:     scannedQuery,
	}

	return row, err
}

// CreateOrUpdate performs insert/update for one savedQuery data.
func (sq *SavedQuery) CreateOrUpdate(accessTokenRow *AccessTokenRow, savedQueryType, savedQuery string) (*SavedQueryRow, error) {
	savedQueryRow, err := sq.GetByAccessTokenAndQuery(accessTokenRow.ClusterID, savedQueryType, savedQuery)
	if err != nil && err.Error() != "not found" {
		return nil, err
	}

	session, err := sq.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	var id int64

	if savedQueryRow == nil {
		id = NewExplicitID()
	} else {
		id = savedQueryRow.ID
	}

	query := fmt.Sprintf("INSERT INTO %v (id, user_id, cluster_id, type, query) VALUES (?, ?, ?, ?, ?)", sq.table)

	err = session.Query(query, id, accessTokenRow.UserID, accessTokenRow.ClusterID, savedQueryType, savedQuery).Exec()
	if err != nil {
		return nil, err
	}

	return &SavedQueryRow{
		ID:        id,
		UserID:    accessTokenRow.UserID,
		ClusterID: accessTokenRow.ClusterID,
		Type:      savedQueryType,
		Query:     savedQuery,
	}, nil
}
