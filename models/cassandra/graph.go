package cassandra

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Sirupsen/logrus"
)

func NewGraph(ctx context.Context) *Graph {
	g := &Graph{}
	g.AppContext = ctx
	g.table = "graphs"

	return g
}

type GraphRowsWithError struct {
	Graphs []*GraphRow
	Error  error
}

type GraphRow struct {
	ID          int64  `db:"id"`
	ClusterID   int64  `db:"cluster_id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Range       string `db:"range"`
	Metrics     string `db:"metrics"`
}

func (gr *GraphRow) MetricsFromJSON() []*MetricRow {
	metricRows := make([]*MetricRow, 0)
	json.Unmarshal([]byte(gr.Metrics), &metricRows)
	return metricRows
}

func (gr *GraphRow) MetricsFromJSONGroupByN(n int) [][]*MetricRow {
	container := make([][]*MetricRow, 0)

	for metricIndex, metricRow := range gr.MetricsFromJSON() {
		if metricIndex%n == 0 {
			container = append(container, make([]*MetricRow, 0))
		}

		container[len(container)-1] = append(container[len(container)-1], metricRow)
	}

	return container
}

func (gr *GraphRow) MetricsFromJSONGroupByThree() [][]*MetricRow {
	return gr.MetricsFromJSONGroupByN(3)
}

type Graph struct {
	Base
}

// GetByClusterIDAndID returns one record by id.
func (g *Graph) GetByClusterIDAndID(clusterID, id int64) (*GraphRow, error) {
	session, err := g.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT id, cluster_id, name, description, range, metrics FROM %v WHERE cluster_id=? AND id=?", g.table)

	var scannedID, scannedClusterID int64
	var scannedName, scannedDescription, scannedRange, scannedMetrics string

	err = session.Query(query, clusterID, id).Scan(&scannedID, &scannedClusterID, &scannedName, &scannedDescription, &scannedRange, &scannedMetrics)
	if err != nil {
		return nil, err
	}

	row := &GraphRow{
		ID:          scannedID,
		ClusterID:   scannedClusterID,
		Name:        scannedName,
		Description: scannedDescription,
		Range:       scannedRange,
		Metrics:     scannedMetrics,
	}

	return row, err
}

// GetByID returns one record by id.
func (g *Graph) GetByID(id int64) (*GraphRow, error) {
	session, err := g.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT id, cluster_id, name, description, range, metrics FROM %v WHERE id=?", g.table)

	var scannedID, scannedClusterID int64
	var scannedName, scannedDescription, scannedRange, scannedMetrics string

	err = session.Query(query, id).Scan(&scannedID, &scannedClusterID, &scannedName, &scannedDescription, &scannedRange, &scannedMetrics)
	if err != nil {
		return nil, err
	}

	row := &GraphRow{
		ID:          scannedID,
		ClusterID:   scannedClusterID,
		Name:        scannedName,
		Description: scannedDescription,
		Range:       scannedRange,
		Metrics:     scannedMetrics,
	}

	return row, err
}

func (g *Graph) Create(clusterID int64, data map[string]interface{}) (*GraphRow, error) {
	session, err := g.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	id := NewExplicitID()

	query := fmt.Sprintf("INSERT INTO %v (id, cluster_id, name, description, range, metrics) VALUES (?, ?, ?, ?, ?, ?)", g.table)

	err = session.Query(query, id, clusterID, data["name"], data["description"], data["range"], "[]").Exec()
	if err != nil {
		return nil, err
	}

	return &GraphRow{
		ID:          id,
		ClusterID:   clusterID,
		Name:        data["name"].(string),
		Description: data["description"].(string),
		Range:       data["range"].(string),
	}, nil
}

// AllByClusterID returns all rows by cluster_id.
func (g *Graph) AllByClusterID(clusterID int64) ([]*GraphRow, error) {
	session, err := g.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	rows := []*GraphRow{}

	query := fmt.Sprintf(`SELECT id, cluster_id, name, description, range, metrics FROM %v WHERE cluster_id=? ALLOW FILTERING`, g.table)

	var scannedID, scannedClusterID int64
	var scannedName, scannedDescription, scannedRange, scannedMetrics string

	iter := session.Query(query, clusterID).Iter()
	for iter.Scan(&scannedID, &scannedClusterID, &scannedName, &scannedDescription, &scannedRange, &scannedMetrics) {
		rows = append(rows, &GraphRow{
			ID:          scannedID,
			ClusterID:   scannedClusterID,
			Name:        scannedName,
			Description: scannedDescription,
			Range:       scannedRange,
			Metrics:     scannedMetrics,
		})
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{"Method": "Graph.AllByClusterID"}).Error(err)

		return nil, err
	}
	return rows, err
}

// AllGraphs returns all rows.
func (g *Graph) AllGraphs() ([]*GraphRow, error) {
	session, err := g.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	rows := []*GraphRow{}

	query := fmt.Sprintf(`SELECT id, cluster_id, name, description, range, metrics FROM %v`, g.table)

	var scannedID, scannedClusterID int64
	var scannedName, scannedDescription, scannedRange, scannedMetrics string

	iter := session.Query(query).Iter()
	for iter.Scan(&scannedID, &scannedClusterID, &scannedName, &scannedDescription, &scannedRange, &scannedMetrics) {
		rows = append(rows, &GraphRow{
			ID:          scannedID,
			ClusterID:   scannedClusterID,
			Name:        scannedName,
			Description: scannedDescription,
			Range:       scannedRange,
			Metrics:     scannedMetrics,
		})
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{"Method": "Graph.AllByClusterID"}).Error(err)

		return nil, err
	}
	return rows, err
}

// DeleteMetricFromGraphs deletes a particular metric from all rows by cluster_id.
func (g *Graph) DeleteMetricFromGraphs(clusterID, metricID int64) error {
	session, err := g.GetCassandraSession()
	if err != nil {
		return err
	}

	rows, err := g.AllByClusterID(clusterID)
	if err != nil {
		return err
	}

	for _, row := range rows {
		metrics := row.MetricsFromJSON()
		newMetrics := make([]*MetricRow, 0)

		for _, metric := range metrics {
			if metric.ID != metricID {
				newMetrics = append(newMetrics, metric)
			}
		}

		newMetricsJSONBytes, err := json.Marshal(newMetrics)
		if err != nil {
			return err
		}

		query := fmt.Sprintf("UPDATE %v SET metrics=? WHERE id=?", g.table)

		err = session.Query(query, string(newMetricsJSONBytes), row.ID).Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateByID updates metrics JSON and then returns record by id.
func (g *Graph) UpdateByID(id int64, data map[string]string) (*GraphRow, error) {
	session, err := g.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	row, err := g.GetByID(id)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("UPDATE %v SET name=?, description=?, range=?, metrics=? WHERE id=?", g.table)

	err = session.Query(query, data["name"], data["description"], data["range"], data["metrics"], row.ID).Exec()
	if err != nil {
		return nil, err
	}

	return g.GetByID(id)
}

// UpdateMetricsByClusterIDAndID updates metrics JSON and then returns record by id.
func (g *Graph) UpdateMetricsByClusterIDAndID(clusterID, id int64, metricsJSON []byte) (*GraphRow, error) {
	session, err := g.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	row, err := g.GetByClusterIDAndID(clusterID, id)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("UPDATE %v SET metrics=? WHERE id=?", g.table)

	err = session.Query(query, string(metricsJSON), row.ID).Exec()
	if err != nil {
		return nil, err
	}

	return g.GetByClusterIDAndID(clusterID, id)
}
