package dal

import (
	_ "github.com/lib/pq"
	"testing"
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

	// Create token for user
	tokenRow, err := u.CreateAccessTokenRow(nil, userRow.ID, "execute")
	if err != nil {
		t.Fatalf("Creating a token for user should work. Error: %v", err)
	}

	// Create host
	hostRow, err := newHostForTest(t).CreateOrUpdate(nil, tokenRow.ID, []byte(`{"/stuff": {"Score": 100, "Host": {"Name": "localhost", "Tags": ["aaa", "bbb"]}}}`))
	if err != nil {
		t.Errorf("Creating a new host should work. Error: %v", err)
	}
	if hostRow.ID <= 0 {
		t.Fatalf("Host ID should be assign properly. hostRow.ID: %v", hostRow.ID)
	}

	// SELECT * FROM hosts
	_, err = newHostForTest(t).AllHosts(nil)
	if err != nil {
		t.Fatalf("Selecting all hosts should not fail. Error: %v", err)
	}

	// SELECT * FROM hosts WHERE id=...
	_, err = newHostForTest(t).GetById(nil, hostRow.ID)
	if err != nil {
		t.Fatalf("Selecting host by id should not fail. Error: %v", err)
	}

	// SELECT * FROM hosts WHERE name=...
	_, err = newHostForTest(t).GetByName(nil, hostRow.Name)
	if err != nil {
		t.Fatalf("Selecting host by name should not fail. Error: %v", err)
	}

	// DELETE FROM hosts WHERE id=...
	_, err = newHostForTest(t).DeleteById(nil, hostRow.ID)
	if err != nil {
		t.Fatalf("Deleting access_tokens by id should not fail. Error: %v", err)
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
