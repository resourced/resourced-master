package dal

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"
	"time"
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
	Data          sqlx_types.JsonText `db:"data"`
}

type TSWatcher struct {
	Base
}

// LastGreenMarker returns a row where affected_hosts is 0.
func (ts *TSWatcher) LastGreenMarker(tx *sqlx.Tx) (*TSWatcherRow, error) {
	row := &TSWatcherRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE affected_hosts=0 ORDER BY cluster_id,watcher_id,created DESC LIMIT 1", ts.table)
	err := ts.db.Get(row, query)

	return row, err
}

// CountViolationsSinceLastGreenMarker
func (ts *TSWatcher) CountViolationsSinceLastGreenMarker(tx *sqlx.Tx) (int, error) {
	lastGreenMarker, err := ts.LastGreenMarker(tx)
	if err != nil {
		return -1, err
	}

	var count int
	query := fmt.Sprintf("SELECT count(*) FROM %v WHERE affected_hosts != 0 AND cluster_id=$1 AND watcher_id=$2 AND created > $3", ts.table)

	err = ts.db.Get(&count, query, lastGreenMarker.ClusterID, lastGreenMarker.WatcherID, lastGreenMarker.Created)
	if err != nil {
		return -1, err
	}

	return count, nil
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
