package dal

import (
	_ "github.com/lib/pq"
	"testing"
)

func newApplicationForTest(t *testing.T) *Application {
	return NewApplication(newDbForTest(t))
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
