package dao

import (
	"encoding/json"
	"errors"
	"fmt"
	resourcedmaster_storage "github.com/resourced/resourced-master/storage"
)

func NewHost(store resourcedmaster_storage.Storer, name string, appId int64) *Host {
	h := &Host{}
	h.Name = name
	h.ApplicationId = appId
	h.store = store

	return h
}

// SaveHostByAppIdJson saves host data in JSON format with app id and hostname as keys.
func SaveHostByAppIdJson(store resourcedmaster_storage.Storer, id int64, hostname string, data []byte) error {
	return store.Update(fmt.Sprintf("applications/id/%v/hosts/names/%v", id, hostname), data)
}

// GetHostByAppIdJson get host data in JSON format with app id and hostname as keys.
func GetHostByAppIdJson(store resourcedmaster_storage.Storer, id int64, hostname string) ([]byte, error) {
	return store.Get(fmt.Sprintf("applications/id/%v/hosts/names/%v", id, hostname))
}

// DeleteHostByAppId delete host data in JSON format with app id and hostname as keys.
func DeleteHostByAppId(store resourcedmaster_storage.Storer, id int64, hostname string) error {
	return store.Delete(fmt.Sprintf("applications/id/%v/hosts/names/%v", id, hostname))
}

// GetHostByAppId get host data with app id and hostname as keys.
func GetHostByAppId(store resourcedmaster_storage.Storer, id int64, hostname string) (*Host, error) {
	jsonBytes, err := store.Get(fmt.Sprintf("applications/id/%v/hosts/names/%v", id, hostname))

	h := &Host{}

	err = json.Unmarshal(jsonBytes, h)
	if err != nil {
		return nil, err
	}
	h.store = store

	return h, nil
}

type Host struct {
	ApplicationId     int64
	Name              string
	Tags              []string
	NetworkInterfaces map[string]map[string]interface{}

	store resourcedmaster_storage.Storer
}

// validateBeforeSave checks various conditions before saving.
func (h *Host) validateBeforeSave() error {
	if h.ApplicationId <= 0 {
		return errors.New("ApplicationId must not be empty.")
	}
	if h.Name == "" {
		return errors.New("Name must not be empty.")
	}
	return nil
}

// Save host in JSON format.
func (h *Host) Save() error {
	err := h.validateBeforeSave()
	if err != nil {
		return err
	}

	jsonBytes, err := json.Marshal(h)
	if err != nil {
		return err
	}

	return SaveHostByAppIdJson(h.store, h.ApplicationId, h.Name, jsonBytes)
}

func (h *Host) Delete() error {
	return DeleteHostByAppId(h.store, h.ApplicationId, h.Name)
}
