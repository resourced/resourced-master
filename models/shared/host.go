package shared

type IHostRow interface {
	GetClusterID() int64
	GetHostname() string
	// GetData() map[string]string
}
