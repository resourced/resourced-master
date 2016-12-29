package cassandra

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gocql/gocql"

	"github.com/resourced/resourced-master/contexthelper"
	"github.com/resourced/resourced-master/models/shared"
)

func NewTSEvent(ctx context.Context) *TSEvent {
	ts := &TSEvent{}
	ts.AppContext = ctx
	ts.table = "ts_events"

	return ts
}

type TSEventRow struct {
	ID          int64  `db:"id"`
	ClusterID   int64  `db:"cluster_id"`
	CreatedFrom int64  `db:"created_from"`
	CreatedTo   int64  `db:"created_to"`
	Description string `db:"description"`
}

type TSEvent struct {
	Base
}

func (ts *TSEvent) GetCassandraSession() (*gocql.Session, error) {
	cassandradbs, err := contexthelper.GetCassandraDBConfig(ts.AppContext)
	if err != nil {
		return nil, err
	}

	return cassandradbs.TSEventSession, nil
}

// AllLinesByClusterIDAndCreatedFromRangeForHighchart returns all rows given created_from range.
func (ts *TSEvent) AllLinesByClusterIDAndCreatedFromRangeForHighchart(clusterID, from, to int64) ([]shared.TSEventHighchartLinePayload, error) {
	session, err := ts.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	rows := []*TSEventRow{}
	query := fmt.Sprintf(`SELECT id, cluster_id, created_from, created_to, description FROM %v WHERE cluster_id=? AND created_from >= ? AND created_from <= ?`, ts.table)

	var scannedID, scannedClusterID, scannedCreatedFrom, scannedCreatedTo int64
	var scannedDescription string

	iter := session.Query(query, clusterID, from, to).Iter()
	for iter.Scan(&scannedID, &scannedClusterID, &scannedCreatedFrom, &scannedCreatedTo, &scannedDescription) {
		rows = append(rows, &TSEventRow{
			ID:          scannedID,
			ClusterID:   scannedClusterID,
			CreatedFrom: scannedCreatedFrom,
			CreatedTo:   scannedCreatedTo,
			Description: scannedDescription,
		})
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{
			"Method":    "TSMetric.AllLinesByClusterIDAndCreatedFromRangeForHighchart",
			"ClusterID": clusterID,
			"From":      from,
			"To":        to,
		}).Error(err)

		return nil, err
	}

	hcRows := make([]shared.TSEventHighchartLinePayload, len(rows))

	for i, row := range rows {
		if row.CreatedFrom == row.CreatedTo {
			hcRow := shared.TSEventHighchartLinePayload{}
			hcRow.ID = row.ID
			hcRow.CreatedFrom = row.CreatedFrom * 1000
			hcRow.CreatedTo = row.CreatedTo * 1000
			hcRow.Description = row.Description

			hcRows[i] = hcRow
		}
	}

	return hcRows, err
}

// AllBandsByClusterIDAndCreatedFromRangeForHighchart returns all rows with time stretch between created_from and created_to.
func (ts *TSEvent) AllBandsByClusterIDAndCreatedFromRangeForHighchart(clusterID, from, to int64) ([]shared.TSEventHighchartLinePayload, error) {
	session, err := ts.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	// id bigint,
	// cluster_id bigint,
	// created_from bigint,
	// created_to bigint,
	// description text,

	rows := []*TSEventRow{}
	query := fmt.Sprintf(`SELECT id, cluster_id, created_from, created_to, description FROM %v WHERE cluster_id=? AND created_from >= ? AND created_from <= ?`, ts.table)

	var scannedID, scannedClusterID, scannedCreatedFrom, scannedCreatedTo int64
	var scannedDescription string

	iter := session.Query(query, clusterID, from, to).Iter()
	for iter.Scan(&scannedID, &scannedClusterID, &scannedCreatedFrom, &scannedCreatedTo, &scannedDescription) {
		rows = append(rows, &TSEventRow{
			ID:          scannedID,
			ClusterID:   scannedClusterID,
			CreatedFrom: scannedCreatedFrom,
			CreatedTo:   scannedCreatedTo,
			Description: scannedDescription,
		})
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{
			"Method":    "TSEvent.AllLinesByClusterIDAndCreatedFromRangeForHighchart",
			"ClusterID": clusterID,
			"From":      from,
			"To":        to,
		}).Error(err)

		return nil, err
	}
	hcRows := make([]shared.TSEventHighchartLinePayload, len(rows))

	for i, row := range rows {
		if row.CreatedFrom < row.CreatedTo {
			hcRow := shared.TSEventHighchartLinePayload{}
			hcRow.ID = row.ID
			hcRow.CreatedFrom = row.CreatedFrom * 1000
			hcRow.CreatedTo = row.CreatedTo * 1000
			hcRow.Description = row.Description

			hcRows[i] = hcRow
		}
	}

	return hcRows, err
}

// GetByClusterIDAndID returns record by cluster_id and id.
func (ts *TSEvent) GetByClusterIDAndID(clusterID, id int64) (*TSEventRow, error) {
	session, err := ts.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	row := &TSEventRow{}
	query := fmt.Sprintf(`SELECT id, cluster_id, created_from, created_to, description FROM %v WHERE cluster_id=? AND id=? LIMIT 1 ALLOW FILTERING`, ts.table)

	var scannedID, scannedClusterID, scannedCreatedFrom, scannedCreatedTo int64
	var scannedDescription string

	iter := session.Query(query, clusterID, id).Iter()
	for iter.Scan(&scannedID, &scannedClusterID, &scannedCreatedFrom, &scannedCreatedTo, &scannedDescription) {
		row = &TSEventRow{
			ID:          scannedID,
			ClusterID:   scannedClusterID,
			CreatedFrom: scannedCreatedFrom,
			CreatedTo:   scannedCreatedTo,
			Description: scannedDescription,
		}
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{
			"Method":      "TSEvent.GetByID",
			"ID":          scannedID,
			"ClusterID":   scannedClusterID,
			"CreatedFrom": scannedCreatedFrom,
			"CreatedTo":   scannedCreatedTo,
			"Description": scannedDescription,
		}).Error(err)

		return nil, err
	}

	return row, nil
}

// GetByID returns record by id.
func (ts *TSEvent) GetByID(id int64) (*TSEventRow, error) {
	session, err := ts.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	row := &TSEventRow{}
	query := fmt.Sprintf(`SELECT id, cluster_id, created_from, created_to, description FROM %v WHERE id=? LIMIT 1`, ts.table)

	var scannedID, scannedClusterID, scannedCreatedFrom, scannedCreatedTo int64
	var scannedDescription string

	iter := session.Query(query, id).Iter()
	for iter.Scan(&scannedID, &scannedClusterID, &scannedCreatedFrom, &scannedCreatedTo, &scannedDescription) {
		row = &TSEventRow{
			ID:          scannedID,
			ClusterID:   scannedClusterID,
			CreatedFrom: scannedCreatedFrom,
			CreatedTo:   scannedCreatedTo,
			Description: scannedDescription,
		}
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{
			"Method":      "TSEvent.GetByID",
			"ID":          scannedID,
			"ClusterID":   scannedClusterID,
			"CreatedFrom": scannedCreatedFrom,
			"CreatedTo":   scannedCreatedTo,
			"Description": scannedDescription,
		}).Error(err)

		return nil, err
	}

	return row, nil
}

func (ts *TSEvent) CreateFromJSON(id, clusterID int64, jsonData []byte, ttl time.Duration) (*TSEventRow, error) {
	session, err := ts.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	payload := &shared.TSEventCreatePayload{}

	err = json.Unmarshal(jsonData, payload)
	if err != nil {
		return nil, err
	}

	err = session.Query(
		fmt.Sprintf(`INSERT INTO %v (id, cluster_id, created_from, created_to, description) VALUES (?, ?, ?, ?, ?) USING TTL ?`, ts.table),
		id,
		clusterID,
		payload.From,
		payload.To,
		payload.Description,
		ttl,
	).Exec()

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Method":    "TSEvent.CreateFromJSON",
			"ID":        id,
			"ClusterID": clusterID,
			"From":      payload.From,
			"To":        payload.To,
		}).Error(err)

		return nil, err
	}

	return ts.GetByClusterIDAndID(clusterID, id)
}

func (ts *TSEvent) DeleteByClusterIDAndID(clusterID, id int64) (err error) {
	session, err := ts.GetCassandraSession()
	if err != nil {
		return err
	}

	query := fmt.Sprintf("DELETE FROM %v WHERE id=? AND cluster_id=?", ts.table)

	logrus.WithFields(logrus.Fields{
		"Method": "TSEvent.DeleteByClusterIDAndID",
		"Query":  query,
	}).Info("Delete Query")

	return session.Query(query, id, clusterID).Exec()
}
