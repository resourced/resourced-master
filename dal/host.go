package dal

import (
	"github.com/jmoiron/sqlx"
)

func NewHost(db *sqlx.DB) *Host {
	host := &Host{}
	host.db = db
	host.table = "hosts"
	host.hasID = true

	return host
}

type Host struct {
	Base
}
