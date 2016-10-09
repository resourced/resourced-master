package shared

type ITSMetricHighchartPayload interface {
	GetName() string
	GetData() [][]interface{}
}
