package shared

import (
	"math"
)

type TSMetricHighchartPayload struct {
	Name string          `json:"name"`
	Data [][]interface{} `json:"data"`
}

type TSMetricAggregateRow struct {
	ClusterID int64   `db:"cluster_id"`
	Key       string  `db:"key"`
	Host      string  `db:"host"`
	Avg       float64 `db:"avg"`
	Max       float64 `db:"max"`
	Min       float64 `db:"min"`
	Sum       float64 `db:"sum"`
}

type TSMetricRow struct {
	ClusterID int64   `db:"cluster_id"`
	MetricID  int64   `db:"metric_id"`
	Created   int64   `db:"created"`
	Key       string  `db:"key"`
	Host      string  `db:"host"`
	Value     float64 `db:"value"`
}

// LTTB down-samples the data to contain only threshold number of points that have the same visual shape as the original data
// Reference: https://github.com/sveinn-steinarsson/flot-downsample/
func LTTB(data [][]interface{}, threshold int) [][]interface{} {
	if threshold >= len(data) || threshold == 0 {
		return data
	}

	sampled := make([][]interface{}, 0, threshold)

	// Bucket size. Leave room for start and end data points
	every := float64(len(data)-2) / float64(threshold-2)

	sampled = append(sampled, data[0]) // Always add the first point

	bucketStart := 0
	bucketCenter := int(math.Floor(every)) + 1

	var a int

	for i := 0; i < threshold-2; i++ {

		bucketEnd := int(math.Floor(float64(i+2)*every)) + 1

		// Calculate point average for next bucket (containing c)
		avgRangeStart := bucketCenter
		avgRangeEnd := bucketEnd

		if avgRangeEnd >= len(data) {
			avgRangeEnd = len(data)
		}

		avgRangeLength := float64(avgRangeEnd - avgRangeStart)

		var avgX, avgY float64
		for ; avgRangeStart < avgRangeEnd; avgRangeStart++ {
			avgX += data[avgRangeStart][0].(float64)
			avgY += data[avgRangeStart][1].(float64)
		}
		avgX /= avgRangeLength
		avgY /= avgRangeLength

		// Get the range for this bucket
		rangeOffs := bucketStart
		rangeTo := bucketCenter

		// Point a
		pointAX := data[a][0].(float64)
		pointAY := data[a][1].(float64)

		var maxArea float64

		var nextA int
		for ; rangeOffs < rangeTo; rangeOffs++ {
			// Calculate triangle area over three buckets
			area := math.Abs((pointAX-avgX)*(data[rangeOffs][1].(float64)-pointAY) - (pointAX-data[rangeOffs][0].(float64))*(avgY-pointAY))
			if area > maxArea {
				maxArea = area
				nextA = rangeOffs // Next a is this b
			}
		}

		sampled = append(sampled, data[nextA]) // Pick this point from the bucket
		a = nextA                              // This a is the next a (chosen b)

		bucketStart = bucketCenter
		bucketCenter = bucketEnd
	}

	sampled = append(sampled, data[len(data)-1]) // Always add last

	return sampled
}
