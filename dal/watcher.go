package dal

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"
)

func NewWatcher(db *sqlx.DB) *Watcher {
	watcher := &Watcher{}
	watcher.db = db
	watcher.table = "watchers"
	watcher.hasID = true

	return watcher
}

type WatcherRow struct {
	ID               int64               `db:"id" json:"-"`
	ClusterID        int64               `db:"cluster_id"`
	SavedQueryID     int64               `db:"saved_query_id"`
	SavedQuery       string              `db:"saved_query"`
	LowThreshold     int64               `db:"low_threshold"`
	HighThreshold    int64               `db:"high_threshold"`
	LowAffectedHosts int64               `db:"low_affected_hosts"`
	HostsLastUpdated string              `db:"hosts_last_updated"`
	WaitInterval     string              `db:"wait_interval"`
	Actions          sqlx_types.JsonText `db:"actions"`
}

type Watcher struct {
	Base
}

// AllByClusterID returns all watchers rows.
func (w *Watcher) AllByClusterID(tx *sqlx.Tx, clusterID int64) ([]*WatcherRow, error) {
	rows := []*WatcherRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1", w.table)
	err := w.db.Select(&rows, query, clusterID)

	return rows, err
}

func (w *Watcher) rowFromSqlResult(tx *sqlx.Tx, sqlResult sql.Result) (*WatcherRow, error) {
	id, err := sqlResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	return w.GetByID(tx, id)
}

// GetByID returns one record by id.
func (w *Watcher) GetByID(tx *sqlx.Tx, id int64) (*WatcherRow, error) {
	watcherRow := &WatcherRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=$1", w.table)
	err := w.db.Get(watcherRow, query, id)

	return watcherRow, err
}

func (w *Watcher) Create(tx *sqlx.Tx, clusterID, savedQueryID int64, savedQuery string, lowThreshold, highThreshold, lowAffectedHosts int64, hostsLastUpdated, waitInterval string, actions []byte) (*WatcherRow, error) {
	data := make(map[string]interface{})
	data["cluster_id"] = clusterID
	data["saved_query_id"] = savedQueryID
	data["saved_query"] = savedQuery
	data["low_threshold"] = lowThreshold
	data["high_threshold"] = highThreshold
	data["low_affected_hosts"] = lowAffectedHosts
	data["hosts_last_updated"] = hostsLastUpdated
	data["wait_interval"] = waitInterval
	data["actions"] = actions

	sqlResult, err := w.InsertIntoTable(tx, data)
	if err != nil {
		return nil, err
	}

	return w.rowFromSqlResult(tx, sqlResult)
}
