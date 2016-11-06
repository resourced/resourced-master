package cassandra

import (
	"testing"

	_ "github.com/lib/pq"

	"github.com/resourced/resourced-master/models/shared"
)

func TestClusterCRUD(t *testing.T) {
	appContext := shared.AppContextForTest()

	u := NewUser(appContext)

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

	c := NewCluster(appContext)

	// Create cluster for user
	clusterRow, err := c.Create(nil, userRow, "cluster-name")
	if err != nil {
		t.Fatalf("Creating a cluster for user should work. Error: %v", err)
	}
	if clusterRow.ID <= 0 {
		t.Fatalf("Cluster ID should be assign properly. clusterRow.ID: %v", clusterRow.ID)
	}

	// SELECT * FROM clusters
	_, err = c.AllByUserID(nil, userRow.ID)
	if err != nil {
		t.Fatalf("Selecting all clusters should not fail. Error: %v, userRow.ID: %v", err, userRow.ID)
	}

	// DELETE FROM clusters WHERE id=...
	_, err = c.DeleteByID(nil, clusterRow.ID)
	if err != nil {
		t.Fatalf("Deleting clusters by id should not fail. Error: %v", err)
	}

	// DELETE FROM users WHERE id=...
	_, err = u.DeleteByID(nil, userRow.ID)
	if err != nil {
		t.Fatalf("Deleting user by id should not fail. Error: %v", err)
	}
}
