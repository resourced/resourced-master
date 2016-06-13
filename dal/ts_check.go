package dal

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"
)

func NewTSCheck(db *sqlx.DB) *TSCheck {
	ts := &TSCheck{}
	ts.db = db
	ts.table = "ts_checks"

	return ts
}

type TSCheckRow struct {
	ClusterID   int64               `db:"cluster_id"`
	CheckID     int64               `db:"check_id"`
	Created     time.Time           `db:"created"`
	Deleted     time.Time           `db:"deleted"`
	Result      bool                `db:"result"`
	Expressions sqlx_types.JSONText `db:"expressions"`
}

func (tsCheckRow *TSCheckRow) GetExpressionsWithoutError() []CheckExpression {
	var expressions []CheckExpression

	json.Unmarshal(tsCheckRow.Expressions, &expressions)

	return expressions
}

type TSCheck struct {
	TSBase
}

// LastByClusterIDCheckIDAndResult returns a row by cluster_id, check_id and result.
func (ts *TSCheck) LastByClusterIDCheckIDAndAffectedHosts(tx *sqlx.Tx, clusterID, CheckID int64, result bool) (*TSCheckRow, error) {
	row := &TSCheckRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND check_id=$2 AND result=$3 ORDER BY cluster_id,check_id,created DESC LIMIT 1", ts.table)
	err := ts.db.Get(row, query, clusterID, CheckID, result)

	logrus.WithFields(logrus.Fields{
		"Method":    "TSCheck.LastByClusterIDCheckIDAndAffectedHosts",
		"ClusterID": clusterID,
		"CheckID":   CheckID,
		"Result":    result,
		"Query":     query,
	}).Info("Select Query")

	return row, err
}

// AllViolationsByClusterIDCheckIDAndInterval returns all ts_checks rows since last good marker.
func (ts *TSCheck) AllViolationsByClusterIDCheckIDAndInterval(tx *sqlx.Tx, clusterID, CheckID, createdIntervalMinute, deletedFrom int64) ([]*TSCheckRow, error) {
	lastGoodOne, err := ts.LastByClusterIDCheckIDAndAffectedHosts(tx, clusterID, CheckID, false)
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

	if lastGoodOne != nil {
		query := fmt.Sprintf(`SELECT * FROM %v WHERE cluster_id=$1 AND
check_id=$2 AND
created > GREATEST($3, to_timestamp($4) at time zone 'utc') AND
result = $5 AND
deleted >= to_timestamp($6) at time zone 'utc'
ORDER BY cluster_id,check_id,created DESC`, ts.table)

		err = ts.db.Select(&rows, query, clusterID, CheckID, lastGoodOne.Created.UTC(), from, true, deletedFrom)

	} else {
		query := fmt.Sprintf(`SELECT * FROM %v WHERE cluster_id=$1 AND
check_id=$2 AND
created > to_timestamp($3) at time zone 'utc' AND
result = $4 AND
deleted >= to_timestamp($5) at time zone 'utc'
ORDER BY cluster_id,check_id,created DESC`, ts.table)

		err = ts.db.Select(&rows, query, clusterID, CheckID, from, true, deletedFrom)
	}

	return rows, err
}

// Create a new record.
func (ts *TSCheck) Create(tx *sqlx.Tx, clusterID, CheckID int64, result bool, expressions []CheckExpression, deletedFrom int64) error {
	expressionsJSON, err := json.Marshal(expressions)
	if err != nil {
		return err
	}

	insertData := make(map[string]interface{})
	insertData["cluster_id"] = clusterID
	insertData["check_id"] = CheckID
	insertData["result"] = result
	insertData["expressions"] = expressionsJSON
	insertData["deleted"] = time.Unix(deletedFrom, 0).UTC()

	_, err = ts.InsertIntoTable(tx, insertData)
	return err
}
