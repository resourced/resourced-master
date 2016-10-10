package check_expression

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/satori/go.uuid"

	"github.com/resourced/resourced-master/libunix"
	"github.com/resourced/resourced-master/models/pg"
	"github.com/resourced/resourced-master/models/shared"
)

func newDbForTest(t *testing.T) *sqlx.DB {
	u, err := libunix.CurrentUser()
	if err != nil {
		t.Fatalf("Getting current user should never fail. Error: %v", err)
	}

	db, err := sqlx.Connect("postgres", fmt.Sprintf("postgres://%v@localhost:5432/resourced-master-test?sslmode=disable", u))
	if err != nil {
		t.Fatalf("Connecting to local postgres should never fail. Error: %v", err)
	}
	return db
}

func newEmailForTest() string {
	return fmt.Sprintf("brotato-%v@example.com", uuid.NewV4().String())
}

func newUserForTest(t *testing.T) *pg.User {
	return pg.NewUser(shared.AppContextForTest())
}

func newClusterForTest(t *testing.T) *pg.Cluster {
	return pg.NewCluster(newDbForTest(t))
}

func newAccessTokenForTest(t *testing.T) *pg.AccessToken {
	return pg.NewAccessToken(newDbForTest(t))
}

func newHostForTest(t *testing.T) *pg.Host {
	return pg.NewHost(newDbForTest(t))
}

func newMetricForTest(t *testing.T) *pg.Metric {
	return pg.NewMetric(newDbForTest(t))
}

func newTSMetricForTest(t *testing.T) *pg.TSMetric {
	return pg.NewTSMetric(newDbForTest(t))
}

func newTSLogForTest(t *testing.T) *pg.TSLog {
	return pg.NewTSLog(newDbForTest(t))
}

func newCheckForTest(t *testing.T) *pg.Check {
	return pg.NewCheck(newDbForTest(t))
}

func checkHostExpressionSetupForTest(t *testing.T) map[string]interface{} {
	setupRows := make(map[string]interface{})

	hostname, _ := os.Hostname()

	// Signup
	u := newUserForTest(t)

	userRow, err := u.Signup(nil, newEmailForTest(), "abc123", "abc123")
	if err != nil {
		t.Fatalf("Signing up user should work. Error: %v", err)
	}
	setupRows["userRow"] = userRow

	// Create cluster for user
	c := newClusterForTest(t)

	clusterRow, err := c.Create(nil, userRow, "cluster-name")
	if err != nil {
		t.Fatalf("Creating a cluster for user should work. Error: %v", err)
	}
	setupRows["clusterRow"] = clusterRow

	// Create access token
	at := newAccessTokenForTest(t)

	tokenRow, err := at.Create(nil, userRow.ID, clusterRow.ID, "write")
	if err != nil {
		t.Fatalf("Creating a token should work. Error: %v", err)
	}
	setupRows["tokenRow"] = tokenRow

	// Create host
	h := newHostForTest(t)

	hostRow, err := h.CreateOrUpdate(nil, tokenRow, []byte(fmt.Sprintf(`{"Host": {"Name": "%v", "Tags": {"aaa": "bbb"}}, "Data": {"/stuff": {"Score": 100}}}`, hostname)))
	if err != nil {
		t.Errorf("Creating a new host should work. Error: %v", err)
	}
	setupRows["hostRow"] = hostRow

	// Create Metric
	m := newMetricForTest(t)

	metricRow, err := m.CreateOrUpdate(nil, clusterRow.ID, "/stuff.Score")
	if err != nil {
		t.Fatalf("Creating a Metric should work. Error: %v", err)
	}
	setupRows["metricRow"] = metricRow

	// Create TSMetric
	tsm := newTSMetricForTest(t)

	err = tsm.Create(nil, clusterRow.ID, metricRow.ID, hostname, "/stuff.Score", float64(100), time.Now().Unix()+int64(900))
	if err != nil {
		t.Fatalf("Creating a TSMetric should work. Error: %v", err)
	}

	return setupRows
}

func checkHostExpressionTeardownForTest(t *testing.T, setupRows map[string]interface{}) {
	// DELETE FROM hosts WHERE id=...
	h := newHostForTest(t)

	_, err := h.DeleteByID(nil, setupRows["hostRow"].(*pg.HostRow).ID)
	if err != nil {
		t.Fatalf("Deleting access_tokens by id should not fail. Error: %v", err)
	}

	// DELETE FROM access_tokens WHERE id=...
	at := newAccessTokenForTest(t)

	_, err = at.DeleteByID(nil, setupRows["tokenRow"].(*pg.AccessTokenRow).ID)
	if err != nil {
		t.Fatalf("Deleting access_tokens by id should not fail. Error: %v", err)
	}

	// DELETE FROM metrics WHERE id=...
	m := newMetricForTest(t)

	_, err = m.DeleteByID(nil, setupRows["metricRow"].(*pg.MetricRow).ID)
	if err != nil {
		t.Fatalf("Deleting access_tokens by id should not fail. Error: %v", err)
	}

	// DELETE FROM clusters WHERE id=...
	c := newClusterForTest(t)

	_, err = c.DeleteByID(nil, setupRows["clusterRow"].(*pg.ClusterRow).ID)
	if err != nil {
		t.Fatalf("Deleting ClusterRow by id should not fail. Error: %v", err)
	}

	// DELETE FROM users WHERE id=...
	u := newUserForTest(t)

	_, err = u.DeleteByID(nil, setupRows["userRow"].(*pg.UserRow).ID)
	if err != nil {
		t.Fatalf("Deleting user by id should not fail. Error: %v", err)
	}
}

func TestCheckEvalRawHostDataExpression(t *testing.T) {
	setupRows := checkHostExpressionSetupForTest(t)

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

	chk := newCheckForTest(t)

	checkRow, err := chk.Create(nil, setupRows["clusterRow"].(*pg.ClusterRow).ID, data)
	if err != nil {
		t.Fatalf("Creating a Check should not fail. Error: %v", err)
	}
	if checkRow.ID <= 0 {
		t.Fatalf("Check ID should be assign properly. CheckRow.ID: %v", checkRow.ID)
	}

	hosts := make([]*pg.HostRow, 1)
	hosts[0] = setupRows["hostRow"].(*pg.HostRow)

	evaluator := CheckExpressionEvaluator{}

	// EvalRawHostDataExpression where hosts list is nil
	// The result should be true, which means that this expression is a fail.
	expression := evaluator.EvalRawHostDataExpression(checkRow, hosts, pg.CheckExpression{})
	if expression.Result.Value != true {
		t.Fatalf("Expression result is not true")
	}

	// EvalRawHostDataExpression where hosts list is not nil and valid expression.
	// This is a basic happy path test.
	// Host data is 100, and we set check to fail at 101.
	// The result should be false, 100 is not greater than 101, which means that this expression does not fail.
	expression = pg.CheckExpression{}
	expression.Metric = "/stuff.Score"
	expression.Operator = ">"
	expression.MinHost = 1
	expression.Value = float64(101)

	expression = evaluator.EvalRawHostDataExpression(checkRow, hosts, expression)
	if expression.Result.Value != false {
		t.Fatalf("Expression result is not false. %v %v %v", setupRows["hostRow"].(*pg.HostRow).DataAsFlatKeyValue()["/stuff"]["Score"], expression.Operator, expression.Value)
	}

	// Arithmetic expression test.
	// Host data is 100, see if the arithmetic comparison works.
	// Greater than is already tested, above.
	expression = pg.CheckExpression{}
	expression.Metric = "/stuff.Score"
	expression.Operator = ">="
	expression.MinHost = 1
	expression.Value = float64(100)

	expression = evaluator.EvalRawHostDataExpression(checkRow, hosts, expression)
	if expression.Result.Value != true {
		// The result should be true, 100 is greater than or equal to 100, which means that this expression fails.
		t.Fatalf("Expression result should be true. %v %v %v", setupRows["hostRow"].(*pg.HostRow).DataAsFlatKeyValue()["/stuff"]["Score"], expression.Operator, expression.Value)
	}

	expression = pg.CheckExpression{}
	expression.Metric = "/stuff.Score"
	expression.Operator = "="
	expression.MinHost = 1
	expression.Value = float64(100)

	expression = evaluator.EvalRawHostDataExpression(checkRow, hosts, expression)
	if expression.Result.Value != true {
		// The result should be true, 100 is equal to 100, which means that this expression fails.
		t.Fatalf("Expression result should be true. %v %v %v", setupRows["hostRow"].(*pg.HostRow).DataAsFlatKeyValue()["/stuff"]["Score"], expression.Operator, expression.Value)
	}

	expression = pg.CheckExpression{}
	expression.Metric = "/stuff.Score"
	expression.Operator = "<"
	expression.MinHost = 1
	expression.Value = float64(200)

	expression = evaluator.EvalRawHostDataExpression(checkRow, hosts, expression)
	if expression.Result.Value != true {
		// The result should be true, 100 is less than 200, which means that this expression fails.
		t.Fatalf("Expression result should be true. %v %v %v", setupRows["hostRow"].(*pg.HostRow).DataAsFlatKeyValue()["/stuff"]["Score"], expression.Operator, expression.Value)
	}

	expression = pg.CheckExpression{}
	expression.Metric = "/stuff.Score"
	expression.Operator = "<="
	expression.MinHost = 1
	expression.Value = float64(100)

	expression = evaluator.EvalRawHostDataExpression(checkRow, hosts, expression)
	if expression.Result.Value != true {
		// The result should be true, 100 is less than or equal to 100, which means that this expression fails.
		t.Fatalf("Expression result should be true. %v %v %v", setupRows["hostRow"].(*pg.HostRow).DataAsFlatKeyValue()["/stuff"]["Score"], expression.Operator, expression.Value)
	}

	// Case when host does not contain a particular metric.
	expression = pg.CheckExpression{}
	expression.Metric = "/stuff.DoesNotExist"
	expression.Operator = "<="
	expression.MinHost = 1
	expression.Value = float64(100)

	expression = evaluator.EvalRawHostDataExpression(checkRow, hosts, expression)
	if expression.Result.Value != true {
		// The result should be true
		// If we cannot find metric on the host, then we assume there's something wrong with the host. Thus, this expression must fails.
		t.Fatalf("Expression result should be true")
	}

	// ----------------------------------------------------------------------

	// DELETE FROM Checks WHERE id=...
	_, err = chk.DeleteByID(nil, checkRow.ID)
	if err != nil {
		t.Fatalf("Deleting Checks by id should not fail. Error: %v", err)
	}

	checkHostExpressionTeardownForTest(t, setupRows)
}

func TestCheckEvalRelativeHostDataExpression(t *testing.T) {
	setupRows := checkHostExpressionSetupForTest(t)

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

	chk := newCheckForTest(t)

	checkRow, err := chk.Create(nil, setupRows["clusterRow"].(*pg.ClusterRow).ID, data)
	if err != nil {
		t.Fatalf("Creating a Check should not fail. Error: %v", err)
	}
	if checkRow.ID <= 0 {
		t.Fatalf("Check ID should be assign properly. CheckRow.ID: %v", checkRow.ID)
	}

	hosts := make([]*pg.HostRow, 1)
	hosts[0] = setupRows["hostRow"].(*pg.HostRow)

	evaluator := CheckExpressionEvaluator{}

	// EvalRelativeHostDataExpression where hosts list is nil
	// The result should be true, which means that this expression is a fail.
	expression := evaluator.EvalRelativeHostDataExpression(checkRow, nil, pg.CheckExpression{})
	if expression.Result.Value != true {
		t.Fatalf("Expression result is not true")
	}

	// EvalRelativeHostDataExpression where hosts list is not nil and valid expression.
	// This is a basic happy path test.
	// The current host data is 100, and we set check to fail at 200% greater than max value of the last 3 minutes.
	expression = pg.CheckExpression{}
	expression.Metric = "/stuff.Score"
	expression.Operator = ">"
	expression.MinHost = 1
	expression.Value = float64(200)
	expression.PrevAggr = "max"
	expression.PrevRange = 3

	expression = evaluator.EvalRelativeHostDataExpression(checkRow, hosts, expression)
	if expression.Result.Value != false {
		// The result should be false. The max value of the last 3 minutes is 100, and 100 is not greater than 100 by more than 200%, which means that this expression does not fail.
		t.Fatalf("Expression result is not false. %v %v %v", setupRows["hostRow"].(*pg.HostRow).DataAsFlatKeyValue()["/stuff"]["Score"], expression.Operator, expression.Value)
	}

	// Arithmetic relative expression test.
	// Create a tiny value for TSMetric
	tsm := newTSMetricForTest(t)

	err = tsm.Create(nil, setupRows["clusterRow"].(*pg.ClusterRow).ID, setupRows["metricRow"].(*pg.MetricRow).ID, hostname, "/stuff.Score", float64(10), time.Now().Unix()+int64(900))
	if err != nil {
		t.Fatalf("Creating a TSMetric should work. Error: %v", err)
	}

	// The min of host data (range: 3 minutes ago from now) is 10, see if the arithmetic comparison works.
	// The current host data is 100, and we set check to fail at 200% greater than min value of the last 3 minutes.
	expression = pg.CheckExpression{}
	expression.Metric = "/stuff.Score"
	expression.Operator = ">"
	expression.MinHost = 1
	expression.Value = float64(200)
	expression.PrevAggr = "min"
	expression.PrevRange = 3

	expression = evaluator.EvalRelativeHostDataExpression(checkRow, hosts, expression)
	if expression.Result.Value != true {
		// The result should be true. The min value of the last 3 minutes is 10, and 100 is greater than 10 by more than 200%, which means that this expression must fail.
		t.Fatalf("Expression result should be true. %v %v %v", setupRows["hostRow"].(*pg.HostRow).DataAsFlatKeyValue()["/stuff"]["Score"], expression.Operator, expression.Value)
	}

	// Case when host does not contain a particular metric.
	expression = pg.CheckExpression{}
	expression.Metric = "/stuff.DoesNotExist"
	expression.Operator = ">"
	expression.MinHost = 1
	expression.Value = float64(200)

	expression = evaluator.EvalRelativeHostDataExpression(checkRow, hosts, expression)
	if expression.Result.Value != true {
		// The result should be true
		// If we cannot find metric on the host, then we assume there's something wrong with the host. Thus, this expression must fail.
		t.Fatalf("Expression result should be true")
	}

	// ----------------------------------------------------------------------

	// DELETE FROM Checks WHERE id=...
	_, err = chk.DeleteByID(nil, checkRow.ID)
	if err != nil {
		t.Fatalf("Deleting Checks by id should not fail. Error: %v", err)
	}

	checkHostExpressionTeardownForTest(t, setupRows)
}

func TestCheckEvalLogDataExpression(t *testing.T) {
	setupRows := checkHostExpressionSetupForTest(t)

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

	chk := newCheckForTest(t)

	// Create a Check
	checkRow, err := chk.Create(nil, setupRows["clusterRow"].(*pg.ClusterRow).ID, data)
	if err != nil {
		t.Fatalf("Creating a Check should not fail. Error: %v", err)
	}
	if checkRow.ID <= 0 {
		t.Fatalf("Check ID should be assign properly. CheckRow.ID: %v", checkRow.ID)
	}

	// Create TSLog
	dataJSONString := fmt.Sprintf(`{"Host": {"Name": "%v", "Tags": {}}, "Data": {"Filename":"", "Loglines": [{"Created": %v, "Content": "aaa"}, {"Created": %v, "Content": "bbb"}]}}`, hostname, time.Now().Unix()+int64(900), time.Now().Unix()+int64(900))

	err = newTSLogForTest(t).CreateFromJSON(nil, setupRows["clusterRow"].(*pg.ClusterRow).ID, []byte(dataJSONString), time.Now().Unix()+int64(900))
	if err != nil {
		t.Fatalf("Creating a TSLog should work. Error: %v", err)
	}

	db := newDbForTest(t)
	defer db.Close()

	// EvalLogDataExpression with valid expression.
	// This is a basic happy path test.
	// Count of log lines containing aaa should be > 0
	expression := pg.CheckExpression{}
	expression.Search = "aaa"
	expression.Operator = ">"
	expression.Value = float64(0)
	expression.MinHost = 1
	expression.PrevRange = 15

	evaluator := CheckExpressionEvaluator{}

	expression = evaluator.EvalLogDataExpression(checkRow, nil, expression)
	if expression.Result.Value != true {
		// The result should be true. The count of logs with "aaa" string is indeed greater than 0, which means that this expression must fail.
		t.Fatalf("Counting log lines containing %v is expected to be %v %v. Value: %v", expression.Search, expression.Operator, expression.Value, expression.Result.Value)
	}

	// ----------------------------------------------------------------------

	// DELETE FROM Checks WHERE id=...
	_, err = chk.DeleteByID(nil, checkRow.ID)
	if err != nil {
		t.Fatalf("Deleting Checks by id should not fail. Error: %v", err)
	}

	checkHostExpressionTeardownForTest(t, setupRows)
}

func TestCheckEvalPingExpression(t *testing.T) {
	setupRows := checkHostExpressionSetupForTest(t)

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

	chk := newCheckForTest(t)

	// Create a Check
	checkRow, err := chk.Create(nil, setupRows["clusterRow"].(*pg.ClusterRow).ID, data)
	if err != nil {
		t.Fatalf("Creating a Check should not fail. Error: %v", err)
	}
	if checkRow.ID <= 0 {
		t.Fatalf("Check ID should be assign properly. CheckRow.ID: %v", checkRow.ID)
	}

	// EvalPingExpression with valid expression.
	// This is a basic happy path test.
	expression := pg.CheckExpression{}
	expression.MinHost = 1

	evaluator := CheckExpressionEvaluator{}

	expression = evaluator.EvalPingExpression(checkRow, nil, expression)
	if expression.Result.Value != false {
		// The result should be false. The ping check should not fail since it is pinging localhost, which means that this expression must not fail.
		t.Fatalf("Expression should not fail")
	}

	// ----------------------------------------------------------------------

	// DELETE FROM Checks WHERE id=...
	_, err = chk.DeleteByID(nil, checkRow.ID)
	if err != nil {
		t.Fatalf("Deleting Checks by id should not fail. Error: %v", err)
	}

	checkHostExpressionTeardownForTest(t, setupRows)
}

func TestCheckEvalSSHExpression(t *testing.T) {
	setupRows := checkHostExpressionSetupForTest(t)

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

	chk := newCheckForTest(t)

	// Create a Check
	checkRow, err := chk.Create(nil, setupRows["clusterRow"].(*pg.ClusterRow).ID, data)
	if err != nil {
		t.Fatalf("Creating a Check should not fail. Error: %v", err)
	}
	if checkRow.ID <= 0 {
		t.Fatalf("Check ID should be assign properly. CheckRow.ID: %v", checkRow.ID)
	}

	// EvalSSHExpression with valid expression.
	// This is a basic happy path test.
	expression := pg.CheckExpression{}
	expression.MinHost = 1
	expression.Port = "22"

	evaluator := CheckExpressionEvaluator{}

	expression = evaluator.EvalSSHExpression(checkRow, nil, expression)
	if expression.Result.Value != false {
		// The result should be false. The check should not fail since it performs SSH to localhost, which means that this expression must not fail.
		t.Fatalf("Expression should not fail")
	}

	// ----------------------------------------------------------------------

	// DELETE FROM Checks WHERE id=...
	_, err = chk.DeleteByID(nil, checkRow.ID)
	if err != nil {
		t.Fatalf("Deleting Checks by id should not fail. Error: %v", err)
	}

	checkHostExpressionTeardownForTest(t, setupRows)
}

// TestCheckEvalHTTPExpression
// Make sure you run Master daemon on port 55655 for this test to pass.
func TestCheckEvalHTTPExpression(t *testing.T) {
	setupRows := checkHostExpressionSetupForTest(t)

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

	chk := newCheckForTest(t)

	// Create a Check
	checkRow, err := chk.Create(nil, setupRows["clusterRow"].(*pg.ClusterRow).ID, data)
	if err != nil {
		t.Fatalf("Creating a Check should not fail. Error: %v", err)
	}
	if checkRow.ID <= 0 {
		t.Fatalf("Check ID should be assign properly. CheckRow.ID: %v", checkRow.ID)
	}

	// EvalHTTPExpression with valid expression.
	// This is a basic happy path test.
	expression := pg.CheckExpression{}
	expression.MinHost = 1
	expression.Port = "55655"
	expression.Protocol = "http"
	expression.HTTPMethod = "GET"

	evaluator := CheckExpressionEvaluator{}

	expression = evaluator.EvalHTTPExpression(checkRow, nil, expression)
	if expression.Result.Value != false {
		// To make this test pass, you should run Master daemon on localhost:55655
		t.Fatalf("Expression should not fail checking HTTP on port %v. To make this test pass, you should run Master daemon on localhost:55655", expression.Port)
	}

	// ----------------------------------------------------------------------

	// DELETE FROM Checks WHERE id=...
	_, err = chk.DeleteByID(nil, checkRow.ID)
	if err != nil {
		t.Fatalf("Deleting Checks by id should not fail. Error: %v", err)
	}

	checkHostExpressionTeardownForTest(t, setupRows)
}
