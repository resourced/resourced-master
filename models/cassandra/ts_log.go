package cassandra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gocql/gocql"

	"github.com/resourced/resourced-master/contexthelper"
	"github.com/resourced/resourced-master/libstring"
	"github.com/resourced/resourced-master/models/cassandra/querybuilder"
	"github.com/resourced/resourced-master/models/shared"
)

func NewTSLog(ctx context.Context) *TSLog {
	ts := &TSLog{}
	ts.AppContext = ctx
	ts.table = "ts_logs"

	return ts
}

type TSLogRowsWithError struct {
	TSLogRows []*TSLogRow
	Error     error
}

type TSLogRow struct {
	ClusterID int64             `db:"cluster_id"`
	Created   int64             `db:"created"`
	Hostname  string            `db:"hostname"`
	Tags      map[string]string `db:"tags"`
	Filename  string            `db:"filename"`
	Logline   string            `db:"logline"`
}

func (tsr *TSLogRow) GetTags() map[string]string {
	return tsr.Tags
}

func (tsr *TSLogRow) CreatedUnix() int64 {
	return tsr.Created
}

type TSLog struct {
	Base
}

func (ts *TSLog) GetCassandraSession() (*gocql.Session, error) {
	cassandradbs, err := contexthelper.GetCassandraDBConfig(ts.AppContext)
	if err != nil {
		return nil, err
	}

	return cassandradbs.TSLogSession, nil
}

func (ts *TSLog) CreateFromJSON(clusterID int64, jsonData []byte, ttl time.Duration) error {
	payload := &shared.AgentLogPayload{}

	err := json.Unmarshal(jsonData, payload)
	if err != nil {
		return err
	}

	return ts.Create(clusterID, payload.Host.Name, payload.Host.Tags, payload.Data.Loglines, payload.Data.Filename, ttl)
}

// Create a new record.
func (ts *TSLog) Create(clusterID int64, hostname string, tags map[string]string, loglines []shared.AgentLoglinePayload, filename string, ttl time.Duration) (err error) {
	session, err := ts.GetCassandraSession()
	if err != nil {
		return err
	}

	query := fmt.Sprintf("INSERT INTO %v (id, cluster_id, hostname, logline, filename, tags, created) VALUES (?, ?, ?, ?, ?, ?, ?) USING TTL ?", ts.table)

	for _, loglinePayload := range loglines {
		id := time.Now().UTC().UnixNano()

		// Try to parse created from payload, if not use current unix timestamp
		created := time.Now().UTC().Unix()
		if loglinePayload.Created > 0 {
			created = loglinePayload.Created
		}

		content := loglinePayload.Content

		// Format JSON to regular text
		if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
			content = libstring.JSONToText(content)
		}

		err = session.Query(
			query,
			id,
			clusterID,
			hostname,
			content,
			filename,
			tags,
			created,
			ttl,
		).Exec()

		if err != nil {
			tagsInJson, err := json.Marshal(tags)
			if err != nil {
				tagsInJson = []byte("{}")
			}

			logrus.WithFields(logrus.Fields{
				"Method":    "TSLog.Create",
				"Query":     query,
				"ClusterID": clusterID,
				"Hostname":  hostname,
				"Logline":   content,
				"Filename":  filename,
				"Tags":      string(tagsInJson),
				"Error":     err.Error(),
			}).Error("Failed to execute insert query")
			continue
		}
	}

	return err
}

// LastByClusterID returns the last row by cluster id.
func (ts *TSLog) LastByClusterID(clusterID int64) (*TSLogRow, error) {
	session, err := ts.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	var scannedClusterID, scannedCreated int64
	var scannedHostname, scannedLogline, scannedFilename string
	var scannedTags map[string]string

	query := fmt.Sprintf(`SELECT cluster_id, hostname, logline, filename, tags, created FROM %v WHERE lucene='{
    filter: {type: "match", field: "cluster_id", value: %v},
    sort: {field: "created", reverse: true}
}' limit 1;`, ts.table, clusterID)

	err = session.Query(query).Scan(&scannedClusterID, &scannedHostname, &scannedLogline, &scannedFilename, &scannedTags, &scannedCreated)
	if err != nil {
		return nil, err
	}

	row := &TSLogRow{
		ClusterID: scannedClusterID,
		Hostname:  scannedHostname,
		Logline:   scannedLogline,
		Filename:  scannedFilename,
		Tags:      scannedTags,
		Created:   scannedCreated,
	}

	return row, err
}

// AllByClusterIDAndRange returns all logs within time range.
func (ts *TSLog) AllByClusterIDAndRange(clusterID int64, from, to int64) ([]*TSLogRow, error) {
	session, err := ts.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	// Default is 15 minutes range
	if to == -1 {
		to = time.Now().UTC().Unix()
	}
	if from == -1 {
		from = to - 900
	}

	rows := []*TSLogRow{}

	query := fmt.Sprintf(`SELECT cluster_id, hostname, logline, filename, tags, created FROM %v WHERE lucene='{
    filter: {
        type: "boolean",
        must: [
            {type: "match", field: "cluster_id", value: %v},
            {type:"range", field:"created", lower:%v, upper:%v}
        ]
    },
    sort: {field: "created", reverse: true}
}';`, ts.table, clusterID, from, to)

	var scannedClusterID, scannedCreated int64
	var scannedLogline, scannedHostname, scannedFilename string
	var scannedTags map[string]string

	iter := session.Query(query).Iter()
	for iter.Scan(&scannedClusterID, &scannedHostname, &scannedLogline, &scannedFilename, &scannedTags, &scannedCreated) {
		rows = append(rows, &TSLogRow{
			ClusterID: scannedClusterID,
			Filename:  scannedFilename,
			Logline:   scannedLogline,
			Hostname:  scannedHostname,
			Tags:      scannedTags,
			Created:   scannedCreated,
		})
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{
			"Method":    "TSMetric.AllByClusterIDAndRange",
			"ClusterID": clusterID,
			"From":      from,
			"To":        to,
		}).Error(err)

		return nil, err
	}

	return rows, err
}

// AllByClusterIDRangeAndQuery returns all rows by cluster id, unix timestamp range, and resourced query.
func (ts *TSLog) AllByClusterIDRangeAndQuery(clusterID int64, from, to int64, resourcedQuery string) ([]*TSLogRow, error) {
	session, err := ts.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	luceneQuery := querybuilder.Parse(resourcedQuery)
	if luceneQuery == "" {
		return ts.AllByClusterIDAndRange(clusterID, from, to)
	}

	rows := []*TSLogRow{}

	query := fmt.Sprintf(`SELECT cluster_id, hostname, logline, filename, tags, created FROM %v WHERE lucene='{
    filter: {
        type: "boolean",
        must: [
            {type: "match", field: "cluster_id", value: %v},
            {type:"range", field:"created", lower:%v, upper:%v}
        ]
    },
    query: %v,
    sort: {field: "created", reverse: true}
}';`, ts.table, clusterID, from, to, luceneQuery)

	var scannedClusterID, scannedCreated int64
	var scannedLogline, scannedHostname, scannedFilename string
	var scannedTags map[string]string

	iter := session.Query(query).Iter()
	for iter.Scan(&scannedClusterID, &scannedHostname, &scannedLogline, &scannedFilename, &scannedTags, &scannedCreated) {
		rows = append(rows, &TSLogRow{
			ClusterID: scannedClusterID,
			Filename:  scannedFilename,
			Logline:   scannedLogline,
			Hostname:  scannedHostname,
			Tags:      scannedTags,
			Created:   scannedCreated,
		})
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{
			"Method":    "TSMetric.AllByClusterIDRangeAndQuery",
			"ClusterID": clusterID,
			"From":      from,
			"To":        to,
		}).Error(err)

		return nil, err
	}

	return rows, err
}

// CountByClusterIDFromTimestampHostAndQuery returns count by cluster id, from unix timestamp, host, and resourced query.
func (ts *TSLog) CountByClusterIDFromTimestampHostAndQuery(clusterID int64, from int64, hostname, resourcedQuery string) (int64, error) {
	session, err := ts.GetCassandraSession()
	if err != nil {
		return -1, err
	}

	luceneQuery := querybuilder.Parse(resourcedQuery)
	if luceneQuery == "" {
		return -1, errors.New("Query is unparsable")
	}

	var count int64

	query := fmt.Sprintf(`SELECT count(logline) FROM %v WHERE lucene='{
    filter: {
        type: "boolean",
        must: [
            {type: "match", field: "cluster_id", value: %v},
            {type: "match", field: "hostname", value: %v},
            {type:"range", field:"created", lower:%v}
        ]
    },
    query: %v
}';`, ts.table, clusterID, hostname, from, luceneQuery)

	err = session.Query(query).Scan(&count)
	if err != nil {
		err = fmt.Errorf("%v. Query: %v, ClusterID: %v, From: %v, Hostname: %v", err.Error(), query, clusterID, from, hostname)
		return -1, err
	}

	return count, err
}
