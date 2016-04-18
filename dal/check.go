package dal

import (
	"database/sql"
	"fmt"
	"encoding/json"

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

// AllChecks returns all rows.
func (a *Check) AllChecks(tx *sqlx.Tx) ([]*CheckRow, error) {
	rows := []*CheckRow{}
	query := fmt.Sprintf("SELECT * FROM %v ORDER BY id DESC", a.table)
	err := a.db.Select(&rows, query)

	return rows, err
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

	for i, trig := range triggers {
		if trig.ID != trigger.ID {
			newTriggers[i] = trig
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

func (checkRow *CheckRow) UnmarshalTriggers() ([]CheckTrigger, error) {
	var container []CheckTrigger

	err := json.Unmarshal(checkRow.Triggers, &container)
	if err != nil {
		return nil, err
	}

	return container, nil
}

type CheckTrigger struct {
	ID                  int64
	LowViolationsCount  int64
	HighViolationsCount int64
	CreatedInterval     string
	Action              CheckTriggerAction
}

type CheckTriggerAction struct {
	Transport string
	Email string
	SMSPhone string
	SMSCarrier string
	PagerDutyServiceKey string
	PagerDutyIncidentKey string
	PagerDutyDescription string
}