package dal

import (
	"database/sql"
	"encoding/json"
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
	Name             string              `db:"name"`
	LowThreshold     int64               `db:"low_threshold"`
	HighThreshold    int64               `db:"high_threshold"`
	LowAffectedHosts int64               `db:"low_affected_hosts"`
	HostsLastUpdated string              `db:"hosts_last_updated"`
	CheckInterval    string              `db:"check_interval"`
	Actions          sqlx_types.JsonText `db:"actions"`
}

func (wr *WatcherRow) ActionTransport() string {
	actions := make(map[string]interface{})

	err := json.Unmarshal(wr.Actions, &actions)
	if err != nil {
		return ""
	}

	transportInterface := actions["Transport"]
	if transportInterface == nil {
		return ""
	}

	return transportInterface.(string)
}

func (wr *WatcherRow) ActionEmail() string {
	actions := make(map[string]interface{})

	err := json.Unmarshal(wr.Actions, &actions)
	if err != nil {
		return ""
	}

	emailInterface := actions["Email"]
	if emailInterface == nil {
		return ""
	}

	return emailInterface.(string)
}

type Watcher struct {
	Base
}

// All returns all watchers rows.
func (w *Watcher) All(tx *sqlx.Tx) ([]*WatcherRow, error) {
	rows := []*WatcherRow{}
	query := fmt.Sprintf("SELECT * FROM %v", w.table)
	err := w.db.Select(&rows, query)

	return rows, err
}

// AllGroupByDaemons returns all watchers rows divided into daemons equally.
func (w *Watcher) AllSplitToDaemons(tx *sqlx.Tx, daemons []string) (map[string][]*WatcherRow, error) {
	watcherRows, err := w.All(tx)
	if err != nil {
		return nil, err
	}

	buckets := make([][]*WatcherRow, len(daemons))
	for i, _ := range daemons {
		buckets[i] = make([]*WatcherRow, 0)
	}

	bucketsPointer := 0
	for _, watcherRow := range watcherRows {
		buckets[bucketsPointer] = append(buckets[bucketsPointer], watcherRow)
		bucketsPointer = bucketsPointer + 1

		if bucketsPointer >= len(buckets) {
			bucketsPointer = 0
		}
	}

	result := make(map[string][]*WatcherRow)

	for i, watchersInbucket := range buckets {
		result[daemons[i]] = watchersInbucket
	}

	return result, err
}

// AllByClusterID returns all watchers rows by cluster_id.
func (w *Watcher) AllByClusterID(tx *sqlx.Tx, clusterID int64) ([]*WatcherRow, error) {
	rows := []*WatcherRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 ORDER BY name ASC", w.table)
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

func (w *Watcher) CreateOrUpdateParameters(clusterID, savedQueryID int64, savedQuery, name string, lowThreshold, highThreshold, lowAffectedHosts int64, hostsLastUpdated, checkInterval string, actions []byte) map[string]interface{} {
	data := make(map[string]interface{})
	data["cluster_id"] = clusterID
	data["saved_query_id"] = savedQueryID
	data["saved_query"] = savedQuery
	data["name"] = name
	data["low_threshold"] = lowThreshold
	data["high_threshold"] = highThreshold
	data["low_affected_hosts"] = lowAffectedHosts
	data["hosts_last_updated"] = hostsLastUpdated
	data["check_interval"] = checkInterval
	data["actions"] = actions

	return data
}

func (w *Watcher) Create(tx *sqlx.Tx, data map[string]interface{}) (*WatcherRow, error) {
	sqlResult, err := w.InsertIntoTable(tx, data)
	if err != nil {
		return nil, err
	}

	return w.rowFromSqlResult(tx, sqlResult)
}
