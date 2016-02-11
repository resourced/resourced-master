package dal

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

func NewTSEvent(db *sqlx.DB) *TSEvent {
	ts := &TSEvent{}
	ts.db = db
	ts.table = "ts_events"

	return ts
}

type TSEventHighchartLinePayload struct {
	ID          int64  `json:"ID"`
	CreatedFrom int64  `json:"CreatedFrom"`
	CreatedTo   int64  `json:"CreatedTo"`
	Description string `json:"Description"`
}

type TSEventCreatePayload struct {
	From        int64  `json:"from"`
	To          int64  `json:"to"`
	Description string `json:"description"`
}

type TSEventRow struct {
	ID          int64     `db:"id"`
	ClusterID   int64     `db:"cluster_id"`
	CreatedFrom time.Time `db:"created_from"`
	CreatedTo   time.Time `db:"created_to"`
	Description string    `db:"description"`
}

type TSEvent struct {
	Base
}

// AllLinesByClusterIDAndCreatedFromInterval returns all rows without time stretch between created_from and created_to.
func (ts *TSEvent) AllLinesByClusterIDAndCreatedFromIntervalForHighchart(tx *sqlx.Tx, clusterID int64, createdInterval string) ([]TSEventHighchartLinePayload, error) {
	rows := []*TSEventRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND created_from = created_to AND created_from >= (NOW() at time zone 'utc' - INTERVAL '%v')", ts.table, createdInterval)
	err := ts.db.Select(&rows, query, clusterID)
	if err != nil {
		return nil, err
	}

	hcRows := make([]TSEventHighchartLinePayload, len(rows))

	for i, row := range rows {
		hcRow := TSEventHighchartLinePayload{}
		hcRow.ID = row.ID
		hcRow.CreatedFrom = row.CreatedFrom.UnixNano() / 1000000
		hcRow.CreatedTo = row.CreatedTo.UnixNano() / 1000000
		hcRow.Description = row.Description

		hcRows[i] = hcRow
	}

	return hcRows, err
}

// GetByID returns record by id.
func (ts *TSEvent) GetByID(tx *sqlx.Tx, id int64) (*TSEventRow, error) {
	row := &TSEventRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=$1", ts.table)
	err := ts.db.Get(row, query, id)

	return row, err
}

func (ts *TSEvent) CreateFromJSON(tx *sqlx.Tx, id, clusterID int64, jsonData []byte) (*TSEventRow, error) {
	payload := &TSEventCreatePayload{}

	err := json.Unmarshal(jsonData, payload)
	if err != nil {
		return nil, err
	}

	return ts.Create(tx, id, clusterID, payload.From, payload.To, payload.Description)
}

// Create a new record.
func (ts *TSEvent) Create(tx *sqlx.Tx, id, clusterID, fromUnix, toUnix int64, description string) (*TSEventRow, error) {
	var from time.Time
	var to time.Time

	if fromUnix <= 0 {
		from = time.Now().UTC()
	} else {
		from = time.Unix(fromUnix, 0)
	}

	if toUnix <= 0 {
		to = from
	} else {
		to = time.Unix(fromUnix, 0)
	}

	insertData := make(map[string]interface{})
	insertData["id"] = id
	insertData["cluster_id"] = clusterID
	insertData["created_from"] = from
	insertData["created_to"] = to
	insertData["description"] = description

	_, err := ts.InsertIntoTable(tx, insertData)
	if err != nil {
		return nil, err
	}

	return ts.GetByID(tx, id)
}

// DeleteByID deletes record by id.
func (ts *TSEvent) DeleteByID(tx *sqlx.Tx, id int64) (sql.Result, error) {
	query := fmt.Sprintf("DELETE FROM %v WHERE id=$1", ts.table)
	return ts.db.Exec(query, id)
}
