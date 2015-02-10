package dao

import (
	"testing"
)

func appForHostTesting(t *testing.T) *Application {
	store := s3StorageForTest(t)

	app, err := NewApplication(store, "default")
	if err != nil {
		t.Fatalf("Creating application struct should work. Error: %v", err)
	}

	err = app.Save()
	if err != nil {
		t.Fatalf("Saving app struct should work. Error: %v", err)
	}

	return app
}

func hostForHostTesting(t *testing.T, app *Application) *Host {
	store := s3StorageForTest(t)
	hostname := "localhost"
	host := NewHost(store, hostname, app.Id)

	err := host.Save()
	if err != nil {
		t.Fatalf("Saving host data should work. Error: %v", err)
	}
	return host
}

func TestCreateGetDeleteHost(t *testing.T) {
	store := s3StorageForTest(t)
	app := appForAppTesting(t)
	host := hostForHostTesting(t, app)

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

func TestCreateGetDeleteHostData(t *testing.T) {
	app := appForAppTesting(t)
	host := hostForHostTesting(t, app)

	path := "/hello/world"
	data := []byte(`{"Message": "Hello World"}`)

	err := host.SaveDataJson(path, data)
	if err != nil {
		t.Errorf("Saving reader data should work. Error: %v", err)
	}

	inJson, err := host.GetDataJson(path)
	if err != nil {
		t.Errorf("Getting reader data should work. Error: %v", err)
	}

	if string(inJson) != string(data) {
		t.Error("Got the wrong reader data.")
	}

	err = host.DeleteDataJson(path)
	if err != nil {
		t.Errorf("Deleting reader data should work. Error: %v", err)
	}

	app.Delete()
	host.Delete()
}
