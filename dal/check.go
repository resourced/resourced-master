package dal

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"
)

func NewCheck(db *sqlx.DB) *Check {
	g := &Check{}
	g.db = db
	g.table = "checks"
	g.hasID = true

	return g
}

type CheckRowsWithError struct {
	Checks []*CheckRow
	Error  error
}

type CheckRow struct {
	ID                    int64               `db:"id"`
	ClusterID             int64               `db:"cluster_id"`
	Name                  string              `db:"name"`
	Interval              string              `db:"interval"`
	IsSilenced            bool                `db:"is_silenced"`
	HostsQuery            string              `db:"hosts_query"`
	HostsList             sqlx_types.JSONText `db:"hosts_list"`
	Expressions           sqlx_types.JSONText `db:"expressions"`
	Triggers              sqlx_types.JSONText `db:"triggers"`
	LastResultHosts       sqlx_types.JSONText `db:"last_result_hosts"`
	LastResultExpressions sqlx_types.JSONText `db:"last_result_expressions"`
}

type CheckExpression struct {
	Type      string
	MinHost   int
	Metric    string
	Operator  string
	Value     float64
	PrevRange int
	PrevAggr  string
	Search    string
	Port      string
	Headers   string
	Username  string
	Password  string
	Result    bool
}

type CheckTrigger struct {
	ID                  int64
	LowViolationsCount  int64
	HighViolationsCount int64
	CreatedInterval     string
	Action              CheckTriggerAction
}

type CheckTriggerAction struct {
	Transport            string
	Email                string
	SMSPhone             string
	SMSCarrier           string
	PagerDutyServiceKey  string
	PagerDutyIncidentKey string
	PagerDutyDescription string
}

type Check struct {
	Base
}

func (a *Check) rowFromSqlResult(tx *sqlx.Tx, sqlResult sql.Result) (*CheckRow, error) {
	id, err := sqlResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	return a.GetByID(tx, id)
}

// GetByID returns one record by id.
func (a *Check) GetByID(tx *sqlx.Tx, id int64) (*CheckRow, error) {
	row := &CheckRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=$1", a.table)
	err := a.db.Get(row, query, id)

	return row, err
}

func (a *Check) Create(tx *sqlx.Tx, clusterID int64, data map[string]interface{}) (*CheckRow, error) {
	data["cluster_id"] = clusterID

	sqlResult, err := a.InsertIntoTable(tx, data)
	if err != nil {
		return nil, err
	}

	return a.rowFromSqlResult(tx, sqlResult)
}

// AllByClusterID returns all rows by cluster_id.
func (a *Check) AllByClusterID(tx *sqlx.Tx, clusterID int64) ([]*CheckRow, error) {
	rows := []*CheckRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 ORDER BY id DESC", a.table)
	err := a.db.Select(&rows, query, clusterID)

	return rows, err
}

// All returns all rows.
func (a *Check) All(tx *sqlx.Tx) ([]*CheckRow, error) {
	rows := []*CheckRow{}
	query := fmt.Sprintf("SELECT * FROM %v ORDER BY id DESC", a.table)
	err := a.db.Select(&rows, query)

	return rows, err
}

// AllSplitToDaemons returns all rows divided into daemons equally.
func (a *Check) AllSplitToDaemons(tx *sqlx.Tx, daemons []string) (map[string][]*CheckRow, error) {
	rows, err := a.All(tx)
	if err != nil {
		return nil, err
	}

	buckets := make([][]*CheckRow, len(daemons))
	for i, _ := range daemons {
		buckets[i] = make([]*CheckRow, 0)
	}

	bucketsPointer := 0
	for _, row := range rows {
		buckets[bucketsPointer] = append(buckets[bucketsPointer], row)
		bucketsPointer = bucketsPointer + 1

		if bucketsPointer >= len(buckets) {
			bucketsPointer = 0
		}
	}

	result := make(map[string][]*CheckRow)

	for i, watchersInbucket := range buckets {
		result[daemons[i]] = watchersInbucket
	}

	return result, err
}

func (a *Check) AddTrigger(tx *sqlx.Tx, checkRow *CheckRow, trigger CheckTrigger) ([]CheckTrigger, error) {
	triggers, err := checkRow.UnmarshalTriggers()
	if err != nil {
		return nil, err
	}

	triggers = append(triggers, trigger)

	triggersJSON, err := json.Marshal(triggers)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	data["triggers"] = triggersJSON

	_, err = a.UpdateByID(tx, data, checkRow.ID)

	return triggers, err
}

func (a *Check) UpdateTrigger(tx *sqlx.Tx, checkRow *CheckRow, trigger CheckTrigger) ([]CheckTrigger, error) {
	triggers, err := checkRow.UnmarshalTriggers()
	if err != nil {
		return nil, err
	}

	for i, trig := range triggers {
		if trig.ID == trigger.ID {
			triggers[i] = trigger
			break
		}
	}

	triggersJSON, err := json.Marshal(triggers)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	data["triggers"] = triggersJSON

	_, err = a.UpdateByID(tx, data, checkRow.ID)

	return triggers, err
}

func (a *Check) DeleteTrigger(tx *sqlx.Tx, checkRow *CheckRow, trigger CheckTrigger) ([]CheckTrigger, error) {
	triggers, err := checkRow.UnmarshalTriggers()
	if err != nil {
		return nil, err
	}

	newTriggers := make([]CheckTrigger, 0)

	for _, trig := range triggers {
		if trig.ID != trigger.ID {
			newTriggers = append(newTriggers, trig)
		}
	}

	triggersJSON, err := json.Marshal(newTriggers)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	data["triggers"] = triggersJSON

	_, err = a.UpdateByID(tx, data, checkRow.ID)

	return newTriggers, err
}

func (checkRow *CheckRow) GetTriggers() []CheckTrigger {
	triggers, _ := checkRow.UnmarshalTriggers()
	return triggers
}

func (checkRow *CheckRow) UnmarshalTriggers() ([]CheckTrigger, error) {
	var container []CheckTrigger

	err := json.Unmarshal(checkRow.Triggers, &container)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (checkRow *CheckRow) GetHostsList() ([]string, error) {
	var container []string

	err := json.Unmarshal(checkRow.HostsList, &container)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (checkRow *CheckRow) GetExpressions() ([]CheckExpression, error) {
	var expressions []CheckExpression

	err := json.Unmarshal(checkRow.Expressions, &expressions)
	if err != nil {
		return expressions, err
	}

	return expressions, nil
}

// EvalExpressions reduces the result of expression into a single true/false.
// 1st value: List of all CheckExpression containing results.
// 2nd value: The value of all expressions.
// 3rd value: Error
func (checkRow *CheckRow) EvalExpressions(hostDB *sqlx.DB, tsMetricDB *sqlx.DB, tsLogDB *sqlx.DB) ([]CheckExpression, bool, error) {
	var hostRows []*HostRow
	var err error

	host := NewHost(hostDB)

	if checkRow.HostsQuery != "" {
		hostRows, err = host.AllByClusterIDQueryAndUpdatedInterval(nil, checkRow.ClusterID, checkRow.HostsQuery, "5m")

	} else {
		hostnames, err := checkRow.GetHostsList()
		if err == nil && len(hostnames) > 0 {
			hostRows, err = host.AllByClusterIDAndHostnames(nil, checkRow.ClusterID, hostnames)
		}
	}

	if err != nil {
		return nil, false, err
	}

	expressions, err := checkRow.GetExpressions()
	if err != nil {
		return nil, false, err
	}

	expressionResults := make([]CheckExpression, 0)
	var finalResult bool
	var lastExpressionBooleanOperator string

	for expIndex, expression := range expressions {
		if expression.Type == "RawHostData" {
			expression = checkRow.EvalRawHostDataExpression(hostRows, expression)

		} else if expression.Type == "RelativeHostData" {
			expression = checkRow.EvalRelativeHostDataExpression(tsMetricDB, hostRows, expression)

		} else if expression.Type == "LogData" {
			expression = checkRow.EvalLogDataExpression(tsLogDB, hostRows, expression)

		} else if expression.Type == "Ping" {

		} else if expression.Type == "SSH" {

		} else if expression.Type == "HTTP" {

		} else if expression.Type == "BooleanOperator" {
			lastExpressionBooleanOperator = expression.Operator
		}

		if expIndex == 0 {
			finalResult = expression.Result

		} else {
			if lastExpressionBooleanOperator == "and" {
				finalResult = finalResult && expression.Result

			} else if lastExpressionBooleanOperator == "or" {
				finalResult = finalResult || expression.Result
			}
		}

		expressionResults = append(expressionResults, expression)
	}

	return expressionResults, finalResult, nil
}

func (checkRow *CheckRow) EvalRawHostDataExpression(hostRows []*HostRow, expression CheckExpression) CheckExpression {
	if hostRows == nil || len(hostRows) <= 0 {
		expression.Result = false
		return expression
	}

	affectedHosts := 0
	var perHostResult bool

	for _, hostRow := range hostRows {
		var val float64

		for prefix, keyAndValue := range hostRow.DataAsFlatKeyValue() {
			if !strings.HasPrefix(expression.Metric, prefix) {
				continue
			}

			for key, value := range keyAndValue {
				if strings.HasSuffix(expression.Metric, key) {
					val = value.(float64)
					break
				}
			}
		}

		if val < float64(0) {
			continue
		}

		if expression.Operator == ">" {
			perHostResult = val > expression.Value

		} else if expression.Operator == ">=" {
			perHostResult = val >= expression.Value

		} else if expression.Operator == "=" {
			perHostResult = val == expression.Value

		} else if expression.Operator == "<" {
			perHostResult = val < expression.Value

		} else if expression.Operator == "<=" {
			perHostResult = val <= expression.Value
		}

		if perHostResult {
			affectedHosts = affectedHosts + 1
		}
	}

	expression.Result = affectedHosts >= expression.MinHost

	return expression
}

func (checkRow *CheckRow) EvalRelativeHostDataExpression(tsMetricDB *sqlx.DB, hostRows []*HostRow, expression CheckExpression) CheckExpression {
	if hostRows == nil || len(hostRows) <= 0 {
		expression.Result = false
		return expression
	}

	affectedHosts := 0
	var perHostResult bool

	for _, hostRow := range hostRows {
		aggregateData, err := NewTSMetric(tsMetricDB).GetAggregateXMinutesByHostnameAndKey(nil, checkRow.ClusterID, expression.PrevRange, hostRow.Hostname, expression.Metric)
		if err != nil {
			continue
		}

		var val float64

		for prefix, keyAndValue := range hostRow.DataAsFlatKeyValue() {
			if !strings.HasPrefix(expression.Metric, prefix) {
				continue
			}

			for key, value := range keyAndValue {
				if strings.HasSuffix(expression.Metric, key) {
					val = value.(float64)
					break
				}
			}
		}

		if val < float64(0) {
			continue
		}

		var prevVal float64

		if expression.PrevAggr == "avg" {
			prevVal = aggregateData.Avg
		} else if expression.PrevAggr == "max" {
			prevVal = aggregateData.Max
		} else if expression.PrevAggr == "min" {
			prevVal = aggregateData.Min
		} else if expression.PrevAggr == "sum" {
			prevVal = aggregateData.Sum
		}

		if prevVal < float64(0) {
			continue
		}

		valPercentage := (val / prevVal) * float64(100)

		if expression.Operator == ">" {
			perHostResult = valPercentage > expression.Value

		} else if expression.Operator == "<" {
			perHostResult = valPercentage < expression.Value
		}

		if perHostResult {
			affectedHosts = affectedHosts + 1
		}
	}

	expression.Result = affectedHosts >= expression.MinHost

	return expression
}

func (checkRow *CheckRow) EvalLogDataExpression(tsLogDB *sqlx.DB, hostRows []*HostRow, expression CheckExpression) CheckExpression {
	hostnames, err := checkRow.GetHostsList()
	if err != nil {
		expression.Result = false
		return expression
	}

	if len(hostnames) == 0 && hostRows != nil && len(hostRows) > 0 {
		hostnames = make([]string, len(hostRows))

		for i, hostRow := range hostRows {
			hostnames[i] = hostRow.Hostname
		}
	}

	if hostnames == nil || len(hostnames) <= 0 {
		expression.Result = false
		return expression
	}

	affectedHosts := 0
	var perHostResult bool

	for _, hostname := range hostnames {
		println("hostname")
		println(hostname)

		now := time.Now().UTC()
		from := now.Add(-1 * time.Duration(expression.PrevRange) * time.Minute).UTC().Unix()
		to := now.Unix()
		searchQuery := fmt.Sprintf(`logline search "%v"`, expression.Search)

		println("time stuff:")
		println(from)
		println(to)

		valInt64, err := NewTSLog(tsLogDB).CountByClusterIDRangeHostAndQuery(nil, checkRow.ClusterID, from, to, hostname, searchQuery)
		if err != nil {
			println(err.Error())
			continue
		}

		val := float64(valInt64)

		println("val")
		println(val)
		println(expression.Value)

		if val < float64(0) {
			continue
		}

		if expression.Operator == ">" {
			perHostResult = val > expression.Value

		} else if expression.Operator == "<" {
			perHostResult = val < expression.Value
		}

		if perHostResult {
			affectedHosts = affectedHosts + 1
		}
	}

	expression.Result = affectedHosts >= expression.MinHost

	return expression
}
