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
	appRow, err := app.CreateRow(nil, "testing-appz")
	if err != nil {
		t.Fatalf("Creating new application should not fail. Error: %v", err)
	}

	// UPDATE applications SET name=$1 WHERE id=$2
	data := make(map[string]interface{})
	data["name"] = "testing-app"

	_, err = app.UpdateById(nil, data, appRow.ID)
	if err != nil {
		t.Errorf("Updating existing application should not fail. Error: %v", err)
	}

	// SELECT * FROM applications WHERE id=$1
	appRowFromDb, err := app.GetById(nil, appRow.ID)
	if err != nil {
		t.Errorf("Getting existing application should not fail. Error: %v", err)
	}
	if appRowFromDb.Name != "testing-app" {
		t.Errorf("Failed to get the right application. appRowFromDb.Name: %v", appRowFromDb.Name)
	}

	// DELETE FROM applications WHERE id=...
	_, err = app.DeleteById(nil, appRow.ID)
	if err != nil {
		t.Fatalf("Deleting application by id should not fail. Error: %v", err)
	}

}
