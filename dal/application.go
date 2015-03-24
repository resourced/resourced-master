package dal

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

func NewApplication(db *sqlx.DB) *Application {
	app := &Application{}
	app.db = db
	app.table = "applications"

	return app
}

type Application struct {
	Base
}

type ApplicationRow struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

func (a *Application) GetById(tx *sqlx.Tx, id int64) (*ApplicationRow, error) {
	app := &ApplicationRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=$1", a.table)
	err := a.db.Get(app, query, id)

	return app, err
}
