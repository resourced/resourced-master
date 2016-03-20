package dal

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/resourced/resourced-master/libstring"
)

func NewAccessToken(db *sqlx.DB) *AccessToken {
	token := &AccessToken{}
	token.db = db
	token.table = "access_tokens"
	token.hasID = true

	return token
}

type AccessTokenRowWithError struct {
	AccessToken *AccessTokenRow
	Error       error
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

func (t *AccessToken) tokenRowFromSqlResult(tx *sqlx.Tx, sqlResult sql.Result) (*AccessTokenRow, error) {
	tokenId, err := sqlResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	return t.GetByID(tx, tokenId)
}

// GetByID returns one record by id.
func (t *AccessToken) GetByID(tx *sqlx.Tx, id int64) (*AccessTokenRow, error) {
	tokenRow := &AccessTokenRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=$1", t.table)
	err := t.db.Get(tokenRow, query, id)

	return tokenRow, err
}

// GetByAccessToken returns one record by token.
func (t *AccessToken) GetByAccessToken(tx *sqlx.Tx, token string) (*AccessTokenRow, error) {
	tokenRow := &AccessTokenRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE token=$1", t.table)
	err := t.db.Get(tokenRow, query, token)

	return tokenRow, err
}

// GetByUserID returns one record by user_id.
func (t *AccessToken) GetByUserID(tx *sqlx.Tx, userID int64) (*AccessTokenRow, error) {
	tokenRow := &AccessTokenRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE user_id=$1", t.table)
	err := t.db.Get(tokenRow, query, userID)

	return tokenRow, err
}

// GetByClusterID returns one record by cluster_id.
func (t *AccessToken) GetByClusterID(tx *sqlx.Tx, clusterID int64) (*AccessTokenRow, error) {
	tokenRow := &AccessTokenRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1", t.table)
	err := t.db.Get(tokenRow, query, clusterID)

	return tokenRow, err
}

func (t *AccessToken) Create(tx *sqlx.Tx, userID, clusterID int64, level string) (*AccessTokenRow, error) {
	token, err := libstring.GeneratePassword(32)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	data["user_id"] = userID
	data["cluster_id"] = clusterID
	data["token"] = token
	data["level"] = level
	data["enabled"] = true

	sqlResult, err := t.InsertIntoTable(tx, data)
	if err != nil {
		return nil, err
	}

	return t.tokenRowFromSqlResult(tx, sqlResult)
}

// AllAccessTokens returns all access tokens.
func (t *AccessToken) AllAccessTokens(tx *sqlx.Tx) ([]*AccessTokenRow, error) {
	accessTokens := []*AccessTokenRow{}
	query := fmt.Sprintf("SELECT * FROM %v", t.table)
	err := t.db.Select(&accessTokens, query)

	return accessTokens, err
}

// AllAccessTokens returns all access tokens by cluster id.
func (t *AccessToken) AllAccessTokensByClusterID(tx *sqlx.Tx, clusterID int64) ([]*AccessTokenRow, error) {
	accessTokens := []*AccessTokenRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1", t.table)
	err := t.db.Select(&accessTokens, query, clusterID)

	return accessTokens, err
}
