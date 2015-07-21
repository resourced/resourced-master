package dal

import (
	_ "github.com/lib/pq"
	"testing"
)

func newClusterForTest(t *testing.T) *Cluster {
	return NewCluster(newDbForTest(t))
}

func TestClusterCRUD(t *testing.T) {
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

	// SELECT * FROM clusters
	_, err = NewCluster(u.db).AllClustersByUserId(nil, userRow.ID)
	if err != nil {
		t.Fatalf("Selecting all clusters should not fail. Error: %v, userRow.ID: %v", err, userRow.ID)
	}

	// DELETE FROM clusters WHERE id=...
	_, err = NewCluster(u.db).DeleteById(nil, clusterRow.ID)
	if err != nil {
		t.Fatalf("Deleting clusters by id should not fail. Error: %v", err)
	}

	// DELETE FROM users WHERE id=...
	_, err = u.DeleteById(nil, userRow.ID)
	if err != nil {
		t.Fatalf("Deleting user by id should not fail. Error: %v", err)
	}
}
