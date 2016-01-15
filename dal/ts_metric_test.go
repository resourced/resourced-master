package dal

import (
	_ "github.com/lib/pq"
	"testing"
	"time"
)

func newTSMetricForTest(t *testing.T) *TSMetric {
	return NewTSMetric(newDbForTest(t))
}

func TestTSMetricCreateValue(t *testing.T) {
	u := newUserForTest(t)

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
	clusterRow, err := newClusterForTest(t).Create(nil, userRow.ID, "cluster-name")
	if err != nil {
		t.Fatalf("Creating a cluster for user should work. Error: %v", err)
	}
	if clusterRow.ID <= 0 {
		t.Fatalf("Cluster ID should be assign properly. clusterRow.ID: %v", clusterRow.ID)
	}

	// Create Metric
	metricRow, err := newMetricForTest(t).CreateOrUpdate(nil, clusterRow.ID, "/test.metric.key")
	if err != nil {
		t.Fatalf("Creating a Metric should work. Error: %v", err)
	}
	if metricRow.ID <= 0 {
		t.Fatalf("Metric ID should be assign properly. MetricRow.ID: %v", metricRow.ID)
	}

	// Create TSMetric
	int64Value := time.Now().UnixNano()
	err = newTSMetricForTest(t).Create(nil, clusterRow.ID, metricRow.ID, "localhost", "/test.metric.key", float64(int64Value))
	if err != nil {
		t.Fatalf("Creating a TSMetric should work. Error: %v", err)
	}

	specificNumber := +3.730022e+009
	err = newTSMetricForTest(t).Create(nil, clusterRow.ID, metricRow.ID, "localhost", "/test.metric.key", float64(specificNumber))
	if err != nil {
		t.Fatalf("Creating a TSMetric should work. Error: %v", err)
	}

	// Delete Metric
	_, err = newMetricForTest(t).DeleteByID(nil, metricRow.ID)
	if err != nil {
		t.Fatalf("Deleting Metrics by id should not fail. Error: %v", err)
	}

	// DELETE FROM clusters WHERE id=...
	_, err = NewCluster(u.db).DeleteByID(nil, clusterRow.ID)
	if err != nil {
		t.Fatalf("Deleting clusters by id should not fail. Error: %v", err)
	}

	// DELETE FROM users WHERE id=...
	_, err = u.DeleteByID(nil, userRow.ID)
	if err != nil {
		t.Fatalf("Deleting user by id should not fail. Error: %v", err)
	}
}
