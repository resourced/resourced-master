package dao

import (
	"testing"
)

func TestNewApplication(t *testing.T) {
	store := s3StorageForTest(t)

	app, err := NewApplication(store, "default")
	if err != nil {
		t.Errorf("Creating application struct should work. Error: %v", err)
	}

	if app.Id <= 0 {
		t.Errorf("app.Id should not be empty. app.Id: %v", app.Id)
	}
	if app.CreatedUnixNano != app.Id {
		t.Errorf("app.Id == app.CreatedUnixNano. app.Id: %v, app.CreatedUnixNano: %v", app.Id, app.CreatedUnixNano)
	}
	if app.Name == "" {
		t.Errorf("app.Name should not be empty. app.Name: %v", app.Name)
	}
	if app.Enabled == false {
		t.Errorf("app.Enabled should be true by default. app.Enabled: %v", app.Enabled)
	}
}

func TestCreateUpdateDeleteApplication(t *testing.T) {
	store := s3StorageForTest(t)

	app, err := NewApplication(store, "default")
	if err != nil {
		t.Errorf("Creating application struct should work. Error: %v", err)
	}

	err = app.Save()
	if err != nil {
		t.Errorf("Saving app struct should work. Error: %v", err)
	}

	fromStorage, err := GetApplicationById(store, app.Id)
	if err != nil {
		t.Errorf("Failed to get application by id. Error: %v", err)
	}

	if fromStorage.Id != app.Id {
		t.Errorf("Got the wrong application by id. fromStorage: %v, app: %v", fromStorage, app)
	}

	apps, err := AllApplications(store)
	if err != nil {
		t.Errorf("Failed to get all applications. Error: %v", err)
	}

	foundId := false
	for _, a := range apps {
		if a.Id == fromStorage.Id {
			foundId = true
			break
		}
	}
	if !foundId {
		t.Errorf("AllApplications did not return everything.")
	}

	err = app.Delete()
	if err != nil {
		t.Errorf("Deleting app should work. Error: %v", err)
	}

}

func TestCreateGetDeleteReaderData(t *testing.T) {
	store := s3StorageForTest(t)

	app, err := NewApplication(store, "default")
	if err != nil {
		t.Errorf("Creating application struct should work. Error: %v", err)
	}

	err = app.Save()
	if err != nil {
		t.Errorf("Saving app struct should work. Error: %v", err)
	}

	data := []byte(`{"Message": "Hello World"}`)

	err = app.SaveReaderWriter("reader", "hello/world", data)
	if err != nil {
		t.Errorf("Saving reader data should work. Error: %v", err)
	}

	inJson, err := app.GetReaderWriterJson("reader", "hello/world")
	if err != nil {
		t.Errorf("Getting reader data should work. Error: %v", err)
	}

	if string(inJson) != string(data) {
		t.Error("Got the wrong reader data.")
	}

	err = app.DeleteReaderWriter("reader", "hello/world")
	if err != nil {
		t.Errorf("Deleting reader data should work. Error: %v", err)
	}

	app.Delete()
}
