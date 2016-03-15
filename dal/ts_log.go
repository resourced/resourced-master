package dal

import (
	"encoding/json"
	"time"

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

	return ts.Create(tx, clusterID, payload.Host.Name, payload.Host.Tags, payload.LogLines, "")
}

// Create a new record.
func (ts *TSLog) Create(tx *sqlx.Tx, clusterID int64, hostname string, tags map[string]string, loglines []string, filename string) error {
	for _, logline := range loglines {
		insertData := make(map[string]interface{})
		insertData["cluster_id"] = clusterID
		insertData["created"] = time.Now().UTC()
		insertData["hostname"] = hostname
		insertData["logline"] = logline
		insertData["filename"] = filename

		tagsInJson, err := json.Marshal(tags)
		if err != nil {
			continue
		}
		insertData["tags"] = tagsInJson

		_, err = ts.InsertIntoTable(tx, insertData)
		if err != nil {
			return err
		}
	}

	return nil
}
