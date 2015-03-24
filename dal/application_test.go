package dal

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"os/user"
	"testing"
)

func newApplicationForTest(t *testing.T) *Application {
	u, err := user.Current()
	if err != nil {
		t.Fatalf("Getting current user should never fail. Error: %v", err)
	}

	db, err := sqlx.Connect("postgres", fmt.Sprintf("postgres://%v@localhost:5432/resourced-master?sslmode=disable", u.Username))
	if err != nil {
		t.Fatalf("Connecting to local postgres should never fail. Error: %v", err)
	}

	return NewApplication(db)
}

func TestApplicationCRUD(t *testing.T) {
	app := newApplicationForTest(t)

	// INSERT INTO applications (name) VALUES (...)
	data := make(map[string]interface{})
	data["name"] = "testing-appz"

	result, err := app.InsertIntoTable(nil, data)
	if err != nil {
		t.Fatalf("Inserting new application should not fail. Error: %v", err)
	}

	lastInsertedId, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Inserting new application should not fail. Error: %v", err)
	}

	// UPDATE applications SET name=$1 WHERE id=$2
	data["name"] = "testing-app"

	_, err = app.UpdateById(nil, data, lastInsertedId)
	if err != nil {
		t.Errorf("Updating existing application should not fail. Error: %v", err)
	}

	// SELECT * FROM applications WHERE id=$1
	appRow, err := app.GetById(nil, lastInsertedId)
	if err != nil {
		t.Errorf("Getting existing application should not fail. Error: %v", err)
	}
	if appRow.Name != data["name"] {
		t.Errorf("Failed to get the right application.")
	}

	// DELETE FROM applications WHERE id=...
	_, err = app.DeleteById(nil, lastInsertedId)
	if err != nil {
		t.Fatalf("Deleting application by id should not fail. Error: %v", err)
	}

}
