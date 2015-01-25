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
	a.CreatedUnixNano = time.Now().UnixNano()
	a.Id = a.CreatedUnixNano

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
func GetApplicationById(store resourcedmaster_storage.Storer, id int64) (*Application, error) {
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

		id, err := strconv.ParseInt(keyInChunk[0], 10, 64)
		if err != nil {
			continue
		}

		a, err := GetApplicationById(store, id)
		if err == nil {
			applications = append(applications, a)
		}
	}

	return applications, nil
}

// SaveApplicationReaderWriter saves application reader data in JSON format with id and path as keys.
func SaveApplicationReaderWriter(store resourcedmaster_storage.Storer, id int64, readerOrWriter, path string, data []byte) error {
	return store.Update(fmt.Sprintf("applications/id/%v/%vs/%v", id, readerOrWriter, path), data)
}

// DeleteApplicationReaderWriter returns error.
func DeleteApplicationReaderWriter(store resourcedmaster_storage.Storer, id int64, readerOrWriter, path string) error {
	return store.Delete(fmt.Sprintf("applications/id/%v/%vs/%v", id, readerOrWriter, path))
}

// GetApplicationReaderWriterJson returns json bytes.
func GetApplicationReaderWriterJson(store resourcedmaster_storage.Storer, id int64, readerOrWriter, path string) ([]byte, error) {
	return store.Get(fmt.Sprintf("applications/id/%v/%vs/%v", id, readerOrWriter, path))
}

// SaveApplicationHost saves application host data in JSON format with id and hostname as keys.
func SaveApplicationHost(store resourcedmaster_storage.Storer, id int64, hostname string, data []byte) error {
	return store.Update(fmt.Sprintf("applications/id/%v/hosts/names/%v", id, hostname), data)
}

// DeleteApplicationHost delete application host data in JSON format with id and hostname as keys.
func DeleteApplicationHost(store resourcedmaster_storage.Storer, id int64, hostname string) error {
	return store.Delete(fmt.Sprintf("applications/id/%v/hosts/names/%v", id, hostname))
}

// GetApplicationHost get application host data in JSON format with id and hostname as keys.
func GetApplicationHost(store resourcedmaster_storage.Storer, id int64, hostname string) ([]byte, error) {
	return store.Get(fmt.Sprintf("applications/id/%v/hosts/names/%v", id, hostname))
}

type Application struct {
	Id              int64
	Name            string
	Enabled         bool
	CreatedUnixNano int64
	store           resourcedmaster_storage.Storer
}

// validateBeforeSave checks various conditions before saving.
func (a *Application) validateBeforeSave() error {
	if a.Id <= 0 {
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

func (a *Application) SaveReaderWriter(readerOrWriter, path string, data []byte) error {
	return SaveApplicationReaderWriter(a.store, a.Id, readerOrWriter, path, data)
}

func (a *Application) DeleteReaderWriter(readerOrWriter, path string) error {
	return DeleteApplicationReaderWriter(a.store, a.Id, readerOrWriter, path)
}

func (a *Application) GetReaderWriterJson(readerOrWriter, path string) ([]byte, error) {
	return GetApplicationReaderWriterJson(a.store, a.Id, readerOrWriter, path)
}

func (a *Application) SaveHost(hostname string, data []byte) error {
	return SaveApplicationHost(a.store, a.Id, hostname, data)
}

func (a *Application) DeleteHost(hostname string) error {
	return DeleteApplicationHost(a.store, a.Id, hostname)
}

func (a *Application) GetHost(hostname string) ([]byte, error) {
	return GetApplicationHost(a.store, a.Id, hostname)
}
