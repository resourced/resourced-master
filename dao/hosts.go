package dao

import (
	"encoding/json"
	"errors"
	"fmt"
	resourcedmaster_storage "github.com/resourced/resourced-master/storage"
	"strings"
)

func NewHost(store resourcedmaster_storage.Storer, name string, appId string) *Host {
	h := &Host{}
	h.Name = name
	h.ApplicationId = appId
	h.store = store

	return h
}

// SaveHostByAppIdJson saves host data in JSON format with app id and hostname as keys.
func SaveHostByAppIdJson(store resourcedmaster_storage.Storer, id string, hostname string, data []byte) error {
	return store.Update(fmt.Sprintf("applications/id/%v/hosts/names/%v/record", id, hostname), data)
}

// DeleteHostByAppId delete host data in JSON format with app id and hostname as keys.
func DeleteHostByAppId(store resourcedmaster_storage.Storer, id string, hostname string) error {
	return store.Delete(fmt.Sprintf("applications/id/%v/hosts/names/%v/record", id, hostname))
}

// GetHostByAppId returns Host struct with app id and hostname as keys.
func GetHostByAppId(store resourcedmaster_storage.Storer, id string, hostname string) (*Host, error) {
	jsonBytes, err := store.Get(fmt.Sprintf("applications/id/%v/hosts/names/%v/record", id, hostname))

	h := &Host{}

	err = json.Unmarshal(jsonBytes, h)
	if err != nil {
		return nil, err
	}
	h.store = store

	return h, nil
}

// AllHosts returns a slice of all Host structs.
func AllHosts(store resourcedmaster_storage.Storer, id string) ([]*Host, error) {
	hostnamesAndData, err := store.List(fmt.Sprintf("applications/id/%v/hosts/names", id))
	if err != nil {
		return nil, err
	}

	hosts := make([]*Host, 0)

	for _, hostnameAndData := range hostnamesAndData {
		pathInChunk := strings.Split(hostnameAndData, "/")
		if len(pathInChunk) < 1 {
			continue
		}

		hostname := pathInChunk[0]

		host, err := GetHostByAppId(store, id, hostname)
		if err == nil {
			hosts = append(hosts, host)
		}
	}

	return hosts, nil
}

type Host struct {
	ApplicationId     string
	Name              string
	Tags              []string
	NetworkInterfaces map[string]map[string]interface{}

	store resourcedmaster_storage.Storer
}

// validateBeforeSave checks various conditions before saving.
func (h *Host) validateBeforeSave() error {
	if h.ApplicationId == "" {
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
