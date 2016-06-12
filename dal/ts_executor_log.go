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
	Deleted   time.Time           `db:"deleted"`
	Hostname  string              `db:"hostname"`
	Tags      sqlx_types.JSONText `db:"tags"`
	Logline   string              `db:"logline"`
}

func (tsr *TSExecutorLogRow) GetTags() map[string]string {
	tags := make(map[string]string)
	tsr.Tags.Unmarshal(&tags)

	return tags
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

// LastByClusterID returns the last row by cluster id.
func (ts *TSExecutorLog) LastByClusterID(tx *sqlx.Tx, clusterID int64) (*TSExecutorLogRow, error) {
	row := &TSExecutorLogRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 ORDER BY created DESC limit 1", ts.table)
	err := ts.db.Get(row, query, clusterID)

	return row, err
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

// AllByClusterIDLastRowIntervalAndQuery returns all rows by cluster id, unix timestamp range since last row, and resourced query.
func (ts *TSExecutorLog) AllByClusterIDLastRowIntervalAndQuery(tx *sqlx.Tx, clusterID int64, createdInterval, resourcedQuery string) ([]*TSExecutorLogRow, error) {
	lastRow, err := ts.LastByClusterID(tx, clusterID)
	if err != nil {
		return nil, err
	}

	pgQuery := querybuilder.Parse(resourcedQuery)
	rows := []*TSExecutorLogRow{}

	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND created >= ($2 at time zone 'utc' - INTERVAL '%v')", ts.table, createdInterval)
	if pgQuery != "" {
		query = fmt.Sprintf("%v AND %v ORDER BY created DESC", query, pgQuery)
	}

	err = ts.db.Select(&rows, query, clusterID, lastRow.Created)

	if err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
	}
	return rows, err
}
