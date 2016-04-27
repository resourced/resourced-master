package dal

import (
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func newCheckForTest(t *testing.T) *Check {
	return NewCheck(newDbForTest(t))
}

func TestCheckCRUD(t *testing.T) {
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

	// Create cluster for user
	clusterRow, err := newClusterForTest(t).Create(nil, userRow.ID, "cluster-name")
	if err != nil {
		t.Fatalf("Creating a cluster for user should work. Error: %v", err)
	}
	if clusterRow.ID <= 0 {
		t.Fatalf("Cluster ID should be assign properly. clusterRow.ID: %v", clusterRow.ID)
	}

	hostname, _ := os.Hostname()

	// Create Check
	data := make(map[string]interface{})
	data["name"] = "check-name"
	data["interval"] = "60s"
	data["hosts_query"] = ""
	data["hosts_list"] = []byte("[\"" + hostname + "\"]")
	data["expressions"] = []byte("[]")
	data["triggers"] = []byte("[]")
	data["last_result_hosts"] = []byte("[]")
	data["last_result_expressions"] = []byte("[]")

	checkRow, err := newCheckForTest(t).Create(nil, clusterRow.ID, data)
	if err != nil {
		t.Fatalf("Creating a Check should not fail. Error: %v", err)
	}
	if checkRow.ID <= 0 {
		t.Fatalf("Check ID should be assign properly. CheckRow.ID: %v", checkRow.ID)
	}

	// DELETE FROM Checks WHERE id=...
	_, err = NewCheck(u.db).DeleteByID(nil, checkRow.ID)
	if err != nil {
		t.Fatalf("Deleting Checks by id should not fail. Error: %v", err)
	}

	// DELETE FROM users WHERE id=...
	_, err = u.DeleteByID(nil, userRow.ID)
	if err != nil {
		t.Fatalf("Deleting user by id should not fail. Error: %v", err)
	}
}

func TestCheckEvalRawHostDataExpression(t *testing.T) {
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

	// Create cluster for user
	clusterRow, err := newClusterForTest(t).Create(nil, userRow.ID, "cluster-name")
	if err != nil {
		t.Fatalf("Creating a cluster for user should work. Error: %v", err)
	}
	if clusterRow.ID <= 0 {
		t.Fatalf("Cluster ID should be assign properly. clusterRow.ID: %v", clusterRow.ID)
	}

	hostname, _ := os.Hostname()

	// Create Check
	data := make(map[string]interface{})
	data["name"] = "check-name"
	data["interval"] = "60s"
	data["hosts_query"] = ""
	data["hosts_list"] = []byte("[\"" + hostname + "\"]")
	data["expressions"] = []byte("[]")
	data["triggers"] = []byte("[]")
	data["last_result_hosts"] = []byte("[]")
	data["last_result_expressions"] = []byte("[]")

	checkRow, err := newCheckForTest(t).Create(nil, clusterRow.ID, data)
	if err != nil {
		t.Fatalf("Creating a Check should not fail. Error: %v", err)
	}
	if checkRow.ID <= 0 {
		t.Fatalf("Check ID should be assign properly. CheckRow.ID: %v", checkRow.ID)
	}

	// EvalRawHostDataExpression where hosts list is nil
	// The result should be true, which means that this expression is a fail.
	expression := checkRow.EvalRawHostDataExpression(nil, CheckExpression{})
	if expression.Result.Value != true {
		t.Fatalf("Expression result is not true")
	}

	at := newAccessTokenForTest(t)

	// Create access token
	tokenRow, err := at.Create(nil, userRow.ID, clusterRow.ID, "execute")
	if err != nil {
		t.Fatalf("Creating a token should work. Error: %v", err)
	}
	if tokenRow.ID <= 0 {
		t.Fatalf("AccessToken ID should be assign properly. tokenRow.ID: %v", tokenRow.ID)
	}

	// Create host
	hostRow, err := newHostForTest(t).CreateOrUpdate(nil, tokenRow, []byte(`{"/stuff": {"Data": {"Score": 100}, "Host": {"Name": "localhost", "Tags": {"aaa": "bbb"}}}}`))
	if err != nil {
		t.Errorf("Creating a new host should work. Error: %v", err)
	}
	if hostRow.ID <= 0 {
		t.Fatalf("Host ID should be assign properly. hostRow.ID: %v", hostRow.ID)
	}

	// EvalRawHostDataExpression where hosts list is not nil and valid expression.
	// This is a basic happy path test.
	// Host data is 100, and we set check to fail at 101.
	// The result should be false, 100 is not greater than 101, which means that this expression does not fail.
	hosts := make([]*HostRow, 1)
	hosts[0] = hostRow

	expression = CheckExpression{}
	expression.Metric = "/stuff.Score"
	expression.Operator = ">"
	expression.MinHost = 1
	expression.Value = float64(101)

	expression = checkRow.EvalRawHostDataExpression(hosts, expression)
	if expression.Result.Value != false {
		t.Fatalf("Expression result is not false. %v > %v", hostRow.DataAsFlatKeyValue()["/stuff"]["Score"], expression.Value)
	}

	// DELETE FROM hosts WHERE id=...
	_, err = newHostForTest(t).DeleteByID(nil, hostRow.ID)
	if err != nil {
		t.Fatalf("Deleting access_tokens by id should not fail. Error: %v", err)
	}

	// DELETE FROM access_tokens WHERE id=...
	_, err = at.DeleteByID(nil, tokenRow.ID)
	if err != nil {
		t.Fatalf("Deleting access_tokens by id should not fail. Error: %v", err)
	}

	// DELETE FROM Checks WHERE id=...
	_, err = NewCheck(u.db).DeleteByID(nil, checkRow.ID)
	if err != nil {
		t.Fatalf("Deleting Checks by id should not fail. Error: %v", err)
	}

	// DELETE FROM users WHERE id=...
	_, err = u.DeleteByID(nil, userRow.ID)
	if err != nil {
		t.Fatalf("Deleting user by id should not fail. Error: %v", err)
	}
}
