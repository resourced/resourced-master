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
