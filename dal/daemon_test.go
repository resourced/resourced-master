package dal

import (
	"testing"

	_ "github.com/lib/pq"
)

func newDaemonForTest(t *testing.T) *Daemon {
	return NewDaemon(newDbForTest(t))
}

func TestDaemonCRUD(t *testing.T) {
	d := newDaemonForTest(t)
	defer d.db.Close()

	// Create Daemon
	daemonRow, err := d.CreateOrUpdate(nil, "hostname")
	if err != nil {
		t.Fatalf("Creating a Daemon should work. Error: %v", err)
	}
	if daemonRow.ID <= 0 {
		t.Fatalf("Daemon ID should be assign properly. DaemonRow.ID: %v", daemonRow.ID)
	}

	daemonRow2, err := d.GetByHostname(nil, "hostname")
	if err != nil {
		t.Fatalf("Getting daemon by hostname should not fail. Error: %v", err)
	}
	if daemonRow.ID != daemonRow2.ID {
		t.Fatalf("Got the wrong daemon")
	}

	_, err = d.DeleteByID(nil, daemonRow.ID)
	if err != nil {
		t.Fatalf("Deleting Daemons by id should not fail. Error: %v", err)
	}
}
