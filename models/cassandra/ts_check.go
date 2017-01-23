package cassandra

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gocql/gocql"

	"github.com/resourced/resourced-master/contexthelper"
)

func NewTSCheck(ctx context.Context) *TSCheck {
	ts := &TSCheck{}
	ts.AppContext = ctx
	ts.table = "ts_checks"

	return ts
}

type TSCheckRow struct {
	ClusterID   int64  `db:"cluster_id"`
	CheckID     int64  `db:"check_id"`
	Created     int64  `db:"created"`
	Result      bool   `db:"result"`
	Expressions string `db:"expressions"`
}

func (tsCheckRow *TSCheckRow) GetExpressionsWithoutError() []CheckExpression {
	var expressions []CheckExpression

	json.Unmarshal([]byte(tsCheckRow.Expressions), &expressions)

	return expressions
}

type TSCheck struct {
	Base
}

func (ts *TSCheck) GetCassandraSession() (*gocql.Session, error) {
	cassandradbs, err := contexthelper.GetCassandraDBConfig(ts.AppContext)
	if err != nil {
		return nil, err
	}

	return cassandradbs.TSCheckSession, nil
}

// LastByClusterIDCheckIDAndLimit returns a row by cluster_id, check_id and result.
func (ts *TSCheck) LastByClusterIDCheckIDAndLimit(clusterID, checkID, limit int64) ([]*TSCheckRow, error) {
	session, err := ts.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	rows := []*TSCheckRow{}

	query := fmt.Sprintf("SELECT cluster_id, check_id, created, result, expressions FROM %v WHERE cluster_id=? AND check_id=? ORDER BY check_id,created DESC LIMIT ? ALLOW FILTERING", ts.table)

	var scannedClusterID, scannedCheckID, scannedCreated int64
	var scannedResult bool
	var scannedExpressions string

	iter := session.Query(query, clusterID, checkID, limit).Iter()
	for iter.Scan(&scannedClusterID, &scannedCheckID, &scannedCreated, &scannedResult, &scannedExpressions) {
		rows = append(rows, &TSCheckRow{
			ClusterID:   scannedClusterID,
			CheckID:     scannedCheckID,
			Created:     scannedCreated,
			Result:      scannedResult,
			Expressions: scannedExpressions,
		})
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{
			"Method":      "TSCheck.LastByClusterIDCheckIDAndLimit",
			"ClusterID":   scannedClusterID,
			"CheckID":     scannedCheckID,
			"Created":     scannedCreated,
			"Result":      scannedResult,
			"Expressions": scannedExpressions,
		}).Error(err)

		return nil, err
	}

	return rows, nil
}

// LastByClusterIDCheckIDAndResult returns a row by cluster_id, check_id and result.
func (ts *TSCheck) LastByClusterIDCheckIDAndResult(clusterID, checkID int64, result bool) (*TSCheckRow, error) {
	session, err := ts.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	row := &TSCheckRow{}

	query := fmt.Sprintf("SELECT cluster_id, check_id, created, result, expressions FROM %v WHERE cluster_id=? AND check_id=? AND result=? ORDER BY check_id,created DESC LIMIT 1 ALLOW FILTERING", ts.table)

	var scannedClusterID, scannedCheckID, scannedCreated int64
	var scannedResult bool
	var scannedExpressions string

	iter := session.Query(query, clusterID, checkID, result).Iter()
	for iter.Scan(&scannedClusterID, &scannedCheckID, &scannedCreated, &scannedResult, &scannedExpressions) {
		row = &TSCheckRow{
			ClusterID:   scannedClusterID,
			CheckID:     scannedCheckID,
			Created:     scannedCreated,
			Result:      scannedResult,
			Expressions: scannedExpressions,
		}
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{
			"Method":    "TSCheck.LastByClusterIDCheckIDAndResult",
			"ClusterID": scannedClusterID,
			"CheckID":   scannedCheckID,
			"Result":    scannedResult,
		}).Error(err)

		return nil, err
	}

	return row, nil
}

// AllViolationsByClusterIDCheckIDAndInterval returns all ts_checks rows since last good marker.
func (ts *TSCheck) AllViolationsByClusterIDCheckIDAndInterval(clusterID, checkID, createdIntervalMinute int64) ([]*TSCheckRow, error) {
	session, err := ts.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	lastGoodOne, err := ts.LastByClusterIDCheckIDAndResult(clusterID, checkID, false)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			lastGoodOne = nil
		} else {
			return nil, err
		}
	}

	now := time.Now().UTC()
	from := now.Add(-1 * time.Minute * time.Duration(createdIntervalMinute)).UTC().Unix()

	rows := []*TSCheckRow{}

	var scannedClusterID, scannedCheckID, scannedCreated int64
	var scannedResult bool
	var scannedExpressions string

	var query string

	if lastGoodOne != nil {
		query = fmt.Sprintf(`SELECT cluster_id, check_id, created, result, expressions FROM %v WHERE cluster_id=? AND check_id=? AND created > ? AND result = ?
ORDER BY check_id,created DESC ALLOW FILTERING`, ts.table)

	} else {
		query = fmt.Sprintf(`SELECT * FROM %v WHERE cluster_id=? AND check_id=? AND created > ? AND result = ? AND ORDER BY check_id,created DESC ALLOW FILTERING`, ts.table)

		from := math.Max(float64(lastGoodOne.Created), float64(from))

		iter := session.Query(query, clusterID, checkID, from, true).Iter()
		for iter.Scan(&scannedClusterID, &scannedCheckID, &scannedCreated, &scannedResult, &scannedExpressions) {
			rows = append(rows, &TSCheckRow{
				ClusterID:   scannedClusterID,
				CheckID:     scannedCheckID,
				Created:     scannedCreated,
				Result:      scannedResult,
				Expressions: scannedExpressions,
			})
		}
		if err := iter.Close(); err != nil {
			err = fmt.Errorf("%v. Query: %v", err.Error(), query)
			logrus.WithFields(logrus.Fields{
				"Method":                "TSCheck.AllViolationsByClusterIDCheckIDAndInterval",
				"ClusterID":             scannedClusterID,
				"CheckID":               scannedCheckID,
				"CreatedIntervalMinute": createdIntervalMinute,
			}).Error(err)

			return nil, err
		}
	}

	return rows, err
}

// Create a new record.
func (ts *TSCheck) Create(clusterID, checkID int64, result bool, expressions []CheckExpression, ttl time.Duration) error {
	expressionsJSON, err := json.Marshal(expressions)
	if err != nil {
		return err
	}

	session, err := ts.GetCassandraSession()
	if err != nil {
		return err
	}

	err = session.Query(
		fmt.Sprintf(`INSERT INTO %v (cluster_id, check_id, created, result, expressions) VALUES (?, ?, ?, ?, ?) USING TTL ?`, ts.table),
		clusterID,
		checkID,
		time.Now().UTC().Unix(),
		result,
		string(expressionsJSON),
		ttl,
	).Exec()

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Method":      "TSCheck.Create",
			"ClusterID":   clusterID,
			"CheckID":     checkID,
			"Result":      result,
			"Expressions": string(expressionsJSON),
		}).Error(err)

		return err
	}

	return nil
}
