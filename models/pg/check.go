package pg

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"
	"github.com/marcw/pagerduty"

	"github.com/resourced/resourced-master/contexthelper"
	"github.com/resourced/resourced-master/libstring"
)

func NewCheck(ctx context.Context) *Check {
	g := &Check{}
	g.AppContext = ctx
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

func (c *Check) rowFromSqlResult(tx *sqlx.Tx, sqlResult sql.Result) (*CheckRow, error) {
	id, err := sqlResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	return c.GetByID(tx, id)
}

// GetByID returns one record by id.
func (c *Check) GetByID(tx *sqlx.Tx, id int64) (*CheckRow, error) {
	pgdb, err := c.GetPGDB()
	if err != nil {
		return nil, err
	}

	row := &CheckRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=$1", c.table)
	err = pgdb.Get(row, query, id)

	return row, err
}

func (c *Check) Create(tx *sqlx.Tx, clusterID int64, data map[string]interface{}) (*CheckRow, error) {
	data["cluster_id"] = clusterID

	sqlResult, err := c.InsertIntoTable(tx, data)
	if err != nil {
		return nil, err
	}

	return c.rowFromSqlResult(tx, sqlResult)
}

// AllByClusterID returns all rows by cluster_id.
func (c *Check) AllByClusterID(tx *sqlx.Tx, clusterID int64) ([]*CheckRow, error) {
	pgdb, err := c.GetPGDB()
	if err != nil {
		return nil, err
	}

	rows := []*CheckRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 ORDER BY id DESC", c.table)
	err = pgdb.Select(&rows, query, clusterID)

	return rows, err
}

// All returns all rows.
func (c *Check) All(tx *sqlx.Tx) ([]*CheckRow, error) {
	pgdb, err := c.GetPGDB()
	if err != nil {
		return nil, err
	}

	rows := []*CheckRow{}
	query := fmt.Sprintf("SELECT * FROM %v ORDER BY id DESC", c.table)
	err = pgdb.Select(&rows, query)

	return rows, err
}

// AllSplitToDaemons returns all rows divided into daemons equally.
func (c *Check) AllSplitToDaemons(tx *sqlx.Tx, daemons []string) (map[string][]*CheckRow, error) {
	result := make(map[string][]*CheckRow)

	if len(daemons) == 0 {
		return result, nil
	}

	rows, err := c.All(tx)
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

func (c *Check) AddTrigger(tx *sqlx.Tx, checkRow *CheckRow, trigger CheckTrigger) ([]CheckTrigger, error) {
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

	_, err = c.UpdateByID(tx, data, checkRow.ID)

	return triggers, err
}

func (c *Check) UpdateTrigger(tx *sqlx.Tx, checkRow *CheckRow, trigger CheckTrigger) ([]CheckTrigger, error) {
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

	_, err = c.UpdateByID(tx, data, checkRow.ID)

	return triggers, err
}

func (c *Check) DeleteTrigger(tx *sqlx.Tx, checkRow *CheckRow, trigger CheckTrigger) ([]CheckTrigger, error) {
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

	_, err = c.UpdateByID(tx, data, checkRow.ID)

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

	clusterRow, err := NewCluster(ctx).GetByID(nil, checkRow.ClusterID)
	if err != nil {
		logrus.Error(err)
		return err
	}

	deletedFrom := clusterRow.GetDeletedFromUNIXTimestampForSelect("ts_checks")

	for _, trigger := range triggers {
		tsCheckRows, err := NewTSCheck(ctx, checkRow.ClusterID).AllViolationsByClusterIDCheckIDAndInterval(nil, checkRow.ClusterID, checkRow.ID, trigger.CreatedIntervalMinute, deletedFrom)
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

	// Update incident key into check action JSON
	// Should we reuse the same incident key? Sounds like we should.
	trigger.Action.PagerDutyIncidentKey = pdResponse.IncidentKey
	// TODO: How do I save this?

	return err
}
