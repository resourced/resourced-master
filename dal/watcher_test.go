package dal

import (
	"testing"

	_ "github.com/lib/pq"
)

func newWatcherForTest(t *testing.T) *Watcher {
	return NewWatcher(newDbForTest(t))
}

func TestWatcherCRUD(t *testing.T) {
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

	at := newAccessTokenForTest(t)

	// Create access token
	tokenRow, err := at.Create(nil, userRow.ID, clusterRow.ID, "execute")
	if err != nil {
		t.Fatalf("Creating a token should work. Error: %v", err)
	}
	if tokenRow.ID <= 0 {
		t.Fatalf("AccessToken ID should be assign properly. tokenRow.ID: %v", tokenRow.ID)
	}

	sq := newSavedQueryForTest(t)

	// Create a new saved query
	sqRow, err := sq.CreateOrUpdate(nil, tokenRow, "true")
	if err != nil {
		t.Fatalf("Failed to create saved query. Error: %v", err)
	}

	// Create a watcher
	w := newWatcherForTest(t)

	data := w.CreateOrUpdateParameters(clusterRow.ID, sqRow.Query, "testing", 1, "5 minutes ago", "10s", nil)

	watcherRow, err := w.Create(nil, data)
	if err != nil {
		t.Errorf("Creating watcher should should work. Error: %v", err)
	}

	// All
	_, err = w.All(nil)
	if err != nil {
		t.Errorf("Fetching all watchers should should work. Error: %v", err)
	}

	// DELETE FROM watchers WHERE id=...
	_, err = w.DeleteByID(nil, watcherRow.ID)
	if err != nil {
		t.Fatalf("Deleting by id should not fail. Error: %v", err)
	}

	// DELETE FROM saved_queries WHERE id=...
	_, err = sq.DeleteByID(nil, sqRow.ID)
	if err != nil {
		t.Fatalf("Deleting saved_queries by id should not fail. Error: %v", err)
	}

	// DELETE FROM access_tokens WHERE id=...
	_, err = at.DeleteByID(nil, tokenRow.ID)
	if err != nil {
		t.Fatalf("Deleting access_tokens by id should not fail. Error: %v", err)
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

func TestWatcherSplitToDaemons(t *testing.T) {
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

	at := newAccessTokenForTest(t)

	// Create access token
	tokenRow, err := at.Create(nil, userRow.ID, clusterRow.ID, "execute")
	if err != nil {
		t.Fatalf("Creating a token should work. Error: %v", err)
	}
	if tokenRow.ID <= 0 {
		t.Fatalf("AccessToken ID should be assign properly. tokenRow.ID: %v", tokenRow.ID)
	}

	sq := newSavedQueryForTest(t)

	// Create a new saved query
	sqRow, err := sq.CreateOrUpdate(nil, tokenRow, "true")
	if err != nil {
		t.Fatalf("Failed to create saved query. Error: %v", err)
	}

	// Create a watcher
	w := newWatcherForTest(t)

	data := w.CreateOrUpdateParameters(clusterRow.ID, sqRow.Query, "testing", 1, "5 minutes ago", "10s", nil)

	watcherRow, err := w.Create(nil, data)
	if err != nil {
		t.Errorf("Creating watcher should should work. Error: %v", err)
	}
	watcher2Row, err := w.Create(nil, data)
	if err != nil {
		t.Errorf("Creating watcher should should work. Error: %v", err)
	}
	watcher3Row, err := w.Create(nil, data)
	if err != nil {
		t.Errorf("Creating watcher should should work. Error: %v", err)
	}

	// AllSplitToDaemons
	daemons := []string{"example.local", "example2.local"}
	byHostname, err := w.AllSplitToDaemons(nil, daemons)
	if err != nil {
		t.Errorf("Fetching watchers and grouping them by daemons should should work. Error: %v", err)
	}
	if len(byHostname) != len(daemons) {
		t.Errorf("Watchers were grouped incorrectly. Error: %v", err)
	}
	if len(byHostname[daemons[0]]) != 2 {
		t.Errorf("Watchers were grouped incorrectly")
	}
	if len(byHostname[daemons[1]]) != 1 {
		t.Errorf("Watchers were grouped incorrectly")
	}

	// DELETE FROM watchers WHERE id=...
	_, err = w.DeleteByID(nil, watcherRow.ID)
	if err != nil {
		t.Fatalf("Deleting by id should not fail. Error: %v", err)
	}
	_, err = w.DeleteByID(nil, watcher2Row.ID)
	if err != nil {
		t.Fatalf("Deleting by id should not fail. Error: %v", err)
	}
	_, err = w.DeleteByID(nil, watcher3Row.ID)
	if err != nil {
		t.Fatalf("Deleting by id should not fail. Error: %v", err)
	}

	// DELETE FROM saved_queries WHERE id=...
	_, err = sq.DeleteByID(nil, sqRow.ID)
	if err != nil {
		t.Fatalf("Deleting saved_queries by id should not fail. Error: %v", err)
	}

	// DELETE FROM access_tokens WHERE id=...
	_, err = at.DeleteByID(nil, tokenRow.ID)
	if err != nil {
		t.Fatalf("Deleting access_tokens by id should not fail. Error: %v", err)
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
