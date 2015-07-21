package dal

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"strings"
)

func NewTask(db *sqlx.DB) *Task {
	taskCron := &Task{}
	taskCron.db = db
	taskCron.table = "tasks"
	taskCron.hasID = true

	return taskCron
}

type TaskRow struct {
	ID     int64  `db:"id"`
	UserID int64  `db:"user_id"`
	Query  string `db:"query"`
	Cron   string `db:"cron"`
}

type Task struct {
	Base
}

func (task *Task) taskRowFromSqlResult(tx *sqlx.Tx, sqlResult sql.Result) (*TaskRow, error) {
	taskID, err := sqlResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	return task.GetByID(tx, taskID)
}

// DeleteByID deletes record by id.
func (task *Task) DeleteByID(tx *sqlx.Tx, id int64) error {
	query := fmt.Sprintf("DELETE FROM %v WHERE id=$1", task.table)
	_, err := task.db.Exec(query, id)

	return err
}

// AllByAccessToken returns all tasks_cron rows.
func (task *Task) AllByAccessToken(tx *sqlx.Tx, accessTokenRow *AccessTokenRow) ([]*TaskRow, error) {
	taskRows := []*TaskRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE user_id=$1", task.table)
	err := task.db.Select(&taskRows, query, accessTokenRow.UserID)

	return taskRows, err
}

// GetByID returns record by id.
func (task *Task) GetByID(tx *sqlx.Tx, id int64) (*TaskRow, error) {
	taskRow := &TaskRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=$1", task.table)
	err := task.db.Get(taskRow, query, id)

	return taskRow, err
}

// GetByAccessTokenQueryAndCron returns record by token, query and cron.
func (task *Task) GetByAccessTokenQueryAndCron(tx *sqlx.Tx, accessTokenRow *AccessTokenRow, savedQuery, cron string) (*TaskRow, error) {
	taskRow := &TaskRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE user_id=$1 AND query=$2 AND cron=$3", task.table)
	err := task.db.Get(taskRow, query, accessTokenRow.UserID, savedQuery, cron)

	return taskRow, err
}

// Create performs insert for one task data.
func (task *Task) Create(tx *sqlx.Tx, accessTokenID int64, savedQuery, cron string) (*TaskRow, error) {
	accessTokenRow, err := NewAccessToken(task.db).GetByID(tx, accessTokenID)
	if err != nil {
		return nil, err
	}

	taskRow, err := task.GetByAccessTokenQueryAndCron(tx, accessTokenRow, savedQuery, cron)

	data := make(map[string]interface{})
	data["user_id"] = accessTokenRow.UserID
	data["query"] = savedQuery
	data["cron"] = cron

	// Perform INSERT
	if taskRow == nil || err != nil {
		sqlResult, err := task.InsertIntoTable(tx, data)
		if err != nil {
			return nil, err
		}

		return task.taskRowFromSqlResult(tx, sqlResult)
	}

	return taskRow, nil
}

func (b *Base) UpdateByID(tx *sqlx.Tx, data map[string]interface{}, id int64) (sql.Result, error) {
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

	// Add id as part of values
	values = append(values, id)

	query := fmt.Sprintf(
		"UPDATE %v SET %v WHERE id=$%v",
		b.table,
		strings.Join(keysWithDollarMarks, ","),
		loopCounter)

	result, err = tx.Exec(query, values...)

	if err != nil {
		return nil, err
	}

	if wrapInSingleTransaction == true {
		err = tx.Commit()
	}

	return result, err
}
