// Package dal is the Data Access Layer between the Application and PostgreSQL database.
package cassandra

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/gocql/gocql"

	"github.com/resourced/resourced-master/contexthelper"
)

var PROJECT_EPOCH = 1451606400

// NewExplicitID uses UNIX timestamp in microseconds as ID.
func NewExplicitID() int64 {
	currentTime := time.Now().UnixNano()
	projectEpochInNanoSeconds := int64(PROJECT_EPOCH * 1000 * 1000 * 1000)

	resultInNanoSeconds := currentTime - projectEpochInNanoSeconds
	resultInMicroSeconds := int64(math.Floor(float64(resultInNanoSeconds / 1000)))

	return resultInMicroSeconds
}

type Base struct {
	AppContext context.Context
	session    *gocql.Session
	table      string
	hasID      bool
}

func (b *Base) GetCassandraSession() (*gocql.Session, error) {
	cassandradbs, err := contexthelper.GetCassandraDBConfig(b.AppContext)
	if err != nil {
		return nil, err
	}

	return cassandradbs.CoreSession, nil
}

// UpdateEnabledByID updates level by id.
func (b *Base) DeleteByID(id int64) error {
	session, err := b.GetCassandraSession()
	if err != nil {
		return err
	}

	query := fmt.Sprintf("DELETE FROM %v WHERE id = ?", b.table)

	return session.Query(query, id).Exec()
}
