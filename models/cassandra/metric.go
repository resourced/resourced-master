package cassandra

import (
	"context"
	"fmt"

	"github.com/Sirupsen/logrus"
)

func NewMetric(ctx context.Context) *Metric {
	m := &Metric{}
	m.AppContext = ctx
	m.table = "metrics"

	return m
}

type MetricRowsWithError struct {
	Metrics []*MetricRow
	Error   error
}

type MetricsMapWithError struct {
	MetricsMap map[string]int64
	Error      error
}

type MetricRow struct {
	ID        int64  `db:"id"`
	ClusterID int64  `db:"cluster_id"`
	Key       string `db:"key"`
}

type Metric struct {
	Base
}

// GetByID returns one record by id.
func (m *Metric) GetByID(id int64) (*MetricRow, error) {
	session, err := m.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	// id bigint primary key,
	// cluster_id bigint,
	// key text

	query := fmt.Sprintf("SELECT id, cluster_id, key FROM %v WHERE id=?", m.table)

	var scannedID, scannedClusterID int64
	var scannedKey string

	err = session.Query(query, id).Scan(&scannedID, &scannedClusterID, &scannedKey)
	if err != nil {
		return nil, err
	}

	row := &MetricRow{
		ID:        scannedID,
		ClusterID: scannedClusterID,
		Key:       scannedKey,
	}

	return row, err
}

// GetByClusterIDAndKey returns one record by cluster_id and key.
func (m *Metric) GetByClusterIDAndKey(clusterID int64, key string) (*MetricRow, error) {
	session, err := m.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	// id bigint primary key,
	// cluster_id bigint,
	// key text

	query := fmt.Sprintf("SELECT id, cluster_id, key FROM %v WHERE id=? AND key=? ALLOW FILTERING", m.table)

	var scannedID, scannedClusterID int64
	var scannedKey string

	err = session.Query(query, clusterID, key).Scan(&scannedID, &scannedClusterID, &scannedKey)
	if err != nil {
		return nil, err
	}

	row := &MetricRow{
		ID:        scannedID,
		ClusterID: scannedClusterID,
		Key:       scannedKey,
	}

	return row, err
}

func (m *Metric) CreateOrUpdate(clusterID int64, key string) (*MetricRow, error) {
	session, err := m.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	metricRow, err := m.GetByClusterIDAndKey(clusterID, key)

	// Perform INSERT
	if metricRow == nil || err != nil {
		id := NewExplicitID()

		query := fmt.Sprintf("INSERT INTO %v (id, cluster_id, key) VALUES (?, ?, ?)", m.table)

		err = session.Query(query, id, clusterID, key).Exec()
		if err != nil {
			return nil, err
		}

		return &MetricRow{
			ID:        id,
			ClusterID: clusterID,
			Key:       key,
		}, nil
	}

	return metricRow, nil
}

// AllByClusterID returns all rows.
func (m *Metric) AllByClusterID(clusterID int64) ([]*MetricRow, error) {
	session, err := m.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	rows := []*MetricRow{}

	query := fmt.Sprintf(`SELECT id, cluster_id, key FROM %v WHERE cluster_id=? ALLOW FILTERING`, m.table)

	var scannedID, scannedClusterID int64
	var scannedKey string

	iter := session.Query(query, clusterID).Iter()
	for iter.Scan(&scannedID, &scannedClusterID, &scannedKey) {
		rows = append(rows, &MetricRow{
			ID:        scannedID,
			ClusterID: scannedClusterID,
			Key:       scannedKey,
		})
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{"Method": "Metric.AllByClusterID"}).Error(err)

		return nil, err
	}
	return rows, err
}

// AllByClusterIDAsMap returns all rows.
func (m *Metric) AllByClusterIDAsMap(clusterID int64) (map[string]int64, error) {
	result := make(map[string]int64)

	rows, err := m.AllByClusterID(clusterID)
	if err != nil {
		return result, err
	}

	for _, row := range rows {
		result[row.Key] = row.ID
	}

	return result, nil
}
