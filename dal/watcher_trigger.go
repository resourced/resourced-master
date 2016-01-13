package dal

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"
)

func NewWatcherTrigger(db *sqlx.DB) *WatcherTrigger {
	wt := &WatcherTrigger{}
	wt.db = db
	wt.table = "watchers_triggers"
	wt.hasID = true

	return wt
}

type WatcherTriggerRow struct {
	ID                  int64               `db:"id" json:"-"`
	ClusterID           int64               `db:"cluster_id"`
	WatcherID           int64               `db:"watcher_id"`
	LowViolationsCount  int64               `db:"low_violations_count"`
	HighViolationsCount int64               `db:"high_violations_count"`
	Actions             sqlx_types.JsonText `db:"actions"`
}

func (wt *WatcherTriggerRow) ActionTransport() string {
	actions := make(map[string]interface{})

	err := json.Unmarshal(wt.Actions, &actions)
	if err != nil {
		return ""
	}

	transportInterface := actions["Transport"]
	if transportInterface == nil {
		return ""
	}

	return transportInterface.(string)
}

func (wt *WatcherTriggerRow) ActionEmail() string {
	actions := make(map[string]interface{})

	err := json.Unmarshal(wt.Actions, &actions)
	if err != nil {
		return ""
	}

	emailInterface := actions["Email"]
	if emailInterface == nil {
		return ""
	}

	return emailInterface.(string)
}

func (wt *WatcherTriggerRow) ActionSMSPhone() string {
	actions := make(map[string]interface{})

	err := json.Unmarshal(wt.Actions, &actions)
	if err != nil {
		return ""
	}

	phoneInterface := actions["SMSPhone"]
	if phoneInterface == nil {
		return ""
	}

	return phoneInterface.(string)
}

func (wt *WatcherTriggerRow) ActionSMSCarrier() string {
	actions := make(map[string]interface{})

	err := json.Unmarshal(wt.Actions, &actions)
	if err != nil {
		return ""
	}

	carrierInterface := actions["SMSCarrier"]
	if carrierInterface == nil {
		return ""
	}

	return carrierInterface.(string)
}

func (wt *WatcherTriggerRow) ActionPagerDutyServiceKey() string {
	actions := make(map[string]interface{})

	err := json.Unmarshal(wt.Actions, &actions)
	if err != nil {
		return ""
	}

	pdServiceKeyInterface := actions["PagerDutyServiceKey"]
	if pdServiceKeyInterface == nil {
		return ""
	}

	return pdServiceKeyInterface.(string)
}

func (wt *WatcherTriggerRow) ActionPagerDutyDescription() string {
	actions := make(map[string]interface{})

	err := json.Unmarshal(wt.Actions, &actions)
	if err != nil {
		return ""
	}

	pdDescriptionInterface := actions["PagerDutyDescription"]
	if pdDescriptionInterface == nil {
		return ""
	}

	return pdDescriptionInterface.(string)
}

type WatcherTrigger struct {
	Base
}

// All returns all rows.
func (w *WatcherTrigger) All(tx *sqlx.Tx) ([]*WatcherTriggerRow, error) {
	rows := []*WatcherTriggerRow{}
	query := fmt.Sprintf("SELECT * FROM %v", w.table)
	err := w.db.Select(&rows, query)

	return rows, err
}

// AllByClusterID returns all rows by cluster_id.
func (w *WatcherTrigger) AllByClusterID(tx *sqlx.Tx, clusterID int64) ([]*WatcherTriggerRow, error) {
	rows := []*WatcherTriggerRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 ORDER BY cluster_id,watcher_id,low_violations_count ASC", w.table)
	err := w.db.Select(&rows, query, clusterID)

	return rows, err
}

// AllByClusterIDAndWatcherID returns all rows by cluster_id and watcher_id.
func (w *WatcherTrigger) AllByClusterIDAndWatcherID(tx *sqlx.Tx, clusterID, watcherID int64) ([]*WatcherTriggerRow, error) {
	rows := []*WatcherTriggerRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND watcher_id=$2 ORDER BY cluster_id,watcher_id,low_violations_count ASC", w.table)
	err := w.db.Select(&rows, query, clusterID, watcherID)

	return rows, err
}

func (w *WatcherTrigger) rowFromSqlResult(tx *sqlx.Tx, sqlResult sql.Result) (*WatcherTriggerRow, error) {
	id, err := sqlResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	return w.GetByID(tx, id)
}

// GetByID returns one record by id.
func (w *WatcherTrigger) GetByID(tx *sqlx.Tx, id int64) (*WatcherTriggerRow, error) {
	watcherRow := &WatcherTriggerRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=$1", w.table)
	err := w.db.Get(watcherRow, query, id)

	return watcherRow, err
}

// CreateOrUpdateParameters builds params for insert or update.
func (w *WatcherTrigger) CreateOrUpdateParameters(clusterID, watcherID, lowViolationsCount, highViolationsCount int64, actionsJson []byte) map[string]interface{} {
	data := make(map[string]interface{})
	data["cluster_id"] = clusterID
	data["watcher_id"] = watcherID
	data["low_violations_count"] = lowViolationsCount
	data["high_violations_count"] = highViolationsCount
	data["actions"] = actionsJson

	return data
}

// Create inserts one row
func (w *WatcherTrigger) Create(tx *sqlx.Tx, data map[string]interface{}) (*WatcherTriggerRow, error) {
	sqlResult, err := w.InsertIntoTable(tx, data)
	if err != nil {
		return nil, err
	}

	return w.rowFromSqlResult(tx, sqlResult)
}
