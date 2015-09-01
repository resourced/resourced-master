package dal

import (
	"testing"

	_ "github.com/lib/pq"
)

func newHostForTest(t *testing.T) *Host {
	return NewHost(newDbForTest(t))
}

func TestHostCRUD(t *testing.T) {
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

	cl := newClusterForTest(t)

	// Create cluster for user
	clusterRow, err := cl.Create(nil, userRow.ID, "cluster-name")
	if err != nil {
		t.Fatalf("Creating a cluster for user should work. Error: %v", err)
	}
	if clusterRow.ID <= 0 {
		t.Fatalf("Cluster ID should be assign properly. clusterRow.ID: %v", clusterRow.ID)
	}

	at := newAccessTokenForTest(t)

	// Create access token
	tokenRow, err := at.Create(nil, userRow.ID, clusterRow.ID, "execute")
	if err != nil {
		t.Fatalf("Creating a token should work. Error: %v", err)
	}
	if tokenRow.ID <= 0 {
		t.Fatalf("AccessToken ID should be assign properly. tokenRow.ID: %v", tokenRow.ID)
	}

	// Create host
	hostRow, err := newHostForTest(t).CreateOrUpdate(nil, tokenRow, []byte(`{"/stuff": {"Score": 100, "Host": {"Name": "localhost", "Tags": {"aaa": "bbb"}}}}`))
	if err != nil {
		t.Errorf("Creating a new host should work. Error: %v", err)
	}
	if hostRow.ID <= 0 {
		t.Fatalf("Host ID should be assign properly. hostRow.ID: %v", hostRow.ID)
	}

	// SELECT * FROM hosts
	_, err = newHostForTest(t).AllByClusterID(nil, tokenRow.ID)
	if err != nil {
		t.Fatalf("Selecting all hosts should not fail. Error: %v", err)
	}

	// SELECT * FROM hosts by query
	_, err = newHostForTest(t).AllByClusterIDAndQuery(nil, tokenRow.ID, `/stuff.Score = 100`)
	if err != nil {
		t.Fatalf("Selecting all hosts by query should not fail. Error: %v", err)
	}

	// SELECT * FROM hosts WHERE id=...
	_, err = newHostForTest(t).GetByID(nil, hostRow.ID)
	if err != nil {
		t.Fatalf("Selecting host by id should not fail. Error: %v", err)
	}

	// SELECT * FROM hosts WHERE name=...
	_, err = newHostForTest(t).GetByName(nil, hostRow.Name)
	if err != nil {
		t.Fatalf("Selecting host by name should not fail. Error: %v", err)
	}

	// DELETE FROM hosts WHERE id=...
	_, err = newHostForTest(t).DeleteByID(nil, hostRow.ID)
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
