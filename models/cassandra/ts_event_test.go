package cassandra

import (
	"testing"
	"time"

	_ "github.com/lib/pq"

	"github.com/resourced/resourced-master/models/shared"
)

func TestTSEventCreate(t *testing.T) {
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
	clusterRow, err := NewCluster(appContext).Create(nil, userRow, "cluster-name")
	if err != nil {
		t.Fatalf("Creating a cluster for user should work. Error: %v", err)
	}
	if clusterRow.ID <= 0 {
		t.Fatalf("Cluster ID should be assign properly. clusterRow.ID: %v", clusterRow.ID)
	}

	tsEvent := NewTSEvent(appContext, clusterRow.ID)

	pgdb, err = tsEvent.GetPGDB()
	if err != nil {
		t.Fatalf("There should be a legit db. Error: %v", err)
	}
	defer pgdb.Close()

	// Create TSEvent without passing dates
	tsEventRow, err := tsEvent.Create(nil, c.NewExplicitID(), clusterRow.ID, -1, -1, "Launched uber feature", time.Now().Unix()+int64(900))
	if err != nil {
		t.Fatalf("Creating a TSEvent should work. Error: %v", err)
	}
	if tsEventRow.CreatedFrom != tsEventRow.CreatedTo {
		t.Fatalf("From and to should be the same")
	}
	_, err = tsEvent.DeleteByID(nil, tsEventRow.ID)
	if err != nil {
		t.Fatalf("Deleting by id should not fail. Error: %v", err)
	}

	// Create TSEvent with from timestamp only
	tsEventRow, err = tsEvent.Create(nil, c.NewExplicitID(), clusterRow.ID, time.Now().Unix(), -1, "Launched uber feature 2", time.Now().Unix()+int64(900))
	if err != nil {
		t.Fatalf("Creating a TSEvent should work. Error: %v", err)
	}
	if tsEventRow.CreatedFrom != tsEventRow.CreatedTo {
		t.Fatalf("From and to should be the same")
	}
	_, err = tsEvent.DeleteByID(nil, tsEventRow.ID)
	if err != nil {
		t.Fatalf("Deleting by id should not fail. Error: %v", err)
	}

	// Create TSEvent with from and to timestamps only
	tsEventRow, err = tsEvent.Create(nil, c.NewExplicitID(), clusterRow.ID, time.Now().Unix(), time.Now().Unix(), "Launched uber feature 3", time.Now().Unix()+int64(900))
	if err != nil {
		t.Fatalf("Creating a TSEvent should work. Error: %v", err)
	}
	_, err = tsEvent.DeleteByID(nil, tsEventRow.ID)
	if err != nil {
		t.Fatalf("Deleting by id should not fail. Error: %v", err)
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
