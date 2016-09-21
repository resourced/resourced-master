package dal

import (
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func newTSLogForTest(t *testing.T) *TSLog {
	return NewTSLog(newDbForTest(t))
}

func TestTSLogCreateValue(t *testing.T) {
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
	clusterRow, err := newClusterForTest(t).Create(nil, userRow, "cluster-name")
	if err != nil {
		t.Fatalf("Creating a cluster for user should work. Error: %v", err)
	}
	if clusterRow.ID <= 0 {
		t.Fatalf("Cluster ID should be assign properly. clusterRow.ID: %v", clusterRow.ID)
	}

	hostname, _ := os.Hostname()

	// Create TSLog
	dataJSONString := fmt.Sprintf(`{"Host": {"Name": "%v", "Tags": {}}, "Data": {"Filename":"", "Loglines": [{"Created": 123, "Content": "aaa"}, {"Created": 123, "Content": "bbb"}]}}`, hostname)

	err = newTSLogForTest(t).CreateFromJSON(nil, clusterRow.ID, []byte(dataJSONString), time.Now().Unix()+int64(900))
	if err != nil {
		t.Fatalf("Creating a TSLog should work. Error: %v", err)
	}

	// DELETE FROM clusters WHERE id=...
	_, err = NewCluster(u.db).DeleteByID(nil, clusterRow.ID)
	if err != nil {
		t.Fatalf("Deleting clusters by id should not fail. Error: %v", err)
	}

	// DELETE FROM users WHERE id=...
	_, err = u.DeleteByID(nil, userRow.ID)
	if err != nil {
		t.Fatalf("Deleting user by id should not fail. Error: %v", err)
	}
}
