package shims

import (
	"context"
	"math"
	"time"
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
}
