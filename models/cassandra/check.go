package cassandra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/Sirupsen/logrus"
	"github.com/marcw/pagerduty"

	"github.com/resourced/resourced-master/contexthelper"
	"github.com/resourced/resourced-master/libstring"
)

func NewCheck(ctx context.Context) *Check {
	c := &Check{}
	c.AppContext = ctx
	c.table = "checks"

	return c
}

type CheckRowsWithError struct {
	Checks []*CheckRow
	Error  error
}

type CheckRow struct {
	ID                    int64  `db:"id"`
	ClusterID             int64  `db:"cluster_id"`
	Name                  string `db:"name"`
	Interval              string `db:"interval"`
	IsSilenced            bool   `db:"is_silenced"`
	HostsQuery            string `db:"hosts_query"`
	HostsList             string `db:"hosts_list"`
	Expressions           string `db:"expressions"`
	Triggers              string `db:"triggers"`
	LastResultHosts       string `db:"last_result_hosts"`
	LastResultExpressions string `db:"last_result_expressions"`
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
	ID                    int64
	LowViolationsCount    int64
	HighViolationsCount   int64
	CreatedIntervalMinute int64
	Action                CheckTriggerAction
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

// GetByID returns one record by id.
func (c *Check) GetByID(id int64) (*CheckRow, error) {
	session, err := c.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT id, cluster_id, name, interval, is_silenced, hosts_query, hosts_list, expressions, triggers, last_result_hosts, last_result_expressions FROM %v WHERE id=?", c.table)

	var scannedID, scannedClusterID int64
	var scannedIsSilenced bool
	var scannedName, scannedInterval, scannedHostsQuery, scannedHostsList, scannedExpressions, scannedTriggers, scannedLastResultHosts, scannedLastResultExpressions string

	err = session.Query(query, id).Scan(&scannedID, &scannedClusterID, &scannedName, &scannedInterval, &scannedIsSilenced, &scannedHostsQuery, &scannedHostsList, &scannedExpressions, &scannedTriggers, &scannedLastResultHosts, &scannedLastResultExpressions)
	if err != nil {
		return nil, err
	}

	row := &CheckRow{
		ID:                    scannedID,
		ClusterID:             scannedClusterID,
		Name:                  scannedName,
		Interval:              scannedInterval,
		IsSilenced:            scannedIsSilenced,
		HostsQuery:            scannedHostsQuery,
		HostsList:             scannedHostsList,
		Expressions:           scannedExpressions,
		Triggers:              scannedTriggers,
		LastResultHosts:       scannedLastResultHosts,
		LastResultExpressions: scannedLastResultExpressions,
	}

	return row, err
}

// AllByClusterID returns all rows by cluster_id.
func (c *Check) AllByClusterID(clusterID int64) ([]*CheckRow, error) {
	session, err := c.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	rows := []*CheckRow{}

	query := fmt.Sprintf(`SELECT id, cluster_id, name, interval, is_silenced, hosts_query, hosts_list, expressions, triggers, last_result_hosts, last_result_expressions FROM %v WHERE cluster_id=? ALLOW FILTERING`, c.table)

	var scannedID, scannedClusterID int64
	var scannedIsSilenced bool
	var scannedName, scannedInterval, scannedHostsQuery, scannedHostsList, scannedExpressions, scannedTriggers, scannedLastResultHosts, scannedLastResultExpressions string

	iter := session.Query(query, clusterID).Iter()
	for iter.Scan(&scannedID, &scannedClusterID, &scannedName, &scannedInterval, &scannedIsSilenced, &scannedHostsQuery, &scannedHostsList, &scannedExpressions, &scannedTriggers, &scannedLastResultHosts, &scannedLastResultExpressions) {
		rows = append(rows, &CheckRow{
			ID:                    scannedID,
			ClusterID:             scannedClusterID,
			Name:                  scannedName,
			Interval:              scannedInterval,
			IsSilenced:            scannedIsSilenced,
			HostsQuery:            scannedHostsQuery,
			HostsList:             scannedHostsList,
			Expressions:           scannedExpressions,
			Triggers:              scannedTriggers,
			LastResultHosts:       scannedLastResultHosts,
			LastResultExpressions: scannedLastResultExpressions,
		})
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{"Method": "Check.AllByClusterID"}).Error(err)

		return nil, err
	}
	return rows, err
}

// All returns all rows.
func (c *Check) All() ([]*CheckRow, error) {
	session, err := c.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	rows := []*CheckRow{}

	query := fmt.Sprintf(`SELECT id, cluster_id, name, interval, is_silenced, hosts_query, hosts_list, expressions, triggers, last_result_hosts, last_result_expressions FROM %v ALLOW FILTERING`, c.table)

	var scannedID, scannedClusterID int64
	var scannedIsSilenced bool
	var scannedName, scannedInterval, scannedHostsQuery, scannedHostsList, scannedExpressions, scannedTriggers, scannedLastResultHosts, scannedLastResultExpressions string

	iter := session.Query(query).Iter()
	for iter.Scan(&scannedID, &scannedClusterID, &scannedName, &scannedInterval, &scannedIsSilenced, &scannedHostsQuery, &scannedHostsList, &scannedExpressions, &scannedTriggers, &scannedLastResultHosts, &scannedLastResultExpressions) {
		rows = append(rows, &CheckRow{
			ID:                    scannedID,
			ClusterID:             scannedClusterID,
			Name:                  scannedName,
			Interval:              scannedInterval,
			IsSilenced:            scannedIsSilenced,
			HostsQuery:            scannedHostsQuery,
			HostsList:             scannedHostsList,
			Expressions:           scannedExpressions,
			Triggers:              scannedTriggers,
			LastResultHosts:       scannedLastResultHosts,
			LastResultExpressions: scannedLastResultExpressions,
		})
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{"Method": "Check.AllByClusterID"}).Error(err)

		return nil, err
	}
	return rows, err
}

// AllSplitToDaemons returns all rows divided into daemons equally.
func (c *Check) AllSplitToDaemons(daemons []string) (map[string][]*CheckRow, error) {
	result := make(map[string][]*CheckRow)

	if len(daemons) == 0 {
		return result, nil
	}

	rows, err := c.All()
	if err != nil {
		return nil, err
	}

	buckets := make([][]*CheckRow, len(daemons))
	for i, _ := range daemons {
		buckets[i] = make([]*CheckRow, 0)
	}

	bucketsPointer := 0
	for _, row := range rows {
		if bucketsPointer >= len(buckets) {
			bucketsPointer = 0
		}

		buckets[bucketsPointer] = append(buckets[bucketsPointer], row)
		bucketsPointer = bucketsPointer + 1
	}

	for i, checksInbucket := range buckets {
		result[daemons[i]] = checksInbucket
	}

	return result, err
}

func (c *Check) Create(clusterID int64, data map[string]interface{}) (*CheckRow, error) {
	session, err := c.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	id := NewExplicitID()

	query := fmt.Sprintf("INSERT INTO %v (id, cluster_id, name, interval, hosts_query, hosts_list, expressions, triggers, last_result_hosts, last_result_expressions) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", c.table)

	err = session.Query(query, id, clusterID, data["name"], data["interval"], data["hosts_query"], data["hosts_list"], data["expressions"], "[]", "[]", "[]").Exec()
	if err != nil {
		return nil, err
	}

	return &CheckRow{
		ID:                    id,
		ClusterID:             clusterID,
		Name:                  data["name"].(string),
		Interval:              data["interval"].(string),
		HostsQuery:            data["hosts_query"].(string),
		HostsList:             data["hosts_query"].(string),
		Expressions:           data["hosts_query"].(string),
		Triggers:              "[]",
		LastResultHosts:       "[]",
		LastResultExpressions: "[]",
	}, nil
}

func (c *Check) AddTrigger(checkRow *CheckRow, trigger CheckTrigger) ([]CheckTrigger, error) {
	session, err := c.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	triggers, err := checkRow.UnmarshalTriggers()
	if err != nil {
		return nil, err
	}

	triggers = append(triggers, trigger)

	triggersJSON, err := json.Marshal(triggers)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("UPDATE %v SET triggers=? WHERE id=?", c.table)

	err = session.Query(query, triggersJSON, checkRow.ID).Exec()
	if err != nil {
		return nil, err
	}

	return triggers, nil
}

func (c *Check) UpdateTrigger(checkRow *CheckRow, trigger CheckTrigger) ([]CheckTrigger, error) {
	session, err := c.GetCassandraSession()
	if err != nil {
		return nil, err
	}

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

	query := fmt.Sprintf("UPDATE %v SET triggers=? WHERE id=?", c.table)

	err = session.Query(query, triggersJSON, checkRow.ID).Exec()
	if err != nil {
		return nil, err
	}

	return triggers, nil
}

func (c *Check) DeleteTrigger(checkRow *CheckRow, trigger CheckTrigger) ([]CheckTrigger, error) {
	session, err := c.GetCassandraSession()
	if err != nil {
		return nil, err
	}

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

	query := fmt.Sprintf("UPDATE %v SET triggers=? WHERE id=?", c.table)

	err = session.Query(query, triggersJSON, checkRow.ID).Exec()
	if err != nil {
		return nil, err
	}

	return triggers, nil
}

// UpdateByID updates check by id.
func (c *Check) UpdateByID(id int64, data map[string]string) (*CheckRow, error) {
	session, err := c.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	row, err := c.GetByID(id)
	if err != nil {
		return nil, err
	}

	fields := make([]string, 0)
	fieldValues := make([]interface{}, 0)

	name, ok := data["name"]
	if ok {
		fields = append(fields, "name=?")
		fieldValues = append(fieldValues, name)
	}

	interval, ok := data["interval"]
	if ok {
		fields = append(fields, "interval=?")
		fieldValues = append(fieldValues, interval)
	}

	hostsQuery, ok := data["hosts_query"]
	if ok {
		fields = append(fields, "hosts_query=?")
		fieldValues = append(fieldValues, hostsQuery)
	}

	hostsList, ok := data["hosts_list"]
	if ok {
		fields = append(fields, "hosts_list=?")
		fieldValues = append(fieldValues, hostsList)
	}

	expressions, ok := data["expressions"]
	if ok {
		fields = append(fields, "expressions=?")
		fieldValues = append(fieldValues, expressions)
	}

	fieldValues = append(fieldValues, row.ID)

	query := fmt.Sprintf("UPDATE %v SET %v WHERE id=?", c.table, strings.Join(fields, ","))

	err = session.Query(query, fieldValues...).Exec()
	if err != nil {
		return nil, err
	}

	return c.GetByID(id)
}

// UpdateSilenceByID updates check is_silence state by id.
func (c *Check) UpdateSilenceByID(id int64, isSilenced bool) (*CheckRow, error) {
	session, err := c.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("UPDATE %v SET is_silenced=? WHERE id=?", c.table)

	err = session.Query(query, isSilenced, id).Exec()
	if err != nil {
		return nil, err
	}

	return c.GetByID(id)
}

func (c *Check) DeleteByClusterIDAndID(clusterID, id int64) (err error) {
	session, err := c.GetCassandraSession()
	if err != nil {
		return err
	}

	query := fmt.Sprintf("DELETE FROM %v WHERE id=? AND cluster_id=?", c.table)

	logrus.WithFields(logrus.Fields{
		"Method": "Check.DeleteByClusterIDAndID",
		"Query":  query,
	}).Info("Delete Query")

	return session.Query(query, id, clusterID).Exec()
}

func (checkRow *CheckRow) GetTriggers() []CheckTrigger {
	triggers, _ := checkRow.UnmarshalTriggers()

	return triggers
}

func (checkRow *CheckRow) UnmarshalTriggers() ([]CheckTrigger, error) {
	var triggers []CheckTrigger

	err := json.Unmarshal([]byte(checkRow.Triggers), &triggers)
	if err != nil {
		return nil, err
	}

	return triggers, nil
}

func (checkRow *CheckRow) GetHostsList() ([]string, error) {
	var container []string

	err := json.Unmarshal([]byte(checkRow.HostsList), &container)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (checkRow *CheckRow) GetExpressions() ([]CheckExpression, error) {
	var expressions []CheckExpression

	err := json.Unmarshal([]byte(checkRow.Expressions), &expressions)
	if err != nil {
		return expressions, err
	}

	return expressions, nil
}

// TODO: finish this
// func (checkRow *CheckRow) RunTriggers(appConfig config.GeneralConfig, coreDB *sqlx.DB, tsCheckDB *sqlx.DB, mailr *mailer.Mailer) error {
func (checkRow *CheckRow) RunTriggers(ctx context.Context) error {
	if checkRow.IsSilenced {
		return nil
	}

	triggers, err := checkRow.UnmarshalTriggers()
	if err != nil {
		logrus.Error(err)
		return err
	}

	for _, trigger := range triggers {
		tsCheckRows, err := NewTSCheck(ctx).AllViolationsByClusterIDCheckIDAndInterval(checkRow.ClusterID, checkRow.ID, trigger.CreatedIntervalMinute)
		if err != nil {
			return err
		}

		if len(tsCheckRows) == 0 {
			continue
		}

		lastViolation := tsCheckRows[0]
		violationsCount := len(tsCheckRows)

		if int64(violationsCount) >= trigger.LowViolationsCount && int64(violationsCount) <= trigger.HighViolationsCount {
			if trigger.Action.Transport == "nothing" {
				// Do nothing

			} else if trigger.Action.Transport == "email" {
				err = checkRow.RunEmailTrigger(ctx, trigger, lastViolation, violationsCount)
				if err != nil {
					logrus.Error(err)
					continue
				}

			} else if trigger.Action.Transport == "sms" {
				err = checkRow.RunSMSTrigger(ctx, trigger, lastViolation, violationsCount)
				if err != nil {
					logrus.Error(err)
					continue
				}

			} else if trigger.Action.Transport == "pagerduty" {
				err = checkRow.RunPagerDutyTrigger(ctx, trigger, lastViolation)
				if err != nil {
					logrus.Error(err)
					continue
				}
			}
		}
	}

	return nil
}

func (checkRow *CheckRow) BuildEmailTriggerContent(lastViolation *TSCheckRow, templateRoot string) (string, error) {
	funcMap := template.FuncMap{
		"addInt": func(left, right int) int {
			return left + right
		},
	}

	t, err := template.New("email-trigger.txt.tmpl").Funcs(funcMap).ParseFiles(libstring.ExpandTildeAndEnv(templateRoot + "/templates/checks/email-trigger.txt.tmpl"))
	if err != nil {
		return "", err
	}

	var contentBuffer bytes.Buffer

	vars := struct {
		Check         *CheckRow
		LastViolation *TSCheckRow
	}{
		checkRow,
		lastViolation,
	}

	err = t.Execute(&contentBuffer, vars)
	if err != nil {
		return "", err
	}

	return contentBuffer.String(), nil
}

func (checkRow *CheckRow) RunEmailTrigger(ctx context.Context, trigger CheckTrigger, lastViolation *TSCheckRow, violationsCount int) (err error) {
	if trigger.Action.Email == "" {
		return fmt.Errorf("Unable to send email because trigger.Action.Email is empty")
	}

	mailr, err := contexthelper.GetMailer(ctx, "GeneralConfig.Checks")
	if err != nil {
		logrus.Error(err)
		return err
	}

	to := trigger.Action.Email
	subject := fmt.Sprintf(`Check(ID: %v): %v, failed %v times`, checkRow.ID, checkRow.Name, violationsCount)
	body := ""

	if lastViolation != nil {
		body, err = checkRow.BuildEmailTriggerContent(lastViolation, ".")
		if err != nil {
			return fmt.Errorf("Unable to send email because of malformed email content. Error: %v", err)
		}
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

func (checkRow *CheckRow) RunSMSTrigger(ctx context.Context, trigger CheckTrigger, lastViolation *TSCheckRow, violationsCount int) (err error) {
	carrier := strings.ToLower(trigger.Action.SMSCarrier)

	generalConfig, err := contexthelper.GetGeneralConfig(ctx)
	if err != nil {
		logrus.Error(err)
		return err
	}

	gateway, ok := generalConfig.Checks.SMSEmailGateway[carrier]
	if !ok {
		return fmt.Errorf("Unable to lookup SMS Gateway for carrier: %v", carrier)
	}

	mailr, err := contexthelper.GetMailer(ctx, "GeneralConfig.Checks")
	if err != nil {
		logrus.Error(err)
		return err
	}

	flattenPhone := libstring.FlattenPhone(trigger.Action.SMSPhone)
	if len(flattenPhone) != 10 {
		logrus.Warningf("Length of phone number is not 10. Flatten phone number: %v. Length: %v", flattenPhone, len(flattenPhone))
		return nil
	}

	to := fmt.Sprintf("%v@%v", flattenPhone, gateway)
	subject := ""
	body := fmt.Sprintf(`Check(ID: %v): %v, failed %v times`, checkRow.ID, checkRow.Name, violationsCount)

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

func (checkRow *CheckRow) RunPagerDutyTrigger(ctx context.Context, trigger CheckTrigger, lastViolation *TSCheckRow) (err error) {
	// Create a new PD "trigger" event
	event := pagerduty.NewTriggerEvent(trigger.Action.PagerDutyServiceKey, trigger.Action.PagerDutyDescription)

	// Add details to PD event
	if lastViolation != nil {
		err = json.Unmarshal([]byte(lastViolation.Expressions), &event.Details)
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

	// Update incident key into check action JSON
	// Should we reuse the same incident key? Sounds like we should.
	trigger.Action.PagerDutyIncidentKey = pdResponse.IncidentKey
	// TODO: How do I save this?

	return err
}
