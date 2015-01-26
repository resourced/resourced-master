package dao

import (
	"testing"
)

func TestCreateGetDeleteHostData(t *testing.T) {
	store := s3StorageForTest(t)

	app, err := NewApplication(store, "default")
	if err != nil {
		t.Errorf("Creating application struct should work. Error: %v", err)
	}

	err = app.Save()
	if err != nil {
		t.Errorf("Saving app struct should work. Error: %v", err)
	}

	hostname := "localhost"

	host := NewHost(store, hostname, app.Id)

	err = host.Save()
	if err != nil {
		t.Errorf("Saving host data should work. Error: %v", err)
	}

	hostFromStorage, err := GetHostByAppId(store, app.Id, "localhost")
	if err != nil {
		t.Errorf("Getting host data should work. Error: %v", err)
	}

	if hostFromStorage.Name != host.Name {
		t.Error("Got the wrong host data.")
	}

	hosts, err := AllHosts(store, app.Id)
	if err != nil {
		t.Errorf("Getting all hosts data should work. Error: %v", err)
	}
	if len(hosts) <= 0 {
		t.Error("There should be at least 1 host data.")
	}

	err = hostFromStorage.Delete()
	if err != nil {
		t.Errorf("Deleting host data should work. Error: %v", err)
	}

	app.Delete()
}
