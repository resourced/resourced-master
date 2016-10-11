package pg

import (
	"testing"
	"time"

	_ "github.com/lib/pq"

	"github.com/resourced/resourced-master/models/shared"
)

func TestTSMetricCreateValue(t *testing.T) {
	appContext := shared.AppContextForTest()

	u := NewUser(appContext)

	pgdb, err := u.GetPGDB()
	if err != nil {
		t.Errorf("There should be a legit db. Error: %v", err)
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
	clusterRow, err := NewCluster(appContext).Create(nil, userRow, "cluster-name")
	if err != nil {
		t.Fatalf("Creating a cluster for user should work. Error: %v", err)
	}
	if clusterRow.ID <= 0 {
		t.Fatalf("Cluster ID should be assign properly. clusterRow.ID: %v", clusterRow.ID)
	}

	m := NewMetric(appContext)

	// Create Metric
	metricRow, err := m.CreateOrUpdate(nil, clusterRow.ID, "/test.metric.key")
	if err != nil {
		t.Fatalf("Creating a Metric should work. Error: %v", err)
	}
	if metricRow.ID <= 0 {
		t.Fatalf("Metric ID should be assign properly. MetricRow.ID: %v", metricRow.ID)
	}

	tsMetric := NewTSMetric(appContext, clusterRow.ID)

	// Create TSMetric
	int64Value := time.Now().UnixNano()
	err = tsMetric.Create(nil, clusterRow.ID, metricRow.ID, "localhost", "/test.metric.key", float64(int64Value), time.Now().Unix()+int64(900))
	if err != nil {
		t.Fatalf("Creating a TSMetric should work. Error: %v", err)
	}

	specificNumber := +3.730022e+009
	err = tsMetric.Create(nil, clusterRow.ID, metricRow.ID, "localhost", "/test.metric.key", float64(specificNumber), time.Now().Unix()+int64(900))
	if err != nil {
		t.Fatalf("Creating a TSMetric should work. Error: %v", err)
	}

	// Delete Metric
	_, err = m.DeleteByID(nil, metricRow.ID)
	if err != nil {
		t.Fatalf("Deleting Metrics by id should not fail. Error: %v", err)
	}

	// DELETE FROM clusters WHERE id=...
	_, err = NewCluster(appContext).DeleteByID(nil, clusterRow.ID)
	if err != nil {
		t.Fatalf("Deleting clusters by id should not fail. Error: %v", err)
	}

	// DELETE FROM users WHERE id=...
	_, err = u.DeleteByID(nil, userRow.ID)
	if err != nil {
		t.Fatalf("Deleting user by id should not fail. Error: %v", err)
	}
}
