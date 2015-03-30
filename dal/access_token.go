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

type AccessTokenRow struct {
	ID     int64  `db:"id"`
	UserID int64  `db:"user_id"`
	Token  string `db:"token"`
	Level  string `db:"level"` // read, write, execute
}

type AccessToken struct {
	Base
}

func (t *AccessToken) tokenRowFromSqlResult(tx *sqlx.Tx, sqlResult sql.Result) (*AccessTokenRow, error) {
	tokenId, err := sqlResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	return t.GetById(tx, tokenId)
}

// GetById returns record by id.
func (t *AccessToken) GetById(tx *sqlx.Tx, id int64) (*AccessTokenRow, error) {
	tokenRow := &AccessTokenRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=$1", t.table)
	err := t.db.Get(tokenRow, query, id)

	return tokenRow, err
}

// GetByAccessToken returns record by token.
func (t *AccessToken) GetByAccessToken(tx *sqlx.Tx, token string) (*AccessTokenRow, error) {
	tokenRow := &AccessTokenRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE token=$1", t.table)
	err := t.db.Get(tokenRow, query, token)

	return tokenRow, err
}

func (t *AccessToken) CreateRow(tx *sqlx.Tx, userId int64, level string) (*AccessTokenRow, error) {
	token, err := libstring.GeneratePassword(32)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	data["user_id"] = userId
	data["token"] = token
	data["level"] = level

	sqlResult, err := t.InsertIntoTable(tx, data)
	if err != nil {
		return nil, err
	}

	return t.tokenRowFromSqlResult(tx, sqlResult)
}

// AllAccessTokens returns all user rows.
func (t *AccessToken) AllAccessTokens(tx *sqlx.Tx) ([]*AccessTokenRow, error) {
	accessTokens := []*AccessTokenRow{}
	query := fmt.Sprintf("SELECT * FROM %v", t.table)
	err := t.db.Select(&accessTokens, query)

	return accessTokens, err
}
