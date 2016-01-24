package dal

import (
	"errors"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
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

// LastByClusterIDWatcherIDAndAffectedHosts returns a row by cluster_id, watcher_id and affected_hosts.
func (ts *TSWatcher) LastByClusterIDWatcherIDAndAffectedHosts(tx *sqlx.Tx, clusterID, watcherID, affectedHosts int64) (*TSWatcherRow, error) {
	row := &TSWatcherRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND watcher_id=$2 AND affected_hosts=$3 ORDER BY cluster_id,watcher_id,created DESC LIMIT 1", ts.table)
	err := ts.db.Get(row, query, clusterID, watcherID, affectedHosts)

	logrus.WithFields(logrus.Fields{
		"Method":        "TSWatcher.LastByClusterIDWatcherIDAndAffectedHosts",
		"ClusterID":     clusterID,
		"WatcherID":     watcherID,
		"AffectedHosts": affectedHosts,
		"Query":         query,
	}).Info("Select Query")

	return row, err
}

// AllViolationsByClusterIDWatcherIDAndInterval returns all rows by cluster_id, watcher_id, affectedHosts and created interval.
func (ts *TSWatcher) AllViolationsByClusterIDWatcherIDAndInterval(tx *sqlx.Tx, clusterID, watcherID, affectedHosts int64, createdInterval string) ([]*TSWatcherRow, error) {
	nonAffectedHosts := affectedHosts - 1

	lastGoodOne, err := ts.LastByClusterIDWatcherIDAndAffectedHosts(tx, clusterID, watcherID, nonAffectedHosts)
	if err != nil {
		return nil, err
	}

	rows := []*TSWatcherRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND watcher_id=$2 AND created > GREATEST($3, (NOW() at time zone 'utc' - INTERVAL '%v')) AND created <= (NOW() at time zone 'utc') AND affected_hosts >= $4 ORDER BY cluster_id,watcher_id,created DESC", ts.table, createdInterval)
	err = ts.db.Select(&rows, query, clusterID, watcherID, lastGoodOne.Created.UTC(), affectedHosts)

	logrus.WithFields(logrus.Fields{
		"Method":          "TSWatcher.AllViolationsByClusterIDWatcherIDAndInterval",
		"ClusterID":       clusterID,
		"WatcherID":       watcherID,
		"AffectedHosts":   affectedHosts,
		"LastGoodCreated": lastGoodOne.Created.UTC(),
		"Query":           query,
	}).Info("Select Query")

	return rows, err
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
