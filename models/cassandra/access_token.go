package cassandra

import (
	"context"
	"errors"
	"fmt"

	"github.com/Sirupsen/logrus"

	"github.com/resourced/resourced-master/libstring"
)

func NewAccessToken(ctx context.Context) *AccessToken {
	token := &AccessToken{}
	token.AppContext = ctx
	token.table = "access_tokens"

	return token
}

type AccessTokenRow struct {
	ID        int64  `db:"id"`
	UserID    int64  `db:"user_id"`
	ClusterID int64  `db:"cluster_id"`
	Token     string `db:"token"`
	Level     string `db:"level"` // read, write, execute
	Enabled   bool   `db:"enabled"`
}

type AccessToken struct {
	Base
}

// GetByID returns one record by id.
func (t *AccessToken) GetByID(id int64) (*AccessTokenRow, error) {
	session, err := t.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT id, user_id, cluster_id, token_, level, enabled FROM %v WHERE id=? LIMIT 1", t.table)

	var scannedID, scannedUserID, scannedClusterID int64
	var scannedToken, scannedLevel string
	var scannedEnabled bool

	err = session.Query(query, id).Scan(&scannedID, &scannedUserID, &scannedClusterID, &scannedToken, &scannedLevel, &scannedEnabled)
	if err != nil {
		return nil, err
	}

	row := &AccessTokenRow{
		ID:        scannedID,
		UserID:    scannedUserID,
		ClusterID: scannedClusterID,
		Token:     scannedToken,
		Level:     scannedLevel,
		Enabled:   scannedEnabled,
	}

	return row, nil
}

// GetByAccessToken returns one record by token.
func (t *AccessToken) GetByAccessToken(token string) (*AccessTokenRow, error) {
	session, err := t.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	// id bigint primary key,
	// user_id bigint,
	// cluster_id bigint,
	// token_ text,
	// level text,
	// enabled boolean

	query := fmt.Sprintf("SELECT id, user_id, cluster_id, token_, level, enabled FROM %v WHERE token_=? LIMIT 1", t.table)

	var scannedID, scannedUserID, scannedClusterID int64
	var scannedToken, scannedLevel string
	var scannedEnabled bool

	err = session.Query(query, token).Scan(&scannedID, &scannedUserID, &scannedClusterID, &scannedToken, &scannedLevel, &scannedEnabled)
	if err != nil {
		return nil, err
	}

	row := &AccessTokenRow{
		ID:        scannedID,
		UserID:    scannedUserID,
		ClusterID: scannedClusterID,
		Token:     scannedToken,
		Level:     scannedLevel,
		Enabled:   scannedEnabled,
	}

	return row, nil
}

// GetByUserID returns one record by user_id.
func (t *AccessToken) GetByUserID(userID int64) (*AccessTokenRow, error) {
	session, err := t.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT id, user_id, cluster_id, token_, level, enabled FROM %v WHERE user_id=? LIMIT 1", t.table)

	var scannedID, scannedUserID, scannedClusterID int64
	var scannedToken, scannedLevel string
	var scannedEnabled bool

	err = session.Query(query, userID).Scan(&scannedID, &scannedUserID, &scannedClusterID, &scannedToken, &scannedLevel, &scannedEnabled)
	if err != nil {
		return nil, err
	}

	row := &AccessTokenRow{
		ID:        scannedID,
		UserID:    scannedUserID,
		ClusterID: scannedClusterID,
		Token:     scannedToken,
		Level:     scannedLevel,
		Enabled:   scannedEnabled,
	}

	return row, nil
}

// GetByClusterID returns one record by cluster_id.
func (t *AccessToken) GetByClusterID(clusterID int64) (*AccessTokenRow, error) {
	session, err := t.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT id, user_id, cluster_id, token_, level, enabled FROM %v WHERE user_id=? LIMIT 1", t.table)

	var scannedID, scannedUserID, scannedClusterID int64
	var scannedToken, scannedLevel string
	var scannedEnabled bool

	err = session.Query(query, clusterID).Scan(&scannedID, &scannedUserID, &scannedClusterID, &scannedToken, &scannedLevel, &scannedEnabled)
	if err != nil {
		return nil, err
	}

	row := &AccessTokenRow{
		ID:        scannedID,
		UserID:    scannedUserID,
		ClusterID: scannedClusterID,
		Token:     scannedToken,
		Level:     scannedLevel,
		Enabled:   scannedEnabled,
	}

	return row, nil
}

func (t *AccessToken) Create(userID, clusterID int64, level string) (*AccessTokenRow, error) {
	token, err := libstring.GeneratePassword(32)
	if err != nil {
		return nil, err
	}

	session, err := t.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	id := NewExplicitID()

	query := fmt.Sprintf("INSERT INTO %v (id, user_id, cluster_id, token_, level, enabled) VALUES (?, ?, ?, ?, ?, ?)", t.table)

	err = session.Query(query, id, userID, clusterID, token, level, true).Exec()
	if err != nil {
		return nil, err
	}

	return &AccessTokenRow{
		ID:        id,
		UserID:    userID,
		ClusterID: clusterID,
		Token:     token,
		Level:     level,
		Enabled:   true,
	}, nil
}

// AllByClusterID returns all access tokens by cluster id.
func (t *AccessToken) AllByClusterID(clusterID int64) ([]*AccessTokenRow, error) {
	session, err := t.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	rows := []*AccessTokenRow{}

	query := fmt.Sprintf(`SELECT id, user_id, cluster_id, token_, level, enabled FROM %v WHERE cluster_id=? ALLOW FILTERING`, t.table)

	var scannedID, scannedUserID, scannedClusterID int64
	var scannedToken, scannedLevel string
	var scannedEnabled bool

	iter := session.Query(query, clusterID).Iter()
	for iter.Scan(&scannedID, &scannedUserID, &scannedClusterID, &scannedToken, &scannedLevel, &scannedEnabled) {
		rows = append(rows, &AccessTokenRow{
			ID:        scannedID,
			UserID:    scannedUserID,
			ClusterID: scannedClusterID,
			Token:     scannedToken,
			Level:     scannedLevel,
			Enabled:   scannedEnabled,
		})
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{"Method": "AccessToken.AllByClusterID"}).Error(err)

		return nil, err
	}
	return rows, err
}

// UpdateLevelByID updates level by id.
func (t *AccessToken) UpdateLevelByID(id int64, level string) error {
	session, err := t.GetCassandraSession()
	if err != nil {
		return err
	}

	if level == "" {
		return errors.New("Level cannot be empty")
	}

	query := fmt.Sprintf("UPDATE %v SET level=? WHERE id=?", t.table)

	return session.Query(query, level).Exec()
}

// UpdateEnabledByID updates level by id.
func (t *AccessToken) UpdateEnabledByID(id int64, enabled bool) error {
	session, err := t.GetCassandraSession()
	if err != nil {
		return err
	}

	query := fmt.Sprintf("UPDATE %v SET enabled=? WHERE id=?", t.table)

	return session.Query(query, enabled).Exec()
}
