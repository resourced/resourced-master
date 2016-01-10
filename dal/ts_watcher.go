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

// AllByClusterIDAndCreatedInterval returns all rows by cluster_id and watcher_id within N interval.
func (ts *TSWatcher) AllByClusterIDAndCreatedInterval(tx *sqlx.Tx, clusterID, watcherID int64, createdInterval string) ([]*TSWatcherRow, error) {
	rows := []*TSWatcherRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND created >= (NOW() - INTERVAL '%v') ORDER BY cluster_id,watcher_id,created DESC", ts.table, createdInterval)
	err := ts.db.Select(&rows, query, clusterID, watcherID)

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
