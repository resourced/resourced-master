package pg

import (
	"testing"

	_ "github.com/lib/pq"
)

func newMetricForTest(t *testing.T) *Metric {
	return NewMetric(newDbForTest(t))
}

func TestMetricCRUD(t *testing.T) {
	u := newUserForTest(t)
	defer u.db.Close()

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

	cl := newClusterForTest(t)
	defer cl.db.Close()

	// Create cluster for user
	clusterRow, err := cl.Create(nil, userRow, "cluster-name")
	if err != nil {
		t.Fatalf("Creating a cluster for user should work. Error: %v", err)
	}
	if clusterRow.ID <= 0 {
		t.Fatalf("Cluster ID should be assign properly. clusterRow.ID: %v", clusterRow.ID)
	}

	// Create Metric
	m := newMetricForTest(t)
	defer m.db.Close()

	metricRow, err := m.CreateOrUpdate(nil, clusterRow.ID, "/test.metric.key")
	if err != nil {
		t.Fatalf("Creating a Metric should work. Error: %v", err)
	}
	if metricRow.ID <= 0 {
		t.Fatalf("Metric ID should be assign properly. MetricRow.ID: %v", metricRow.ID)
	}

	_, err = m.DeleteByID(nil, metricRow.ID)
	if err != nil {
		t.Fatalf("Deleting Metrics by id should not fail. Error: %v", err)
	}

	// DELETE FROM clusters WHERE id=...
	_, err = cl.DeleteByID(nil, clusterRow.ID)
	if err != nil {
		t.Fatalf("Deleting clusters by id should not fail. Error: %v", err)
	}

	// DELETE FROM users WHERE id=...
	_, err = u.DeleteByID(nil, userRow.ID)
	if err != nil {
		t.Fatalf("Deleting user by id should not fail. Error: %v", err)
	}
}
