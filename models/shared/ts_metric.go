package shared

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
