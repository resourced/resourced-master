package dao

import (
	"encoding/json"
	"errors"
	"fmt"
	resourcedmaster_storage "github.com/resourced/resourced-master/storage"
	"strconv"
	"strings"
	"time"
)

// NewApplication is constructor for Application struct.
func NewApplication(store resourcedmaster_storage.Storer, name string) (*Application, error) {
	a := &Application{}
	a.store = store
	a.Name = name
	a.Enabled = true
	a.Id = strconv.FormatInt(time.Now().UnixNano(), 10)

	return a, nil
}

// SaveApplicationByAccessToken saves application data in JSON format with accessToken as key.
func SaveApplicationByAccessToken(store resourcedmaster_storage.Storer, accessToken string, data []byte) error {
	return store.Update("/applications/access-token/"+accessToken, data)
}

// GetApplicationByAccessToken returns Application struct with name as key.
func GetApplicationByAccessToken(store resourcedmaster_storage.Storer, accessToken string) (*Application, error) {
	jsonBytes, err := store.Get("/applications/access-token/" + accessToken)
	if err != nil {
		return nil, err
	}

	a := &Application{}

	err = json.Unmarshal(jsonBytes, a)
	if err != nil {
		return nil, err
	}

	a.store = store

	return a, nil
}

// DeleteApplicationByAccessToken returns error.
func DeleteApplicationByAccessToken(store resourcedmaster_storage.Storer, accessToken string) error {
	return store.Delete(fmt.Sprintf("/applications/access-token/%v", accessToken))
}

// GetApplicationById returns Application struct with name as key.
func GetApplicationById(store resourcedmaster_storage.Storer, id string) (*Application, error) {
	jsonBytes, err := store.Get(fmt.Sprintf("/applications/id/%v/record", id))
	if err != nil {
		return nil, err
	}

	a := &Application{}

	err = json.Unmarshal(jsonBytes, a)
	if err != nil {
		return nil, err
	}

	a.store = store

	return a, nil
}

// AllApplications returns a slice of all Application structs.
func AllApplications(store resourcedmaster_storage.Storer) ([]*Application, error) {
	idList, err := store.List("/applications/id")
	if err != nil {
		return nil, err
	}

	applications := make([]*Application, 0)

	for _, keyWithoutFullpath := range idList {
		keyInChunk := strings.Split(keyWithoutFullpath, "/")
		if len(keyInChunk) < 1 {
			continue
		}

		id := keyInChunk[0]
		a, err := GetApplicationById(store, id)
		if err == nil {
			applications = append(applications, a)
		}
	}

	return applications, nil
}

// SaveApplicationDataByHostJson saves reader data in JSON format with application id + host + path as key.
func SaveApplicationDataByHostJson(store resourcedmaster_storage.Storer, id string, host, path string, data []byte) error {
	return store.Update(fmt.Sprintf("applications/id/%v/hosts/names/%v/data/%v", id, host, path), data)
}

// DeleteApplicationDataByHostJson deletes reader data in JSON format with application id + host + path as key.
func DeleteApplicationDataByHostJson(store resourcedmaster_storage.Storer, id string, host, path string) error {
	return store.Delete(fmt.Sprintf("applications/id/%v/hosts/names/%v/data/%v", id, host, path))
}

// GetApplicationDataByHostJson returns reader data in JSON format with application id + host + path as key.
func GetApplicationDataByHostJson(store resourcedmaster_storage.Storer, id string, host, path string) ([]byte, error) {
	return store.Get(fmt.Sprintf("applications/id/%v/hosts/names/%v/data/%v", id, host, path))
}

// AllApplicationDataByHost returns a slice of all JSON data with application id + host as key.
func AllApplicationDataByHost(store resourcedmaster_storage.Storer, id string, host string) (map[string]interface{}, error) {
	paths, err := store.List(fmt.Sprintf("applications/id/%v/hosts/names/%v/data", id, host))
	if err != nil {
		return nil, err
	}

	allJsonData := make(map[string]interface{})

	for _, path := range paths {
		jsonData, err := GetApplicationDataByHostJson(store, id, host, path)

		if err == nil {
			var data map[string]interface{}

			err = json.Unmarshal(jsonData, &data)
			if err == nil {
				allJsonData["/"+path] = data
			}
		}
	}

	return allJsonData, nil
}

type Application struct {
	Id      string
	Name    string
	Enabled bool
	store   resourcedmaster_storage.Storer
}

// validateBeforeSave checks various conditions before saving.
func (a *Application) validateBeforeSave() error {
	if a.Id == "" {
		return errors.New("Id must not be empty.")
	}
	if a.Name == "" {
		return errors.New("Name must not be empty.")
	}
	return nil
}

// Save application in JSON format.
func (a *Application) Save() error {
	err := a.validateBeforeSave()
	if err != nil {
		return err
	}

	jsonBytes, err := json.Marshal(a)
	if err != nil {
		return err
	}

	err = CommonSaveById(a.store, "applications", a.Id, jsonBytes)
	if err != nil {
		return err
	}

	return nil
}

func (a *Application) Delete() error {
	return CommonDeleteById(a.store, "applications", a.Id)
}

func (a *Application) SaveDataJson(host, path string, data []byte) error {
	return SaveApplicationDataByHostJson(a.store, a.Id, host, path, data)
}

func (a *Application) DeleteDataJson(host, path string) error {
	return DeleteApplicationDataByHostJson(a.store, a.Id, host, path)
}

func (a *Application) GetDataJson(host, path string) ([]byte, error) {
	return GetApplicationDataByHostJson(a.store, a.Id, host, path)
}
