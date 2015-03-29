package dal

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
)

func NewApplication(db *sqlx.DB) *Application {
	app := &Application{}
	app.db = db
	app.table = "applications"
	app.hasID = true

	return app
}

type ApplicationRow struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

type Application struct {
	Base
}

func (a *Application) applicationRowFromSqlResult(tx *sqlx.Tx, sqlResult sql.Result) (*ApplicationRow, error) {
	appId, err := sqlResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	return a.GetById(tx, appId)
}

func (a *Application) GetById(tx *sqlx.Tx, id int64) (*ApplicationRow, error) {
	app := &ApplicationRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=$1", a.table)
	err := a.db.Get(app, query, id)

	return app, err
}

func (a *Application) CreateRow(tx *sqlx.Tx, appName string) (*ApplicationRow, error) {
	data := make(map[string]interface{})
	data["name"] = appName

	sqlResult, err := a.InsertIntoTable(tx, data)
	if err != nil {
		return nil, err
	}

	return a.applicationRowFromSqlResult(tx, sqlResult)
}

// AllApplicationsByUserID returns all user rows.
func (a *Application) AllApplicationsByUserID(tx *sqlx.Tx, userId int64) ([]*ApplicationRow, error) {
	apps := []*ApplicationRow{}
	query := fmt.Sprintf("SELECT applications.id, applications.name FROM %v INNER JOIN applications_users ON %v.id = applications_users.application_id WHERE applications_users.user_id=$1;", a.table, a.table)
	err := a.db.Select(&apps, query, userId)

	return apps, err
}
