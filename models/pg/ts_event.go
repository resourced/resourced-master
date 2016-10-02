package pg

import (
	"encoding/json"
	"errors"
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
	Deleted     time.Time `db:"deleted"`
	Description string    `db:"description"`
}

type TSEvent struct {
	TSBase
}

// AllLinesByClusterIDAndCreatedFromRangeForHighchart returns all rows given created_from range.
func (ts *TSEvent) AllLinesByClusterIDAndCreatedFromRangeForHighchart(tx *sqlx.Tx, clusterID, from, to, deletedFrom int64) ([]TSEventHighchartLinePayload, error) {
	rows := []*TSEventRow{}
	query := fmt.Sprintf(`SELECT * FROM %v WHERE cluster_id=$1 AND
created_from = created_to AND
created_from >= to_timestamp($2) at time zone 'utc' AND
created_from <= to_timestamp($3) at time zone 'utc' AND
deleted >= to_timestamp($4) at time zone 'utc'`, ts.table)

	err := ts.db.Select(&rows, query, clusterID, from, to, deletedFrom)
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

// AllBandsByClusterIDAndCreatedFromRangeForHighchart returns all rows with time stretch between created_from and created_to.
func (ts *TSEvent) AllBandsByClusterIDAndCreatedFromRangeForHighchart(tx *sqlx.Tx, clusterID, from, to, deletedFrom int64) ([]TSEventHighchartLinePayload, error) {
	rows := []*TSEventRow{}
	query := fmt.Sprintf(`SELECT * FROM %v WHERE cluster_id=$1 AND
created_from < created_to AND
created_from >= to_timestamp($2) at time zone 'utc' AND
created_from <= to_timestamp($3) at time zone 'utc' AND
deleted >= to_timestamp($4) at time zone 'utc'`, ts.table)

	err := ts.db.Select(&rows, query, clusterID, from, to, deletedFrom)
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

func (ts *TSEvent) CreateFromJSON(tx *sqlx.Tx, id, clusterID int64, jsonData []byte, deletedFrom int64) (*TSEventRow, error) {
	payload := &TSEventCreatePayload{}

	err := json.Unmarshal(jsonData, payload)
	if err != nil {
		return nil, err
	}

	return ts.Create(tx, id, clusterID, payload.From, payload.To, payload.Description, deletedFrom)
}

// Create a new record.
func (ts *TSEvent) Create(tx *sqlx.Tx, id, clusterID, fromUnix, toUnix int64, description string, deletedFrom int64) (*TSEventRow, error) {
	var from time.Time
	var to time.Time

	if fromUnix <= 0 {
		from = time.Now().UTC()
	} else {
		from = time.Unix(fromUnix, 0).UTC()
	}

	if toUnix <= 0 {
		to = from
	} else {
		to = time.Unix(toUnix, 0).UTC()
	}

	insertData := make(map[string]interface{})
	insertData["id"] = id
	insertData["cluster_id"] = clusterID
	insertData["created_from"] = from
	insertData["created_to"] = to
	insertData["description"] = description
	insertData["deleted"] = time.Unix(deletedFrom, 0).UTC()

	_, err := ts.InsertIntoTable(tx, insertData)
	if err != nil {
		return nil, err
	}

	return ts.GetByID(tx, id)
}

// DeleteDeleted deletes all record older than x days ago.
func (ts *TSEvent) DeleteDeleted(tx *sqlx.Tx, clusterID int64) error {
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

	now := time.Now().UTC().Unix()
	query := fmt.Sprintf("DELETE FROM %v WHERE cluster_id=$1 AND deleted < to_timestamp($2) at time zone 'utc'", ts.table)

	_, err = tx.Exec(query, clusterID, now)

	if wrapInSingleTransaction == true {
		err = tx.Commit()
	}

	return err
}
