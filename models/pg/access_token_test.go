package pg

import (
	"testing"

	_ "github.com/lib/pq"

	"github.com/resourced/resourced-master/models/shared"
)

func TestAccessTokenCRUD(t *testing.T) {
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

	// SELECT * FROM access_tokens
	tokenRowFromDb, err := at.GetByClusterID(nil, clusterRow.ID)
	if err != nil {
		t.Fatalf("Selecting recently created access_tokens should not fail. Error: %v", err)
	}
	if tokenRow.ClusterID != tokenRowFromDb.ClusterID {
		t.Fatalf("Fetched the wrong access token")
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
