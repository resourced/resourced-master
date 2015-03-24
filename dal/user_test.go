package dal

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"os/user"
	"testing"
)

func newUserForTest(t *testing.T) *User {
	u, err := user.Current()
	if err != nil {
		t.Fatalf("Getting current user should never fail. Error: %v", err)
	}

	db, err := sqlx.Connect("postgres", fmt.Sprintf("postgres://%v@localhost:5432/resourced-master?sslmode=disable", u.Username))
	if err != nil {
		t.Fatalf("Connecting to local postgres should never fail. Error: %v", err)
	}

	return NewUser(db)
}

func TestUserCRUD(t *testing.T) {
	u := newUserForTest(t)

	// Signup
	userRow, err := u.Signup(nil, "brotato@example.com", "abc123")
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
	_, err = u.DeleteById(nil, userRow.ID)
	if err != nil {
		t.Fatalf("Deleting user by id should not fail. Error: %v", err)
	}

}

func TestAccessTokenCRUD(t *testing.T) {
	u := newUserForTest(t)
	appId := int64(1)

	// CreateAccessToken
	userRow, err := u.CreateAccessToken(nil, appId)
	if err != nil {
		t.Errorf("Creating access token should work. Error: %v", err)
	}
	if userRow == nil {
		t.Fatal("Creating access token should work.")
	}
	if userRow.ID <= 0 {
		t.Fatal("Creating access token should work.")
	}

	// DELETE FROM users WHERE id=...
	_, err = u.DeleteById(nil, userRow.ID)
	if err != nil {
		t.Fatalf("Deleting user by id should not fail. Error: %v", err)
	}

}
