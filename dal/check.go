package dal

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"
	"github.com/marcw/pagerduty"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/libstring"
	"github.com/resourced/resourced-master/mailer"
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
	Type       string
	MinHost    int
	Metric     string
	Operator   string
	Value      float64
	PrevRange  int
	PrevAggr   string
	Search     string
	Protocol   string
	Port       string
	Headers    string
	Username   string
	Password   string
	HTTPMethod string
	HTTPBody   string
	Result     struct {
		Value         bool
		Message       string
		BadHostnames  []string
		GoodHostnames []string
	}
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
	var triggers []CheckTrigger

	err := json.Unmarshal(checkRow.Triggers, &triggers)
	if err != nil {
		return nil, err
	}

	return triggers, nil
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
			expression = checkRow.EvalPingExpression(hostRows, expression)

		} else if expression.Type == "SSH" {
			expression = checkRow.EvalSSHExpression(hostRows, expression)

		} else if expression.Type == "HTTP" || expression.Type == "HTTPS" {
			expression = checkRow.EvalHTTPExpression(hostRows, expression)

		} else if expression.Type == "BooleanOperator" {
			lastExpressionBooleanOperator = expression.Operator
		}

		if expIndex == 0 {
			finalResult = expression.Result.Value

		} else {
			if lastExpressionBooleanOperator == "and" {
				finalResult = finalResult && expression.Result.Value

			} else if lastExpressionBooleanOperator == "or" {
				finalResult = finalResult || expression.Result.Value
			}
		}

		expressionResults = append(expressionResults, expression)
	}

	return expressionResults, finalResult, nil
}

func (checkRow *CheckRow) EvalRawHostDataExpression(hostRows []*HostRow, expression CheckExpression) CheckExpression {
	if hostRows == nil || len(hostRows) <= 0 {
		expression.Result.Value = true
		expression.Result.Message = "There are no hosts to check"
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

		// If a Host does not contain a particular metric,
		// We assume that there's something wrong with it.
		if val < float64(0) {
			perHostResult = true
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

	expression.Result.Value = affectedHosts >= expression.MinHost

	return expression
}

func (checkRow *CheckRow) EvalRelativeHostDataExpression(tsMetricDB *sqlx.DB, hostRows []*HostRow, expression CheckExpression) CheckExpression {
	if hostRows == nil || len(hostRows) <= 0 {
		expression.Result.Value = true
		expression.Result.Message = "There are no hosts to check"
		return expression
	}

	affectedHosts := 0
	var perHostResult bool

	for _, hostRow := range hostRows {

		aggregateData, err := NewTSMetric(tsMetricDB).GetAggregateXMinutesByHostnameAndKey(nil, checkRow.ClusterID, expression.PrevRange, hostRow.Hostname, expression.Metric)
		if err != nil {
			// If a Host does not contain historical data of a particular metric,
			// We assume that there's something wrong with it.
			if strings.Contains(err.Error(), "no rows in result set") {
				perHostResult = true
				affectedHosts = affectedHosts + 1
			}
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

		// If a Host does not contain a particular metric,
		// We assume that there's something wrong with it.
		if val < float64(0) {
			perHostResult = true
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

	expression.Result.Value = affectedHosts >= expression.MinHost

	return expression
}

func (checkRow *CheckRow) EvalLogDataExpression(tsLogDB *sqlx.DB, hostRows []*HostRow, expression CheckExpression) CheckExpression {
	hostnames, err := checkRow.GetHostsList()
	if err != nil {
		expression.Result.Value = false
		return expression
	}

	if len(hostnames) == 0 && hostRows != nil && len(hostRows) > 0 {
		hostnames = make([]string, len(hostRows))

		for i, hostRow := range hostRows {
			hostnames[i] = hostRow.Hostname
		}
	}

	if hostnames == nil || len(hostnames) <= 0 {
		expression.Result.Value = false
		return expression
	}

	affectedHosts := 0
	var perHostResult bool

	for _, hostname := range hostnames {
		now := time.Now().UTC()
		from := now.Add(-1 * time.Duration(expression.PrevRange) * time.Minute).UTC().Unix()
		searchQuery := fmt.Sprintf(`logline search "%v"`, expression.Search)

		valInt64, err := NewTSLog(tsLogDB).CountByClusterIDFromTimestampHostAndQuery(nil, checkRow.ClusterID, from, hostname, searchQuery)
		if err != nil {
			continue
		}

		println(valInt64)

		val := float64(valInt64)

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

	expression.Result.Value = affectedHosts >= expression.MinHost

	return expression
}

func (checkRow *CheckRow) CheckPing(hostname string) (outBytes []byte, err error) {
	return exec.Command("ping", "-c", "1", hostname).CombinedOutput()
}

func (checkRow *CheckRow) EvalPingExpression(hostRows []*HostRow, expression CheckExpression) CheckExpression {
	hostnames, err := checkRow.GetHostsList()
	if err != nil {
		expression.Result.Value = false
		return expression
	}

	if len(hostnames) == 0 && hostRows != nil && len(hostRows) > 0 {
		hostnames = make([]string, len(hostRows))

		for i, hostRow := range hostRows {
			hostnames[i] = hostRow.Hostname
		}
	}

	if hostnames == nil || len(hostnames) <= 0 {
		expression.Result.Value = false
		return expression
	}

	affectedHosts := 0

	for _, hostname := range hostnames {
		outputBytes, err := checkRow.CheckPing(hostname)

		println(string(outputBytes))

		if err != nil {
			println(err.Error())
			affectedHosts = affectedHosts + 1
		}
	}

	expression.Result.Value = affectedHosts >= expression.MinHost

	return expression
}

func (checkRow *CheckRow) CheckSSH(hostname, port, user string) (outBytes []byte, err error) {
	sshOptions := []string{"-o BatchMode=yes", "-o ConnectTimeout=10"}

	if port != "" {
		sshOptions = append(sshOptions, []string{"-p", port}...)
	}

	userAtHost := hostname

	if user != "" {
		userAtHost = fmt.Sprintf("%v@%v", user, hostname)
	}

	sshOptions = append(sshOptions, userAtHost)

	return exec.Command("ssh", sshOptions...).CombinedOutput()
}

func (checkRow *CheckRow) EvalSSHExpression(hostRows []*HostRow, expression CheckExpression) CheckExpression {
	hostnames, err := checkRow.GetHostsList()
	if err != nil {
		expression.Result.Value = false
		return expression
	}

	if len(hostnames) == 0 && hostRows != nil && len(hostRows) > 0 {
		hostnames = make([]string, len(hostRows))

		for i, hostRow := range hostRows {
			hostnames[i] = hostRow.Hostname
		}
	}

	if hostnames == nil || len(hostnames) <= 0 {
		expression.Result.Value = false
		return expression
	}

	affectedHosts := 0

	for _, hostname := range hostnames {
		outputBytes, err := checkRow.CheckSSH(hostname, expression.Port, expression.Username)
		outputString := string(outputBytes)

		if err != nil && !strings.Contains(outputString, "Permission denied") && !strings.Contains(outputString, "Host key verification failed") {
			println(err.Error())
			affectedHosts = affectedHosts + 1
		}
	}

	expression.Result.Value = affectedHosts >= expression.MinHost

	return expression
}

func (checkRow *CheckRow) CheckHTTP(hostname, scheme, port, method, user, pass string, headers map[string]string, body string) (resp *http.Response, err error) {
	url := fmt.Sprintf("%v://%v:%s", scheme, hostname, port)

	client := &http.Client{}

	var req *http.Request

	if body != "" {
		req, err = http.NewRequest(strings.ToUpper(method), url, strings.NewReader(body))

		// Detect if POST body is JSON and set content-type
		if strings.HasPrefix(body, "{") || strings.HasPrefix(body, "[") {
			req.Header.Set("Content-Type", "application/json")
		} else {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}

	} else {
		req, err = http.NewRequest(strings.ToUpper(method), url, nil)
	}

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Method":     "http.NewRequest",
			"URL":        url,
			"HTTPMethod": method,
		}).Error(err)
		return nil, err
	}

	for headerKey, headerVal := range headers {
		req.Header.Add(headerKey, headerVal)
	}

	if user != "" || pass != "" {
		req.SetBasicAuth(user, pass)
	}

	resp, err = client.Do(req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Method":     "http.Client{}.Do",
			"URL":        url,
			"HTTPMethod": method,
		}).Error(err)
		return nil, err

	} else if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}

	return resp, err
}

func (checkRow *CheckRow) EvalHTTPExpression(hostRows []*HostRow, expression CheckExpression) CheckExpression {
	hostnames, err := checkRow.GetHostsList()
	if err != nil {
		expression.Result.Value = false
		return expression
	}

	if len(hostnames) == 0 && hostRows != nil && len(hostRows) > 0 {
		hostnames = make([]string, len(hostRows))

		for i, hostRow := range hostRows {
			hostnames[i] = hostRow.Hostname
		}
	}

	if hostnames == nil || len(hostnames) <= 0 {
		expression.Result.Value = false
		return expression
	}

	affectedHosts := 0

	for _, hostname := range hostnames {
		headers := make(map[string]string)

		for _, headersNewLine := range strings.Split(expression.Headers, "\n") {
			for _, kvString := range strings.Split(headersNewLine, ",") {
				if strings.Contains(kvString, "=") {
					kvSlice := strings.Split(kvString, "=")
					if len(kvSlice) >= 2 {
						headers[strings.TrimSpace(kvSlice[0])] = strings.TrimSpace(kvSlice[1])
					}
				}
			}
		}

		resp, err := checkRow.CheckHTTP(hostname, expression.Protocol, expression.Port, expression.HTTPMethod, expression.Username, expression.Password, headers, expression.HTTPBody)
		if err != nil || (resp != nil && resp.StatusCode != 200) {
			affectedHosts = affectedHosts + 1
		}
	}

	expression.Result.Value = affectedHosts >= expression.MinHost

	return expression
}

func (checkRow *CheckRow) RunTriggers(appConfig config.GeneralConfig, tsCheckDB *sqlx.DB, mailr *mailer.Mailer) error {
	if checkRow.IsSilenced {
		return nil
	}

	triggers, err := checkRow.UnmarshalTriggers()
	if err != nil {
		logrus.Error(err)
		return err
	}

	for _, trigger := range triggers {
		tsCheckRows, err := NewTSCheck(tsCheckDB).AllViolationsByClusterIDCheckIDAndInterval(nil, checkRow.ClusterID, checkRow.ID, trigger.CreatedInterval)
		if err != nil {
			return err
		}

		if len(tsCheckRows) == 0 {
			continue
		}

		lastViolation := tsCheckRows[0]
		violationsCount := len(tsCheckRows)

		println("violationsCount")
		println(violationsCount)
		println(trigger.LowViolationsCount)
		println(trigger.HighViolationsCount)

		if int64(violationsCount) >= trigger.LowViolationsCount && int64(violationsCount) <= trigger.HighViolationsCount {
			if trigger.Action.Transport == "nothing" {
				// Do nothing

			} else if trigger.Action.Transport == "email" {
				println("I should be here")
				err = checkRow.RunEmailTrigger(trigger, lastViolation, violationsCount, mailr, appConfig)
				if err != nil {
					continue
				}

			} else if trigger.Action.Transport == "sms" {
				err = checkRow.RunSMSTrigger(trigger, lastViolation, violationsCount, mailr, appConfig)
				if err != nil {
					continue
				}

			} else if trigger.Action.Transport == "pagerduty" {
				err = checkRow.RunPagerDutyTrigger(trigger, lastViolation)
				if err != nil {
					logrus.Error(err)
					continue
				}
			}
		}
	}

	return nil
}

func (checkRow *CheckRow) RunEmailTrigger(trigger CheckTrigger, lastViolation *TSCheckRow, violationsCount int, mailr *mailer.Mailer, appConfig config.GeneralConfig) (err error) {
	if trigger.Action.Email == "" {
		return fmt.Errorf("Unable to send email because trigger.Action.Email is empty")
	}

	to := trigger.Action.Email
	subject := fmt.Sprintf(`%v Check(ID: %v): %v, failed %v times`, appConfig.Checks.Email.SubjectPrefix, checkRow.ID, checkRow.Name, violationsCount)
	body := ""

	if lastViolation != nil {
		bodyBytes, err := libstring.PrettyPrintJSON([]byte(lastViolation.Expressions.String()))
		if err != nil {
			return fmt.Errorf("Unable to send email because of malformed check expressions")
		}

		body = string(bodyBytes)
	}

	err = mailr.Send(to, subject, body)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Method":      "checkRow.RunEmailTrigger",
			"Transport":   trigger.Action.Transport,
			"HostAndPort": mailr.HostAndPort,
			"From":        mailr.From,
			"To":          to,
			"Subject":     subject,
		}).Error(err)
	}

	return err
}

func (checkRow *CheckRow) RunSMSTrigger(trigger CheckTrigger, lastViolation *TSCheckRow, violationsCount int, mailr *mailer.Mailer, appConfig config.GeneralConfig) (err error) {
	carrier := strings.ToLower(trigger.Action.SMSCarrier)

	gateway, ok := appConfig.Checks.SMSEmailGateway[carrier]
	if !ok {
		return fmt.Errorf("Unable to lookup SMS Gateway for carrier: %v", carrier)
	}

	flattenPhone := libstring.FlattenPhone(trigger.Action.SMSPhone)
	if len(flattenPhone) != 10 {
		logrus.Warningf("Length of phone number is not 10. Flatten phone number: %v. Length: %v", flattenPhone, len(flattenPhone))
		return nil
	}

	to := fmt.Sprintf("%v@%v", flattenPhone, gateway)
	subject := ""
	body := fmt.Sprintf(`%v Check(ID: %v): %v, failed %v times`, appConfig.Checks.Email.SubjectPrefix, checkRow.ID, checkRow.Name, violationsCount)

	err = mailr.Send(to, subject, body)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Method":      "checkRow.RunSMSTrigger",
			"Transport":   trigger.Action.Transport,
			"HostAndPort": mailr.HostAndPort,
			"From":        mailr.From,
			"To":          to,
			"Subject":     subject,
		}).Error(err)
	}

	return err
}

func (checkRow *CheckRow) RunPagerDutyTrigger(trigger CheckTrigger, lastViolation *TSCheckRow) (err error) {
	// Create a new PD "trigger" event
	event := pagerduty.NewTriggerEvent(trigger.Action.PagerDutyServiceKey, trigger.Action.PagerDutyDescription)

	// Add details to PD event
	if lastViolation != nil {
		err = lastViolation.Expressions.Unmarshal(&event.Details)
		if err != nil {
			return err
		}
	}

	hostname, _ := os.Hostname()

	// Add Client to PD event
	event.Client = fmt.Sprintf("ResourceD Master on: %v", hostname)

	// Submit PD event
	pdResponse, _, err := pagerduty.Submit(event)
	if err != nil {
		return err
	}
	if pdResponse == nil {
		return nil
	}

	// Update incident key into watchers_triggers row
	trigger.Action.PagerDutyIncidentKey = pdResponse.IncidentKey
	// TODO: How do I save this?

	return err
}
