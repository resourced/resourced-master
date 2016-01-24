package dal

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"
)

func NewTSWatcher(db *sqlx.DB) *TSWatcher {
	ts := &TSWatcher{}
	ts.db = db
	ts.table = "ts_watchers"

	return ts
}

type TSWatcherRow struct {
	ClusterID     int64               `db:"cluster_id"`
	WatcherID     int64               `db:"watcher_id"`
	AffectedHosts int64               `db:"affected_hosts"`
	Created       time.Time           `db:"created"`
	Data          sqlx_types.JSONText `db:"data"`
}

type TSWatcher struct {
	Base
}

// LastGreenMarker returns a row where affected_hosts is 0.
func (ts *TSWatcher) LastGreenMarker(tx *sqlx.Tx, clusterID, watcherID int64) (*TSWatcherRow, error) {
	row := &TSWatcherRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND watcher_id=$2 AND affected_hosts=0 ORDER BY cluster_id,watcher_id,created DESC LIMIT 1", ts.table)
	err := ts.db.Get(row, query, clusterID, watcherID)

	return row, err
}

// CountViolationsSinceLastGreenMarker
func (ts *TSWatcher) CountViolationsSinceLastGreenMarker(tx *sqlx.Tx, clusterID, watcherID int64) (int, error) {
	lastGreenMarker, err := ts.LastGreenMarker(tx, clusterID, watcherID)
	if err != nil {
		if !strings.Contains(err.Error(), "no rows in result set") {
			return -1, err
		}
	}

	var count int

	if lastGreenMarker == nil {
		query := fmt.Sprintf("SELECT count(*) FROM %v WHERE cluster_id=$1 AND watcher_id=$2 AND affected_hosts != 0", ts.table)

		err = ts.db.Get(&count, query, clusterID, watcherID)
		if err != nil {
			if strings.Contains(err.Error(), "no rows in result set") {
				return 0, nil
			}
			return -1, err
		}

	} else {
		query := fmt.Sprintf("SELECT count(*) FROM %v WHERE cluster_id=$1 AND watcher_id=$2 AND affected_hosts != 0 AND created > $3", ts.table)

		err = ts.db.Get(&count, query, clusterID, watcherID, lastGreenMarker.Created)
		if err != nil {
			if strings.Contains(err.Error(), "no rows in result set") {
				return 0, nil
			}
			return -1, err
		}
	}

	return count, nil
}

// LastViolation returns the last row where affected_hosts is not 0.
func (ts *TSWatcher) LastViolation(tx *sqlx.Tx, clusterID, watcherID int64) (*TSWatcherRow, error) {
	row := &TSWatcherRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND watcher_id=$2 AND affected_hosts != 0 ORDER BY cluster_id,watcher_id,created DESC LIMIT 1", ts.table)
	err := ts.db.Get(row, query, clusterID, watcherID)

	return row, err
}

// Create a new record.
func (ts *TSWatcher) Create(tx *sqlx.Tx, clusterID, watcherID, affectedHosts int64, data []byte) error {
	insertData := make(map[string]interface{})
	insertData["cluster_id"] = clusterID
	insertData["watcher_id"] = watcherID
	insertData["affected_hosts"] = affectedHosts
	insertData["data"] = data

	_, err := ts.InsertIntoTable(tx, insertData)
	return err
}

// DeleteByDayInterval deletes all record older than x days ago.
func (ts *TSWatcher) DeleteByDayInterval(tx *sqlx.Tx, dayInterval int) error {
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
