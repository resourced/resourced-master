package dal

import (
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

// GetByID returns record by id.
func (ts *TSEvent) GetByID(tx *sqlx.Tx, id int64) (*TSEventRow, error) {
	row := &TSEventRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=$1", ts.table)
	err := ts.db.Get(row, query, id)

	return row, err
}

// Create a new record.
func (ts *TSEvent) Create(tx *sqlx.Tx, clusterID, fromUnix, toUnix int64, description string) (*TSEventRow, error) {
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

	id := ts.NewExplicitID()

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
