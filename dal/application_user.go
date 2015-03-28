package dal

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/resourced/resourced-master/libstring"
)

func NewApplicationUser(db *sqlx.DB) *ApplicationUser {
	app := &ApplicationUser{}
	app.db = db
	app.table = "applications_users"
	app.hasID = true

	return app
}

type ApplicationUserRow struct {
	ID            int64  `db:"id"`
	ApplicationID int64  `db:"application_id"`
	UserID        int64  `db:"user_id"`
	Token         string `db:"token"`
	Level         string `db:"level"`
}

type ApplicationUser struct {
	Base
}

func (a *ApplicationUser) GetByPrimaryKey(tx *sqlx.Tx, appId, userId int64) (*ApplicationUserRow, error) {
	app := &ApplicationUserRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE application_id=$1 AND user_id=$2", a.table)
	err := a.db.Get(app, query, appId, userId)

	return app, err
}

func (a *ApplicationUser) CreateRow(tx *sqlx.Tx, appId, userId int64, level string) (*ApplicationUserRow, error) {
	token, err := libstring.GeneratePassword(32)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	data["application_id"] = appId
	data["user_id"] = userId
	data["token"] = token
	data["level"] = level

	// TODO: This insert is not inserting and returns no error.
	_, err = a.InsertIntoTable(tx, data)
	if err != nil {
		return nil, err
	}

	return a.GetByPrimaryKey(tx, appId, userId)
}
