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

func TestCreateApplicationForUser(t *testing.T) {
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

	// Create application for user
	appName := "brotato-test-app"
	userRow, err = u.CreateApplication(nil, userRow.ID, appName)
	if err != nil {
		t.Fatalf("Creating an application for user should work. Error: %v", err)
	}
	if userRow.ApplicationID.Int64 <= 0 {
		t.Fatalf("Application ID should be assign properly. userRow.ApplicationID: %v", userRow.ApplicationID)
	}

	// DELETE FROM users WHERE id=...
	_, err = u.DeleteById(nil, userRow.ID)
	if err != nil {
		t.Fatalf("Deleting user by id should not fail. Error: %v", err)
	}

	// DELETE FROM applications WHERE id=...
	app := NewApplication(u.db)

	_, err = app.DeleteById(nil, userRow.ApplicationID.Int64)
	if err != nil {
		t.Fatalf("Deleting application by id should not fail. Error: %v", err)
	}

}
