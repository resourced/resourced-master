package dal

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"os/user"
	"testing"
)

func newDbForTest(t *testing.T) *sqlx.DB {
	u, err := user.Current()
	if err != nil {
		t.Fatalf("Getting current user should never fail. Error: %v", err)
	}

	db, err := sqlx.Connect("postgres", fmt.Sprintf("postgres://%v@localhost:5432/resourced-master-test?sslmode=disable", u.Username))
	if err != nil {
		t.Fatalf("Connecting to local postgres should never fail. Error: %v", err)
	}
	return db
}

func newBaseForTest(t *testing.T) *Base {
	base := &Base{}
	base.db = newDbForTest(t)

	return base
}

func TestNewTransactionIfNeeded(t *testing.T) {
	base := newBaseForTest(t)

	// New Transaction block
	tx, wrapInSingleTransaction, err := base.newTransactionIfNeeded(nil)
	if err != nil {
		t.Fatalf("Creating new transaction block should not fail. Error: %v", err)
	}
	if wrapInSingleTransaction != true {
		t.Fatalf("Creating new transaction block should set wrapInSingleTransaction == true.")
	}
	if tx == nil {
		t.Fatalf("Creating new transaction block should not fail. Error: %v", err)
	}

	// Existing Transaction block
	tx2, wrapInSingleTransaction, err := base.newTransactionIfNeeded(tx)
	if err != nil {
		t.Fatalf("Receiving existing transaction block should not fail. Error: %v", err)
	}
	if wrapInSingleTransaction != false {
		t.Fatalf("Receiving existing transaction block should set wrapInSingleTransaction == false.")
	}
	if tx2 == nil {
		t.Fatalf("Receiving existing transaction block should not fail. Error: %v", err)
	}
	if tx2 != tx {
		t.Fatalf("Receiving existing transaction block should not fail. Error: %v", err)
	}
}

func TestCreateDeleteGeneric(t *testing.T) {
	base := newBaseForTest(t)
	base.table = "applications"

	// INSERT INTO applications (name) VALUES (...)
	data := make(map[string]interface{})
	data["name"] = "testing-app"

	result, err := base.InsertIntoTable(nil, data)
	if err != nil {
		t.Fatalf("Inserting new application should not fail. Error: %v", err)
	}

	lastInsertedId, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Inserting new application should not fail. Error: %v", err)
	}

	// DELETE FROM applications WHERE id=...
	where := fmt.Sprintf("id=%v", lastInsertedId)

	_, err = base.DeleteFromTable(nil, where)
	if err != nil {
		t.Fatalf("Deleting application by id should not fail. Error: %v", err)
	}

}

func TestCreateDeleteById(t *testing.T) {
	base := newBaseForTest(t)
	base.table = "applications"

	// INSERT INTO applications (name) VALUES (...)
	data := make(map[string]interface{})
	data["name"] = "testing-app"

	result, err := base.InsertIntoTable(nil, data)
	if err != nil {
		t.Fatalf("Inserting new application should not fail. Error: %v", err)
	}

	lastInsertedId, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Inserting new application should not fail. Error: %v", err)
	}

	// DELETE FROM applications WHERE id=...
	_, err = base.DeleteById(nil, lastInsertedId)
	if err != nil {
		t.Fatalf("Deleting application by id should not fail. Error: %v", err)
	}

}

func TestCreateUpdateGenericDelete(t *testing.T) {
	base := newBaseForTest(t)
	base.table = "applications"

	// INSERT INTO applications (name) VALUES (...)
	data := make(map[string]interface{})
	data["name"] = "testing-appz"

	result, err := base.InsertIntoTable(nil, data)
	if err != nil {
		t.Fatalf("Inserting new application should not fail. Error: %v", err)
	}

	lastInsertedId, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Inserting new application should not fail. Error: %v", err)
	}

	// UPDATE applications SET name=$1 WHERE id=$2
	data["name"] = "testing-app"
	where := fmt.Sprintf("id=%v", lastInsertedId)

	_, err = base.UpdateFromTable(nil, data, where)
	if err != nil {
		t.Errorf("Updating existing application should not fail. Error: %v", err)
	}

	// DELETE FROM applications WHERE id=...
	_, err = base.DeleteById(nil, lastInsertedId)
	if err != nil {
		t.Fatalf("Deleting application by id should not fail. Error: %v", err)
	}

}

func TestCreateUpdateByIdDelete(t *testing.T) {
	base := newBaseForTest(t)
	base.table = "applications"

	// INSERT INTO applications (name) VALUES (...)
	data := make(map[string]interface{})
	data["name"] = "testing-appz"

	result, err := base.InsertIntoTable(nil, data)
	if err != nil {
		t.Fatalf("Inserting new application should not fail. Error: %v", err)
	}

	lastInsertedId, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Inserting new application should not fail. Error: %v", err)
	}

	// UPDATE applications SET name=$1 WHERE id=$2
	data["name"] = "testing-app"

	_, err = base.UpdateById(nil, data, lastInsertedId)
	if err != nil {
		t.Errorf("Updating existing application should not fail. Error: %v", err)
	}

	// DELETE FROM applications WHERE id=...
	_, err = base.DeleteById(nil, lastInsertedId)
	if err != nil {
		t.Fatalf("Deleting application by id should not fail. Error: %v", err)
	}

}
