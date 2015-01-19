package dao

import (
	"encoding/json"
	"testing"
)

func TestNewAccessToken(t *testing.T) {
	store := s3StorageForTest(t)

	app, err := NewApplication(store, "default")
	if err != nil {
		t.Errorf("Creating application struct should work. Error: %v", err)
	}

	access, err := NewAccessToken(store, app)
	if err != nil {
		t.Errorf("Creating access token struct should work. Error: %v", err)
	}

	if access.Id <= 0 {
		t.Errorf("access.Id should not be empty. access.Id: %v", access.Id)
	}
	if access.ApplicationId <= 0 {
		t.Errorf("access.ApplicationId should not be empty. access.ApplicationId: %v", access.ApplicationId)
	}
	if access.CreatedUnixNano != access.Id {
		t.Errorf("access.Id == access.CreatedUnixNano. access.Id: %v, access.CreatedUnixNano: %v", access.Id, access.CreatedUnixNano)
	}
	if access.Token == "" {
		t.Errorf("access.Token should not be empty. access.Token: %v", access.Token)
	}
	if access.Enabled == false {
		t.Errorf("access.Enabled should be true by default. access.Enabled: %v", access.Enabled)
	}
	if access.Level == "" {
		t.Errorf("access.Level should not be empty. access.Level: %v", access.Level)
	}
}

func TestCreateUpdateDeleteAccessToken(t *testing.T) {
	store := s3StorageForTest(t)

	app, _ := NewApplication(store, "default")
	app.Save()

	access, err := NewAccessToken(store, app)
	if err != nil {
		t.Errorf("Creating access token struct should work. Error: %v", err)
	}

	err = access.Save()
	if err != nil {
		t.Errorf("Saving access token struct should work. Error: %v", err)
	}

	jsonBytes, _ := json.Marshal(app)

	err = SaveApplicationByAccessToken(app.store, access.Token, jsonBytes)
	if err != nil {
		t.Errorf("Saving application by access token struct should work. Error: %v", err)
	}

	appFromStorage, err := GetApplicationByAccessToken(app.store, access.Token)
	if err != nil {
		t.Errorf("Getting application by access token should work. Error: %v", err)
	}

	if appFromStorage.Id != app.Id {
		t.Errorf("Got the wrong application by access token. appFromStorage.Id: %v, app.Id: %v", appFromStorage.Id, app.Id)
	}

	err = access.Delete()
	if err != nil {
		t.Errorf("Deleting access token should work. Error: %v", err)
	}

	err = app.Delete()
	if err != nil {
		t.Errorf("Deleting application should work. Error: %v", err)
	}
}
