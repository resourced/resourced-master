package pg

import (
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"github.com/resourced/resourced-master/models/shared"
)

func checkHostExpressionSetupForTest(t *testing.T) map[string]interface{} {
	setupRows := make(map[string]interface{})

	hostname, _ := os.Hostname()

	appContext := shared.AppContextForTest()

	u := NewUser(appContext)

	pgdb, err := u.GetPGDB()
	if err != nil {
		t.Errorf("There should be a legit db. Error: %v", err)
	}
	defer pgdb.Close()

	// Signup
	userRow, err := u.Signup(nil, newEmailForTest(), "abc123", "abc123")
	if err != nil {
		t.Fatalf("Signing up user should work. Error: %v", err)
	}
	setupRows["userRow"] = userRow

	// Create cluster for user
	c := NewCluster(appContext)

	clusterRow, err := c.Create(nil, userRow, "cluster-name")
	if err != nil {
		t.Fatalf("Creating a cluster for user should work. Error: %v", err)
	}
	setupRows["clusterRow"] = clusterRow

	// Create access token
	at := NewAccessToken(appContext)

	tokenRow, err := at.Create(nil, userRow.ID, clusterRow.ID, "write")
	if err != nil {
		t.Fatalf("Creating a token should work. Error: %v", err)
	}
	setupRows["tokenRow"] = tokenRow

	// Create host
	h := NewHost(appContext, clusterRow.ID)

	hostRow, err := h.CreateOrUpdate(nil, tokenRow, []byte(fmt.Sprintf(`{"Host": {"Name": "%v", "Tags": {"aaa": "bbb"}}, "Data": {"/stuff": {"Score": 100}}}`, hostname)))
	if err != nil {
		t.Errorf("Creating a new host should work. Error: %v", err)
	}
	setupRows["hostRow"] = hostRow

	// Create Metric
	m := NewMetric(appContext)

	metricRow, err := m.CreateOrUpdate(nil, clusterRow.ID, "/stuff.Score")
	if err != nil {
		t.Fatalf("Creating a Metric should work. Error: %v", err)
	}
	setupRows["metricRow"] = metricRow

	// Create TSMetric
	tsm := NewTSMetric(appContext, clusterRow.ID)

	err = tsm.Create(nil, clusterRow.ID, metricRow.ID, hostname, "/stuff.Score", float64(100), time.Now().Unix()+int64(900))
	if err != nil {
		t.Fatalf("Creating a TSMetric should work. Error: %v", err)
	}

	return setupRows
}

func checkHostExpressionTeardownForTest(t *testing.T, setupRows map[string]interface{}) {
	appContext := shared.AppContextForTest()

	// DELETE FROM hosts WHERE id=...
	h := NewHost(appContext, setupRows["clusterRow"].(*ClusterRow).ID)

	_, err := h.DeleteByID(nil, setupRows["hostRow"].(*HostRow).ID)
	if err != nil {
		t.Fatalf("Deleting access_tokens by id should not fail. Error: %v", err)
	}

	// DELETE FROM access_tokens WHERE id=...
	at := NewAccessToken(appContext)

	_, err = at.DeleteByID(nil, setupRows["tokenRow"].(*AccessTokenRow).ID)
	if err != nil {
		t.Fatalf("Deleting access_tokens by id should not fail. Error: %v", err)
	}

	// DELETE FROM metrics WHERE id=...
	m := NewMetric(appContext)

	_, err = m.DeleteByID(nil, setupRows["metricRow"].(*MetricRow).ID)
	if err != nil {
		t.Fatalf("Deleting access_tokens by id should not fail. Error: %v", err)
	}

	// DELETE FROM clusters WHERE id=...
	c := NewCluster(appContext)

	_, err = c.DeleteByID(nil, setupRows["clusterRow"].(*ClusterRow).ID)
	if err != nil {
		t.Fatalf("Deleting ClusterRow by id should not fail. Error: %v", err)
	}

	// DELETE FROM users WHERE id=...
	u := NewUser(appContext)

	pgdb, err := u.GetPGDB()
	if err != nil {
		t.Errorf("There should be a legit db. Error: %v", err)
	}
	defer pgdb.Close()

	_, err = u.DeleteByID(nil, setupRows["userRow"].(*UserRow).ID)
	if err != nil {
		t.Fatalf("Deleting user by id should not fail. Error: %v", err)
	}
}

func TestCheckCRUD(t *testing.T) {
	appContext := shared.AppContextForTest()

	u := NewUser(appContext)

	pgdb, err := u.GetPGDB()
	if err != nil {
		t.Errorf("There should be a legit db. Error: %v", err)
	}
	defer pgdb.Close()

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
	c := NewCluster(appContext)

	clusterRow, err := c.Create(nil, userRow, "cluster-name")
	if err != nil {
		t.Fatalf("Creating a cluster for user should work. Error: %v", err)
	}
	if clusterRow.ID <= 0 {
		t.Fatalf("Cluster ID should be assign properly. clusterRow.ID: %v", clusterRow.ID)
	}

	// ----------------------------------------------------------------------
	// Real test begins here

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

	chk := NewCheck(appContext)

	checkRow, err := chk.Create(nil, clusterRow.ID, data)
	if err != nil {
		t.Fatalf("Creating a Check should not fail. Error: %v", err)
	}
	if checkRow.ID <= 0 {
		t.Fatalf("Check ID should be assign properly. CheckRow.ID: %v", checkRow.ID)
	}

	// DELETE FROM Checks WHERE id=...
	_, err = chk.DeleteByID(nil, checkRow.ID)
	if err != nil {
		t.Fatalf("Deleting Checks by id should not fail. Error: %v", err)
	}

	// ----------------------------------------------------------------------

	// DELETE FROM users WHERE id=...
	_, err = u.DeleteByID(nil, userRow.ID)
	if err != nil {
		t.Fatalf("Deleting user by id should not fail. Error: %v", err)
	}
}

func TestBuildEmailTriggerContent(t *testing.T) {
	setupRows := checkHostExpressionSetupForTest(t)

	hosts := make([]*HostRow, 1)
	hosts[0] = setupRows["hostRow"].(*HostRow)

	appContext := shared.AppContextForTest()

	// ----------------------------------------------------------------------
	// Real test begins here

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

	chk := NewCheck(appContext)

	// Create a Check
	checkRow, err := chk.Create(nil, setupRows["clusterRow"].(*ClusterRow).ID, data)
	if err != nil {
		t.Fatalf("Creating a Check should not fail. Error: %v", err)
	}

	expression := CheckExpression{}
	expression.Type = "RawHostData"
	expression.Metric = "/stuff.Score"
	expression.Operator = ">="
	expression.MinHost = 1
	expression.Value = float64(100)
	expression.Result.Value = true

	expressionResults := make([]CheckExpression, 1)
	expressionResults[0] = expression

	tsCheck := NewTSCheck(appContext, checkRow.ClusterID)

	err = tsCheck.Create(nil, checkRow.ClusterID, checkRow.ID, true, expressionResults, time.Now().Unix()+int64(900))
	if err != nil {
		t.Fatalf("Creating a TSCheck should not fail. Error: %v", err)
	}

	lastViolation, err := tsCheck.LastByClusterIDCheckIDAndResult(nil, checkRow.ClusterID, checkRow.ID, true)
	if err != nil {
		t.Fatalf("Fetching a TSCheck should not fail. Error: %v", err)
	}

	_, err = checkRow.BuildEmailTriggerContent(lastViolation, "$GOPATH/src/github.com/resourced/resourced-master")
	if err != nil {
		t.Fatalf("Generating the content of email alert should not fail. Error: %v", err)
	}

	// ----------------------------------------------------------------------

	// DELETE FROM Checks WHERE id=...
	_, err = chk.DeleteByID(nil, checkRow.ID)
	if err != nil {
		t.Fatalf("Deleting Checks by id should not fail. Error: %v", err)
	}

	checkHostExpressionTeardownForTest(t, setupRows)
}
