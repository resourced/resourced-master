package dal

import (
	"fmt"
	_ "github.com/lib/pq"
	"github.com/satori/go.uuid"
	"testing"
)

func newEmailForTest() string {
	return fmt.Sprintf("brotato-%v@example.com", uuid.NewV4().String())
}

func newUserForTest(t *testing.T) *User {
	return NewUser(newDbForTest(t))
}

func TestUserCRUD(t *testing.T) {
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

	// Create application for user
	appName := "brotato-test-app"
	appRow, err := u.CreateApplicationRow(nil, userRow.ID, appName)
	if err != nil {
		t.Fatalf("Creating an application for user should work. Error: %v", err)
	}
	if appRow.ID <= 0 {
		t.Fatalf("Application ID should be assign properly. appRow.ID: %v", appRow.ID)
	}

	// DELETE FROM users WHERE id=...
	_, err = u.DeleteById(nil, userRow.ID)
	if err != nil {
		t.Fatalf("Deleting user by id should not fail. Error: %v", err)
	}

	// DELETE FROM applications WHERE id=...
	_, err = NewApplication(u.db).DeleteById(nil, userRow.ApplicationID.Int64)
	if err != nil {
		t.Fatalf("Deleting application by id should not fail. Error: %v", err)
	}

}
