package dal

import (
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type TSBase struct {
	Base
}

// DeleteByDayInterval deletes all record older than x days ago.
func (ts *TSBase) DeleteByDayInterval(tx *sqlx.Tx, dayInterval int) error {
	if ts.table == "" {
		return errors.New("Table must not be empty.")
	}

	tx, wrapInSingleTransaction, err := ts.newTransactionIfNeeded(tx)
	if tx == nil {
		return errors.New("Transaction struct must not be empty.")
	}
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	from := now.Add(-24 * time.Hour * time.Duration(dayInterval)).UTC().Unix()

	query := fmt.Sprintf("DELETE FROM %v WHERE created < to_timestamp($1) at time zone 'utc'", ts.table)

	_, err = tx.Exec(query, from)

	if wrapInSingleTransaction == true {
		err = tx.Commit()
	}

	return err
}
