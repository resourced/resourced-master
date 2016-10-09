package shims

import (
	"github.com/gocql/gocql"
	"github.com/jmoiron/sqlx"
)

type Parameters struct {
	PGDB             *sqlx.DB
	PGTx             *sqlx.Tx
	CassandraSession *gocql.Session
	Table            string
	DBType           string
}
