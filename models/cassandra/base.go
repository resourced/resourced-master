// Package dal is the Data Access Layer between the Application and PostgreSQL database.
package cassandra

import (
	"context"
	"math"
	"time"

	"github.com/gocql/gocql"
)

var PROJECT_EPOCH = 1451606400

type Base struct {
	AppContext context.Context
	session    *gocql.Session
	table      string
	hasID      bool
}

// NewExplicitID uses UNIX timestamp in microseconds as ID.
func NewExplicitID() int64 {
	currentTime := time.Now().UnixNano()
	projectEpochInNanoSeconds := int64(PROJECT_EPOCH * 1000 * 1000 * 1000)

	resultInNanoSeconds := currentTime - projectEpochInNanoSeconds
	resultInMicroSeconds := int64(math.Floor(float64(resultInNanoSeconds / 1000)))

	return resultInMicroSeconds
}
