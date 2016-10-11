package pg

import (
	"testing"

	_ "github.com/lib/pq"

	"github.com/resourced/resourced-master/models/shared"
)

func TestHostCRUD(t *testing.T) {
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

	cl := NewCluster(appContext)

	// Create cluster for user
	clusterRow, err := cl.Create(nil, userRow, "cluster-name")
	if err != nil {
		t.Fatalf("Creating a cluster for user should work. Error: %v", err)
	}
	if clusterRow.ID <= 0 {
		t.Fatalf("Cluster ID should be assign properly. clusterRow.ID: %v", clusterRow.ID)
	}

	at := NewAccessToken(appContext)

	// Create access token
	tokenRow, err := at.Create(nil, userRow.ID, clusterRow.ID, "write")
	if err != nil {
		t.Fatalf("Creating a token should work. Error: %v", err)
	}
	if tokenRow.ID <= 0 {
		t.Fatalf("AccessToken ID should be assign properly. tokenRow.ID: %v", tokenRow.ID)
	}

	h := NewHost(appContext, clusterRow.ID)

	pgdb, err = h.GetPGDB()
	if err != nil {
		t.Fatalf("There should be a legit db. Error: %v", err)
	}
	defer pgdb.Close()

	// Create host
	hostRow, err := h.CreateOrUpdate(nil, tokenRow, []byte(`{"/stuff": {"Data": {"Score": 100}, "Host": {"Name": "localhost", "Tags": {"aaa": "bbb"}}}}`))
	if err != nil {
		t.Errorf("Creating a new host should work. Error: %v", err)
	}
	if hostRow.ID <= 0 {
		t.Fatalf("Host ID should be assign properly. hostRow.ID: %v", hostRow.ID)
	}

	// SELECT * FROM hosts
	hostRows, err := h.AllByClusterID(nil, clusterRow.ID)
	if err != nil {
		t.Fatalf("Selecting all hosts should not fail. Error: %v", err)
	}
	if len(hostRows) <= 0 {
		t.Fatalf("There should be at least one host. Hosts length: %v", len(hostRows))
	}

	hostRows, err = h.AllByClusterIDAndUpdatedInterval(nil, clusterRow.ID, "5 minutes")
	if err != nil {
		t.Fatalf("Getting all hosts should not fail. Error: %v", err)
	}
	if len(hostRows) <= 0 {
		t.Fatalf("Counting all hosts should not fail. Count: %v", hostRows)
	}

	// SELECT * FROM hosts by query
	_, err = h.AllCompactByClusterIDQueryAndUpdatedInterval(nil, clusterRow.ID, `/stuff.Score = 100`, "1h")
	if err != nil {
		t.Fatalf("Selecting all hosts by query should not fail. Error: %v", err)
	}

	_, err = h.AllByClusterIDQueryAndUpdatedInterval(nil, clusterRow.ID, `/stuff.Score = 100`, "5 minutes")
	if err != nil {
		t.Fatalf("Counting all hosts by query should not fail. Error: %v", err)
	}

	// SELECT * FROM hosts WHERE id=...
	_, err = h.GetByID(nil, hostRow.ID)
	if err != nil {
		t.Fatalf("Selecting host by id should not fail. Error: %v", err)
	}

	// SELECT * FROM hosts WHERE name=...
	_, err = h.GetByHostname(nil, hostRow.Hostname)
	if err != nil {
		t.Fatalf("Selecting host by name should not fail. Error: %v", err)
	}

	// DELETE FROM hosts WHERE id=...
	_, err = h.DeleteByID(nil, hostRow.ID)
	if err != nil {
		t.Fatalf("Deleting access_tokens by id should not fail. Error: %v", err)
	}

	// DELETE FROM access_tokens WHERE id=...
	_, err = at.DeleteByID(nil, tokenRow.ID)
	if err != nil {
		t.Fatalf("Deleting access_tokens by id should not fail. Error: %v", err)
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
