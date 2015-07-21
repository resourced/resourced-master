package dal

import (
	_ "github.com/lib/pq"
	"testing"
)

func newSavedQueryForTest(t *testing.T) *SavedQuery {
	return NewSavedQuery(newDbForTest(t))
}

func TestSavedQueryCRUD(t *testing.T) {
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

	// Create token for user
	tokenRow, err := u.CreateAccessTokenRow(nil, userRow.ID, "execute")
	if err != nil {
		t.Fatalf("Creating a token for user should work. Error: %v", err)
	}
	if tokenRow.ID <= 0 {
		t.Fatalf("AccessToken ID should be assign properly. tokenRow.ID: %v", tokenRow.ID)
	}

	sq := newSavedQueryForTest(t)

	// Create a new saved query
	sqRow, err := sq.CreateOrUpdate(nil, tokenRow.ID, "true")
	if err != nil {
		t.Fatalf("Failed to create saved query. Error: %v", err)
	}

	// Update existing saved query
	data := make(map[string]interface{})
	data["user_id"] = tokenRow.UserID
	data["query"] = "false"

	_, err = sq.UpdateByAccessTokenAndSavedQuery(nil, data, tokenRow, "true")
	if err != nil {
		t.Fatalf("Failed to update saved query. Error: %v", err)
	}

	// SELECT FROM saved_queries
	sqRowFromDb, err := sq.GetByAccessTokenAndQuery(nil, tokenRow, "false")
	if err != nil {
		t.Fatalf("Failed to get saved query. Error: %v", err)
	}
	if sqRow.ID != sqRowFromDb.ID {
		t.Fatalf("Failed to get the correct saved query.")
	}
	if sqRowFromDb.Query != "false" {
		t.Fatalf("Update did not work correctly. Query: %v", sqRowFromDb.Query)
	}

	// DELETE FROM saved_queries WHERE id=...
	_, err = sq.DeleteById(nil, sqRowFromDb.ID)
	if err != nil {
		t.Fatalf("Deleting saved_queries by id should not fail. Error: %v", err)
	}

}
