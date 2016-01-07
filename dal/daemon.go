package dal

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

func NewDaemon(db *sqlx.DB) *Daemon {
	application := &Daemon{}
	application.db = db
	application.table = "daemons"
	application.hasID = true

	return application
}

type DaemonRow struct {
	ID       int64     `db:"id"`
	Hostname string    `db:"hostname"`
	Updated  time.Time `db:"updated"`
}

type Daemon struct {
	Base
}

func (d *Daemon) daemonRowFromSqlResult(tx *sqlx.Tx, sqlResult sql.Result) (*DaemonRow, error) {
	id, err := sqlResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	return d.GetById(tx, id)
}

// GetById returns one record by id.
func (d *Daemon) GetById(tx *sqlx.Tx, id int64) (*DaemonRow, error) {
	row := &DaemonRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=$1", d.table)
	err := d.db.Get(row, query, id)

	return row, err
}

// GetByHostname returns one record by hostname.
func (d *Daemon) GetByHostname(tx *sqlx.Tx, hostname string) (*DaemonRow, error) {
	row := &DaemonRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE hostname=$1", d.table)
	err := d.db.Get(row, query, hostname)

	return row, err
}

func (d *Daemon) CreateOrUpdate(tx *sqlx.Tx, hostname string) (*DaemonRow, error) {
	daemonRow, err := d.GetByHostname(tx, hostname)

	data := make(map[string]interface{})
	data["hostname"] = hostname

	// Perform INSERT
	if daemonRow == nil || err != nil {
		sqlResult, err := d.InsertIntoTable(tx, data)
		if err != nil {
			return nil, err
		}

		return d.daemonRowFromSqlResult(tx, sqlResult)
	}

	// Perform UPDATE
	_, err = d.UpdateByKeyValueString(tx, data, "hostname", hostname)
	if err != nil {
		return nil, err
	}

	return daemonRow, nil
}

// AllDaemons returns all clusters rows.
func (d *Daemon) AllDaemons(tx *sqlx.Tx) ([]*DaemonRow, error) {
	rows := []*DaemonRow{}
	query := fmt.Sprintf("SELECT * FROM %v", d.table)
	err := d.db.Select(&rows, query)

	return rows, err
}
