package dal

import (
	_ "github.com/lib/pq"
	"testing"
)

func newAccessTokenForTest(t *testing.T) *AccessToken {
	return NewAccessToken(newDbForTest(t))
}

func TestAccessTokenCRUD(t *testing.T) {
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

	// DELETE FROM access_tokens WHERE id=...
	_, err = NewAccessToken(u.db).DeleteById(nil, tokenRow.ID)
	if err != nil {
		t.Fatalf("Deleting access_tokens by id should not fail. Error: %v", err)
	}

	// DELETE FROM users WHERE id=...
	_, err = u.DeleteById(nil, userRow.ID)
	if err != nil {
		t.Fatalf("Deleting user by id should not fail. Error: %v", err)
	}
}
