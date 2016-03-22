package dal

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"

	"github.com/resourced/resourced-master/querybuilder"
)

func NewTSLog(db *sqlx.DB) *TSLog {
	ts := &TSLog{}
	ts.db = db
	ts.table = "ts_logs"

	return ts
}

type AgentLogPayload struct {
	Host struct {
		Name string
		Tags map[string]string
	}
	Data struct {
		LogLines []string
		Filename string
	}
}

type TSLogRow struct {
	ClusterID int64               `db:"cluster_id"`
	Created   time.Time           `db:"created"`
	Hostname  string              `db:"hostname"`
	Tags      sqlx_types.JSONText `db:"tags"`
	Filename  string              `db:"filename"`
	Logline   string              `db:"logline"`
}

func (tsr *TSLogRow) GetTags() map[string]string {
	tags := make(map[string]string)
	tsr.Tags.Unmarshal(&tags)

	return tags
}

type TSLog struct {
	TSBase
}

func (ts *TSLog) CreateFromJSON(tx *sqlx.Tx, clusterID int64, jsonData []byte) error {
	payload := &AgentLogPayload{}

	err := json.Unmarshal(jsonData, payload)
	if err != nil {
		return err
	}

	return ts.Create(tx, clusterID, payload.Host.Name, payload.Host.Tags, payload.Data.LogLines, payload.Data.Filename)
}

// Create a new record.
func (ts *TSLog) Create(tx *sqlx.Tx, clusterID int64, hostname string, tags map[string]string, loglines []string, filename string) (err error) {
	if tx == nil {
		tx, err = ts.db.Beginx()
		if err != nil {
			return err
		}
	}

	query := fmt.Sprintf("INSERT INTO %v (cluster_id, created, hostname, logline, filename, tags) VALUES ($1, $2, $3, $4, $5, $6)", ts.table)

	prepared, err := ts.db.Preparex(query)
	if err != nil {
		return err
	}

	for _, logline := range loglines {
		tagsInJson, err := json.Marshal(tags)
		if err != nil {
			continue
		}

		created := time.Now().UTC()

		logFields := logrus.Fields{
			"Method":    "TSLog.Create",
			"Query":     query,
			"ClusterID": clusterID,
			"Created":   created,
			"Hostname":  hostname,
			"Logline":   logline,
			"Filename":  filename,
			"Tags":      string(tagsInJson),
		}

		_, err = prepared.Exec(clusterID, created, hostname, logline, filename, tagsInJson)
		if err != nil {
			logFields["Error"] = err.Error()
			logrus.WithFields(logFields).Error("Failed to execute insert query")
		}
		logrus.WithFields(logFields).Info("Insert Query")
	}
	return tx.Commit()
}

// AllByClusterIDAndRange returns all logs withing time range.
func (ts *TSLog) AllByClusterIDAndRange(tx *sqlx.Tx, clusterID int64, from, to int64) ([]*TSLogRow, error) {
	// Default is 15 minutes range
	if to == -1 {
		to = time.Now().UTC().Unix()
	}
	if from == -1 {
		from = to - 900
	}

	rows := []*TSLogRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND created >= to_timestamp($2) at time zone 'utc' AND created <= to_timestamp($3) at time zone 'utc' ORDER BY created DESC", ts.table)
	err := ts.db.Select(&rows, query, clusterID, from, to)

	if err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
	}
	return rows, err
}

// AllByClusterIDRangeAndQuery returns all rows by resourced query.
func (ts *TSLog) AllByClusterIDRangeAndQuery(tx *sqlx.Tx, clusterID int64, from, to int64, resourcedQuery string) ([]*TSLogRow, error) {
	pgQuery := querybuilder.Parse(resourcedQuery)
	if pgQuery == "" {
		return ts.AllByClusterIDAndRange(tx, clusterID, from, to)
	}

	rows := []*TSLogRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND created >= to_timestamp($2) at time zone 'utc' AND created <= to_timestamp($3) AND %v", ts.table, pgQuery)
	err := ts.db.Select(&rows, query, clusterID, from, to)

	return rows, err
}
