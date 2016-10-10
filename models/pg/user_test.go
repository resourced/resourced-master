package pg

import (
	"testing"

	_ "github.com/lib/pq"

	"github.com/resourced/resourced-master/models/shared"
)

func newUserForTest(t *testing.T) *User {
	return NewUser(shared.AppContextForTest())
}

func TestUserCRUD(t *testing.T) {
	u := newUserForTest(t)
	defer u.db.Close()

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

	// DELETE FROM users WHERE id=...
	_, err = u.DeleteByID(nil, userRow.ID)
	if err != nil {
		t.Fatalf("Deleting user by id should not fail. Error: %v", err)
	}

}
