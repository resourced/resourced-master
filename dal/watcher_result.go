package dal

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

func NewWatcherResult(db *sqlx.DB) *WatcherResult {
	wr := &WatcherResult{}
	wr.db = db
	wr.table = "watchers_results"

	return wr
}

type WatcherResultRow struct {
	ClusterID     int64     `db:"cluster_id"`
	WatcherID     int64     `db:"watcher_id"`
	AffectedHosts int64     `db:"affected_hosts"`
	Count         int64     `db:"count"`
	Updated       time.Time `db:"updated"`
}

type WatcherResult struct {
	Base
}

// AllByClusterID returns all watchers_results rows by cluster_id.
func (wr *WatcherResult) AllByClusterID(tx *sqlx.Tx, clusterID int64) ([]*WatcherResultRow, error) {
	rows := []*WatcherResultRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 ORDER BY cluster_id,watcher_id ASC", wr.table)
	err := wr.db.Select(&rows, query, clusterID)

	return rows, err
}

// GetByClusterIDAndWatcherID returns one watchers_results row by cluster_id and watcher_id
func (wr *WatcherResult) GetByClusterIDAndWatcherID(tx *sqlx.Tx, clusterID, watcherID int64) (*WatcherResultRow, error) {
	row := &WatcherResultRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 and watcher_id=$2", wr.table)
	err := wr.db.Get(row, query, clusterID, watcherID)

	return row, err
}

func (wr *WatcherResult) IncrementCountByClusterIDAndWatcherID(tx *sqlx.Tx, clusterID, watcherID, count int64) (*WatcherResultRow, error) {
	query := fmt.Sprintf("UPDATE %v SET count = count + $1, updated = NOW() WHERE cluster_id=$2 and watcher_id=$3", wr.table)

	_, err := tx.Exec(query, count, clusterID, watcherID)
	if err != nil {
		return nil, err
	}

	return wr.GetByClusterIDAndWatcherID(tx, clusterID, watcherID)
}

func (wr *WatcherResult) ResetCountByClusterIDAndWatcherID(tx *sqlx.Tx, clusterID, watcherID int64) (*WatcherResultRow, error) {
	query := fmt.Sprintf("UPDATE %v SET count = 0, updated = NOW() WHERE cluster_id=$1 and watcher_id=$2", wr.table)

	_, err := tx.Exec(query, clusterID, watcherID)
	if err != nil {
		return nil, err
	}

	return wr.GetByClusterIDAndWatcherID(tx, clusterID, watcherID)
}
