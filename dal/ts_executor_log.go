package dal

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"

	"github.com/resourced/resourced-master/querybuilder"
)

func NewTSExecutorLog(db *sqlx.DB) *TSExecutorLog {
	ts := &TSExecutorLog{}
	ts.db = db
	ts.table = "ts_executor_logs"

	return ts
}

type TSExecutorLogRow struct {
	ClusterID int64               `db:"cluster_id"`
	Created   time.Time           `db:"created"`
	Hostname  string              `db:"hostname"`
	Tags      sqlx_types.JSONText `db:"tags"`
	Logline   string              `db:"logline"`
}

type TSExecutorLogRowsWithError struct {
	TSExecutorLogRows []*TSExecutorLogRow
	Error             error
}

type TSExecutorLog struct {
	TSBase
}

func (ts *TSExecutorLog) CreateFromJSON(tx *sqlx.Tx, clusterID int64, jsonData []byte) error {
	payload := &AgentLogPayload{}

	err := json.Unmarshal(jsonData, payload)
	if err != nil {
		return err
	}

	return ts.Create(tx, clusterID, payload.Host.Name, payload.Host.Tags, payload.Data.Loglines)
}

// Create a new record.
func (ts *TSExecutorLog) Create(tx *sqlx.Tx, clusterID int64, hostname string, tags map[string]string, loglines []string) error {
	for _, logline := range loglines {
		insertData := make(map[string]interface{})
		insertData["cluster_id"] = clusterID
		insertData["hostname"] = hostname
		insertData["logline"] = logline

		tagsInJson, err := json.Marshal(tags)
		if err == nil {
			insertData["tags"] = tagsInJson
		}

		_, err = ts.InsertIntoTable(tx, insertData)
		if err != nil {
			return err
		}
	}

	return nil
}

// AllByClusterIDAndRange returns all logs withing time range.
func (ts *TSExecutorLog) AllByClusterIDAndRange(tx *sqlx.Tx, clusterID int64, from, to int64) ([]*TSExecutorLogRow, error) {
	// Default is 15 minutes range
	if to == -1 {
		to = time.Now().UTC().Unix()
	}
	if from == -1 {
		from = to - 900
	}

	rows := []*TSExecutorLogRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND created >= to_timestamp($2) at time zone 'utc' AND created <= to_timestamp($3) at time zone 'utc' ORDER BY created DESC", ts.table)
	err := ts.db.Select(&rows, query, clusterID, from, to)

	if err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
	}
	return rows, err
}

// AllByClusterIDRangeAndQuery returns all rows by resourced query.
func (ts *TSExecutorLog) AllByClusterIDRangeAndQuery(tx *sqlx.Tx, clusterID int64, from, to int64, resourcedQuery string) ([]*TSExecutorLogRow, error) {
	pgQuery := querybuilder.Parse(resourcedQuery)
	if pgQuery == "" {
		return ts.AllByClusterIDAndRange(tx, clusterID, from, to)
	}

	rows := []*TSExecutorLogRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND created >= to_timestamp($2) at time zone 'utc' AND created <= to_timestamp($3) at time zone 'utc' AND %v ORDER BY created DESC", ts.table, pgQuery)
	err := ts.db.Select(&rows, query, clusterID, from, to)

	if err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
	}
	return rows, err
}
