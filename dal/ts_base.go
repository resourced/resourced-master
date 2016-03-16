package dal

import (
	"errors"
	"fmt"

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

	query := fmt.Sprintf("DELETE FROM %v WHERE created < (NOW() at time zone 'utc' - INTERVAL '%v day')", ts.table, dayInterval)

	_, err = tx.Exec(query)

	if wrapInSingleTransaction == true {
		err = tx.Commit()
	}

	return err
}
