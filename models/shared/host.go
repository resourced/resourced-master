package shared

type IHostRow interface {
	DataAsFlatKeyValue() map[string]map[string]interface{}
	GetClusterID() int64
	GetHostname() string
	GetData() map[string]string
}
