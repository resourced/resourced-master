package dal

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"strings"
)

func NewSavedQuery(db *sqlx.DB) *SavedQuery {
	savedQuery := &SavedQuery{}
	savedQuery.db = db
	savedQuery.table = "saved_queries"
	savedQuery.hasID = true

	return savedQuery
}

type SavedQueryRow struct {
	ID     int64  `db:"id"`
	UserID int64  `db:"user_id"`
	Query  string `db:"query"`
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

// AllByAccessTokenID returns all saved_query rows.
func (sq *SavedQuery) AllByAccessTokenID(tx *sqlx.Tx, accessTokenID int64) ([]*SavedQueryRow, error) {
	accessTokenRow, err := NewAccessToken(sq.db).GetByID(tx, accessTokenID)
	if err != nil {
		return nil, err
	}

	return sq.AllByAccessToken(tx, accessTokenRow)
}

// AllByAccessToken returns all saved_query rows.
func (sq *SavedQuery) AllByAccessToken(tx *sqlx.Tx, accessTokenRow *AccessTokenRow) ([]*SavedQueryRow, error) {
	savedQueries := []*SavedQueryRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE user_id=$1", sq.table)
	err := sq.db.Select(&savedQueries, query, accessTokenRow.UserID)

	return savedQueries, err
}

// GetByID returns record by id.
func (sq *SavedQuery) GetByID(tx *sqlx.Tx, id int64) (*SavedQueryRow, error) {
	savedQueryRow := &SavedQueryRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=$1", sq.table)
	err := sq.db.Get(savedQueryRow, query, id)

	return savedQueryRow, err
}

// GetByAccessTokenAndQuery returns record by savedQuery.
func (sq *SavedQuery) GetByAccessTokenAndQuery(tx *sqlx.Tx, accessTokenRow *AccessTokenRow, savedQuery string) (*SavedQueryRow, error) {
	savedQueryRow := &SavedQueryRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE user_id=$1 AND query=$2", sq.table)
	err := sq.db.Get(savedQueryRow, query, accessTokenRow.UserID, savedQuery)

	return savedQueryRow, err
}

// CreateOrUpdate performs insert/update for one savedQuery data.
func (sq *SavedQuery) CreateOrUpdate(tx *sqlx.Tx, accessTokenID int64, savedQuery string) (*SavedQueryRow, error) {
	accessTokenRow, err := NewAccessToken(sq.db).GetByID(tx, accessTokenID)
	if err != nil {
		return nil, err
	}

	savedQueryRow, err := sq.GetByAccessTokenAndQuery(tx, accessTokenRow, savedQuery)

	data := make(map[string]interface{})
	data["user_id"] = accessTokenRow.UserID
	data["query"] = savedQuery

	// Perform INSERT
	if savedQueryRow == nil || err != nil {
		sqlResult, err := sq.InsertIntoTable(tx, data)
		if err != nil {
			return nil, err
		}

		return sq.savedQueryRowFromSqlResult(tx, sqlResult)
	}

	// Perform UPDATE
	_, err = sq.UpdateByAccessTokenAndSavedQuery(tx, data, accessTokenRow, savedQuery)
	if err != nil {
		return nil, err
	}

	return savedQueryRow, nil
}

func (b *Base) UpdateByAccessTokenAndSavedQuery(tx *sqlx.Tx, data map[string]interface{}, accessTokenRow *AccessTokenRow, savedQuery string) (sql.Result, error) {
	var result sql.Result

	if b.table == "" {
		return nil, errors.New("Table must not be empty.")
	}

	tx, wrapInSingleTransaction, err := b.newTransactionIfNeeded(tx)
	if tx == nil {
		return nil, errors.New("Transaction struct must not be empty.")
	}
	if err != nil {
		return nil, err
	}

	keysWithDollarMarks := make([]string, 0)
	values := make([]interface{}, 0)

	loopCounter := 1
	for key, value := range data {
		keysWithDollarMark := fmt.Sprintf("%v=$%v", key, loopCounter)
		keysWithDollarMarks = append(keysWithDollarMarks, keysWithDollarMark)
		values = append(values, value)

		loopCounter++
	}

	// Add userID and savedQuery as part of values
	values = append(values, accessTokenRow.UserID, savedQuery)

	query := fmt.Sprintf(
		"UPDATE %v SET %v WHERE user_id=$%v AND query=$%v",
		b.table,
		strings.Join(keysWithDollarMarks, ","),
		loopCounter,
		loopCounter+1)

	result, err = tx.Exec(query, values...)

	if err != nil {
		return nil, err
	}

	if wrapInSingleTransaction == true {
		err = tx.Commit()
	}

	return result, err
}
