package cassandra

import (
	"testing"
	"time"

	_ "github.com/lib/pq"

	"github.com/resourced/resourced-master/models/pg"
	"github.com/resourced/resourced-master/models/shared"
)

type mockHostRow struct {
	ClusterID int64
	Hostname  string
}

func (mock *mockHostRow) DataAsFlatKeyValue() map[string]map[string]interface{} {
	innerData := make(map[string]interface{})
	innerData["metric.key"] = float64(1)

	data := make(map[string]map[string]interface{})
	data["/test"] = innerData
	return data
}

func (mock *mockHostRow) GetClusterID() int64 {
	return mock.ClusterID
}

func (mock *mockHostRow) GetHostname() string {
	return "localhost"
}

func TestTSMetricCreateValue(t *testing.T) {
	appContext := shared.AppContextCassandraForTest()

	u := pg.NewUser(appContext)

	pgdb, err := u.GetPGDB()
	if err != nil {
		t.Fatalf("There should be a legit db. Error: %v", err)
	}
	defer pgdb.Close()

	// Signup
	userRow, err := u.Signup(nil, newEmailForTest(), "abc123", "abc123")
	if err != nil {
		t.Errorf("Signing up user should work. Error: %v", err)
	}
	if userRow == nil {
		t.Fatal("Signing up user should work.")
	}
	if userRow.ID <= 0 {
		t.Fatal("Signing up user should work.")
	}

	// Create cluster for user
	clusterRow, err := pg.NewCluster(appContext).Create(nil, userRow, "cluster-name")
	if err != nil {
		t.Fatalf("Creating a cluster for user should work. Error: %v", err)
	}
	if clusterRow.ID <= 0 {
		t.Fatalf("Cluster ID should be assign properly. clusterRow.ID: %v", clusterRow.ID)
	}

	m := pg.NewMetric(appContext)

	// Create Metric
	metricRow, err := m.CreateOrUpdate(nil, clusterRow.ID, "/test.metric.key")
	if err != nil {
		t.Fatalf("Creating a Metric should work. Error: %v", err)
	}
	if metricRow.ID <= 0 {
		t.Fatalf("Metric ID should be assign properly. MetricRow.ID: %v", metricRow.ID)
	}

	mock := &mockHostRow{
		ClusterID: clusterRow.ID,
	}

	metricsMap := make(map[string]int64)
	metricsMap["/test.metric.key"] = metricRow.ID

	tsMetric := NewTSMetric(appContext)

	// Create TSMetric
	err = tsMetric.CreateByHostRow(mock, metricsMap, time.Duration(60))
	if err != nil {
		t.Fatalf("Creating a TSMetric should work. Error: %v", err)
	}

	to := time.Now().UTC().Unix()
	from := to - 60

	result, err := tsMetric.AllByMetricIDAndRange(clusterRow.ID, metricRow.ID, from, to)
	if err != nil {
		t.Errorf("Fetching TSMetric should not fail. Error: %v", err)
	}
	if len(result) <= 0 {
		t.Errorf("There should be a legit result.")
	}

	// Delete Metric
	_, err = m.DeleteByID(nil, metricRow.ID)
	if err != nil {
		t.Fatalf("Deleting Metrics by id should not fail. Error: %v", err)
	}

	// DELETE FROM clusters WHERE id=...
	_, err = pg.NewCluster(appContext).DeleteByID(nil, clusterRow.ID)
	if err != nil {
		t.Fatalf("Deleting clusters by id should not fail. Error: %v", err)
	}

	// DELETE FROM users WHERE id=...
	_, err = u.DeleteByID(nil, userRow.ID)
	if err != nil {
		t.Fatalf("Deleting user by id should not fail. Error: %v", err)
	}
}
