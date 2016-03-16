package dal

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"
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
	LogLines []string
	Filename string
}

type TSLogRow struct {
	ClusterID int64               `db:"cluster_id"`
	Created   time.Time           `db:"created"`
	Hostname  string              `db:"hostname"`
	Tags      sqlx_types.JSONText `db:"tags"`
	Filename  string              `db:"filename"`
	Logline   string              `db:"logline"`
}

type TSLog struct {
	Base
}

func (ts *TSLog) CreateFromJSON(tx *sqlx.Tx, clusterID int64, jsonData []byte) error {
	payload := &AgentLogPayload{}

	err := json.Unmarshal(jsonData, payload)
	if err != nil {
		return err
	}

	return ts.Create(tx, clusterID, payload.Host.Name, payload.Host.Tags, payload.LogLines, payload.Filename)
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
