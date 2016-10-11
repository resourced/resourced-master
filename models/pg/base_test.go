package pg

import (
	"fmt"
	"testing"

	"github.com/Sirupsen/logrus"
	_ "github.com/lib/pq"
	"github.com/satori/go.uuid"

	"github.com/resourced/resourced-master/models/shared"
)

func init() {
	logrus.SetLevel(logrus.ErrorLevel)
}

func newEmailForTest() string {
	return fmt.Sprintf("brotato-%v@example.com", uuid.NewV4().String())
}

func newBaseForTest(t *testing.T) *Base {
	base := &Base{}
	base.AppContext = shared.AppContextForTest()

	return base
}

func TestNewTransactionIfNeeded(t *testing.T) {
	base := newBaseForTest(t)

	pgdb, err := base.GetPGDB()
	if err != nil {
		t.Errorf("There should be a legit db. Error: %v", err)
	}
	defer pgdb.Close()

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
	base.table = "users"

	pgdb, err := base.GetPGDB()
	if err != nil {
		t.Errorf("There should be a legit db. Error: %v", err)
	}
	defer pgdb.Close()

	// INSERT INTO users (name) VALUES (...)
	data := make(map[string]interface{})
	data["email"] = newEmailForTest()
	data["password"] = "abc123"

	result, err := base.InsertIntoTable(nil, data)
	if err != nil {
		t.Fatalf("Inserting new user should not fail. Error: %v", err)
	}

	lastInsertedId, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Inserting new user should not fail. Error: %v", err)
	}

	// DELETE FROM users WHERE id=...
	where := fmt.Sprintf("id=%v", lastInsertedId)

	_, err = base.DeleteFromTable(nil, where)
	if err != nil {
		t.Fatalf("Deleting user by id should not fail. Error: %v", err)
	}

}

func TestCreateDeleteByID(t *testing.T) {
	base := newBaseForTest(t)
	base.table = "users"

	pgdb, err := base.GetPGDB()
	if err != nil {
		t.Errorf("There should be a legit db. Error: %v", err)
	}
	defer pgdb.Close()

	// INSERT INTO users (...) VALUES (...)
	data := make(map[string]interface{})
	data["email"] = newEmailForTest()
	data["password"] = "abc123"

	result, err := base.InsertIntoTable(nil, data)
	if err != nil {
		t.Fatalf("Inserting new user should not fail. Error: %v", err)
	}

	lastInsertedId, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Inserting new user should not fail. Error: %v", err)
	}

	// DELETE FROM users WHERE id=...
	_, err = base.DeleteByID(nil, lastInsertedId)
	if err != nil {
		t.Fatalf("Deleting user by id should not fail. Error: %v", err)
	}

}

func TestCreateUpdateGenericDelete(t *testing.T) {
	base := newBaseForTest(t)
	base.table = "users"

	pgdb, err := base.GetPGDB()
	if err != nil {
		t.Errorf("There should be a legit db. Error: %v", err)
	}
	defer pgdb.Close()

	// INSERT INTO users (...) VALUES (...)
	data := make(map[string]interface{})
	data["email"] = newEmailForTest()
	data["password"] = "abc123"

	result, err := base.InsertIntoTable(nil, data)
	if err != nil {
		t.Fatalf("Inserting new user should not fail. Error: %v", err)
	}

	lastInsertedId, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Inserting new user should not fail. Error: %v", err)
	}

	// UPDATE users SET email=$1 WHERE id=$2
	data["email"] = newEmailForTest()
	where := fmt.Sprintf("id=%v", lastInsertedId)

	_, err = base.UpdateFromTable(nil, data, where)
	if err != nil {
		t.Errorf("Updating existing user should not fail. Error: %v", err)
	}

	// DELETE FROM users WHERE id=...
	_, err = base.DeleteByID(nil, lastInsertedId)
	if err != nil {
		t.Fatalf("Deleting user by id should not fail. Error: %v", err)
	}

}

func TestCreateUpdateByIDDelete(t *testing.T) {
	base := newBaseForTest(t)
	base.table = "users"

	pgdb, err := base.GetPGDB()
	if err != nil {
		t.Errorf("There should be a legit db. Error: %v", err)
	}
	defer pgdb.Close()

	// INSERT INTO users (...) VALUES (...)
	data := make(map[string]interface{})
	data["email"] = newEmailForTest()
	data["password"] = "abc123"

	result, err := base.InsertIntoTable(nil, data)
	if err != nil {
		t.Fatalf("Inserting new user should not fail. Error: %v", err)
	}

	lastInsertedId, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Inserting new user should not fail. Error: %v", err)
	}

	// UPDATE users SET name=$1 WHERE id=$2
	data["email"] = newEmailForTest()

	_, err = base.UpdateByID(nil, data, lastInsertedId)
	if err != nil {
		t.Errorf("Updating existing user should not fail. Error: %v", err)
	}

	// DELETE FROM users WHERE id=...
	_, err = base.DeleteByID(nil, lastInsertedId)
	if err != nil {
		t.Fatalf("Deleting user by id should not fail. Error: %v", err)
	}

}

func TestCreateUpdateByKeyValueStringDelete(t *testing.T) {
	base := newBaseForTest(t)
	base.table = "users"

	pgdb, err := base.GetPGDB()
	if err != nil {
		t.Errorf("There should be a legit db. Error: %v", err)
	}
	defer pgdb.Close()

	originalEmail := newEmailForTest()

	// INSERT INTO users (...) VALUES (...)
	data := make(map[string]interface{})
	data["email"] = newEmailForTest()
	data["password"] = originalEmail

	result, err := base.InsertIntoTable(nil, data)
	if err != nil {
		t.Fatalf("Inserting new user should not fail. Error: %v", err)
	}

	lastInsertedId, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Inserting new user should not fail. Error: %v", err)
	}

	// UPDATE users SET name=$1 WHERE id=$2
	data["email"] = newEmailForTest()

	_, err = base.UpdateByKeyValueString(nil, data, "email", originalEmail)
	if err != nil {
		t.Errorf("Updating existing user should not fail. Error: %v", err)
	}

	// DELETE FROM users WHERE id=...
	_, err = base.DeleteByID(nil, lastInsertedId)
	if err != nil {
		t.Fatalf("Deleting user by id should not fail. Error: %v", err)
	}

}
