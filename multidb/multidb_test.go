package multidb

import (
	"fmt"
	"testing"

	_ "github.com/lib/pq"
	"github.com/resourced/resourced-master/libunix"
)

func newMultidbForTest(t *testing.T, numDNS, replicationPercentage int) (*MultiDB, error) {
	u, err := libunix.CurrentUser()
	if err != nil {
		t.Fatalf("Getting current user should never fail. Error: %v", err)
	}

	dsns := make([]string, numDNS)

	for i := 0; i < numDNS; i++ {
		dsns[i] = fmt.Sprintf("postgres://%v@localhost:5432/resourced-master-test?sslmode=disable", u)
	}

	return New(dsns, replicationPercentage)
}

func TestPickRandom(t *testing.T) {
	m, err := newMultidbForTest(t, 2, 100)
	if err != nil {
		t.Fatalf("Creating multidb should not fail. Error: %v", err)
	}

	db := m.PickRandom()
	if db == nil {
		t.Fatalf("Picking 1 db should not return nil")
	}
}

func TestPickNext(t *testing.T) {
	m, err := newMultidbForTest(t, 2, 100)
	if err != nil {
		t.Fatalf("Creating multidb should not fail. Error: %v", err)
	}

	indexBeforeNext := m.currentIndex

	db := m.PickNext()
	if db == nil {
		t.Fatalf("Picking the next db should not return nil")
	}

	if db != m.DBs[indexBeforeNext+1] {
		t.Fatalf("Got the wrong db")
	}
}

func TestNumOfConnectionsByReplicationPercentage(t *testing.T) {
	m, err := newMultidbForTest(t, 2, 100)
	expectedConnections := 2

	if err != nil {
		t.Fatalf("Creating multidb should not fail. Error: %v", err)
	}

	numConnections := m.NumOfConnectionsByReplicationPercentage()
	if numConnections != 2 {
		t.Fatalf("Failed to count all connections when replication percentage is 100. Num connections: %v", numConnections)
	}

	m, _ = newMultidbForTest(t, 2, 50)
	expectedConnections = 1

	numConnections = m.NumOfConnectionsByReplicationPercentage()
	if numConnections != expectedConnections {
		t.Fatalf("Failed to count half of the connections when replication percentage is 50. Num connections: %v. Expected: %v", numConnections, expectedConnections)
	}

	m, _ = newMultidbForTest(t, 3, 50)
	expectedConnections = 2 // 1.5 will be rounded up to 2

	numConnections = m.NumOfConnectionsByReplicationPercentage()
	if numConnections != expectedConnections {
		t.Fatalf("Failed to count half of the connections when replication percentage is 50. Num connections: %v. Expected: %v", numConnections, expectedConnections)
	}

	m, _ = newMultidbForTest(t, 10, 25)
	expectedConnections = 3 // 2.5 will be rounded up to 3

	numConnections = m.NumOfConnectionsByReplicationPercentage()
	if numConnections != expectedConnections {
		t.Fatalf("Failed to count half of the connections when replication percentage is 50. Num connections: %v. Expected: %v", numConnections, expectedConnections)
	}

	m, _ = newMultidbForTest(t, 10, 72)
	expectedConnections = 8 // 7.2 will be rounded up to 8

	numConnections = m.NumOfConnectionsByReplicationPercentage()
	if numConnections != expectedConnections {
		t.Fatalf("Failed to count half of the connections when replication percentage is 50. Num connections: %v. Expected: %v", numConnections, expectedConnections)
	}
}

func TestPickMultipleForWrites(t *testing.T) {
	m, _ := newMultidbForTest(t, 3, 50)
	expectedConnections := 2 // 1.5 will be rounded up to 2

	dbs := m.PickMultipleForWrites()
	if len(dbs) != expectedConnections {
		t.Fatalf("Failed to count half of the connections when replication percentage is 50. Num connections: %v. Expected: %v", len(dbs), expectedConnections)
	}
	if len(dbs) == len(m.DBs) {
		t.Fatalf("Returned too many databases. Count: %v, Total: %v", len(dbs), len(m.DBs))
	}

	m, _ = newMultidbForTest(t, 10, 72)
	expectedConnections = 8 // 7.2 will be rounded up to 8

	dbs = m.PickMultipleForWrites()
	if len(dbs) != expectedConnections {
		t.Fatalf("Failed to count half of the connections when replication percentage is 50. Num connections: %v. Expected: %v", len(dbs), expectedConnections)
	}
	if len(dbs) == len(m.DBs) {
		t.Fatalf("Returned too many databases. Count: %v, Total: %v", len(dbs), len(m.DBs))
	}
}
