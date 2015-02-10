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

// SaveHostDataByHostJson saves reader data in JSON format with application id + host + path as key.
func SaveHostDataByHostJson(store resourcedmaster_storage.Storer, id string, host, path string, data []byte) error {
	return store.Update(fmt.Sprintf("applications/id/%v/hosts/names/%v/data/%v", id, host, path), data)
}

// DeleteHostDataByHostJson deletes reader data in JSON format with application id + host + path as key.
func DeleteHostDataByHostJson(store resourcedmaster_storage.Storer, id string, host, path string) error {
	return store.Delete(fmt.Sprintf("applications/id/%v/hosts/names/%v/data/%v", id, host, path))
}

// GetHostDataByHostJson returns reader data in JSON format with application id + host + path as key.
func GetHostDataByHostJson(store resourcedmaster_storage.Storer, id string, host, path string) ([]byte, error) {
	return store.Get(fmt.Sprintf("applications/id/%v/hosts/names/%v/data/%v", id, host, path))
}

// AllHostDataByHost returns a slice of all JSON data with application id + host as key.
func AllHostDataByHost(store resourcedmaster_storage.Storer, id string, host string) (map[string]interface{}, error) {
	paths, err := store.List(fmt.Sprintf("applications/id/%v/hosts/names/%v/data", id, host))
	if err != nil {
		return nil, err
	}

	allJsonData := make(map[string]interface{})

	for _, path := range paths {
		jsonData, err := GetHostDataByHostJson(store, id, host, path)

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

	allData, err := AllHostDataByHost(store, id, hostname)
	if err != nil {
		return h, err
	}
	h.Data = allData

	return h, nil
}

func allHostnames(store resourcedmaster_storage.Storer, id string) ([]string, error) {
	hostnamesAndData, err := store.List(fmt.Sprintf("applications/id/%v/hosts/names", id))
	if err != nil {
		return nil, err
	}

	hosts := make([]string, 0)

	for _, hostnameAndData := range hostnamesAndData {
		pathInChunk := strings.Split(hostnameAndData, "/")
		if len(pathInChunk) < 2 {
			continue
		}

		hostname := pathInChunk[0]
		if pathInChunk[1] == "record" {
			hosts = append(hosts, hostname)
		}
	}

	return hosts, nil
}

// AllHosts returns a slice of all Host structs.
func AllHosts(store resourcedmaster_storage.Storer, id string) ([]*Host, error) {
	hosts := make([]*Host, 0)

	hostnames, err := allHostnames(store, id)
	if err != nil {
		return hosts, err
	}

	for _, hostname := range hostnames {
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

	// We fill in Data as we perform GET requests.
	Data map[string]interface{}

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

func (h *Host) SaveDataJson(path string, data []byte) error {
	return SaveHostDataByHostJson(h.store, h.ApplicationId, h.Name, path, data)
}

func (h *Host) DeleteDataJson(path string) error {
	return DeleteHostDataByHostJson(h.store, h.ApplicationId, h.Name, path)
}

func (h *Host) GetDataJson(path string) ([]byte, error) {
	return GetHostDataByHostJson(h.store, h.ApplicationId, h.Name, path)
}
