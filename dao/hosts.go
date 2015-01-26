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

// DeleteHostByAppId delete host data in JSON format with app id and hostname as keys.
func DeleteHostByAppId(store resourcedmaster_storage.Storer, id int64, hostname string) error {
	return store.Delete(fmt.Sprintf("applications/id/%v/hosts/names/%v", id, hostname))
}

// GetHostByAppId returns Host struct with app id and hostname as keys.
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

// SaveHostByAppIdAndHardwareAddrJson saves host data in JSON format with app id and hardware address as keys.
func SaveHostByAppIdAndHardwareAddrJson(store resourcedmaster_storage.Storer, id int64, address string, data []byte) error {
	return store.Update(fmt.Sprintf("applications/id/%v/hosts/hardware-addr/%v", id, address), data)
}

// DeleteHostByAppIdAndHardwareAddrJson delete host data in JSON format with app id and hardware address as keys.
func DeleteHostByAppIdAndHardwareAddrJson(store resourcedmaster_storage.Storer, id int64, address string) error {
	return store.Delete(fmt.Sprintf("applications/id/%v/hosts/hardware-addr/%v", id, address))
}

// SaveHostByAppIdAndIpAddrJson saves host data in JSON format with app id and ip address as keys.
func SaveHostByAppIdAndIpAddrJson(store resourcedmaster_storage.Storer, id int64, address string, data []byte) error {
	return store.Update(fmt.Sprintf("applications/id/%v/hosts/ip-addr/%v", id, address), data)
}

// DeleteHostByAppIdAndIpAddrJson delete host data in JSON format with app id and ip address as keys.
func DeleteHostByAppIdAndIpAddrJson(store resourcedmaster_storage.Storer, id int64, address string) error {
	return store.Delete(fmt.Sprintf("applications/id/%v/hosts/ip-addr/%v", id, address))
}

func getHostByAppIdAndAddress(store resourcedmaster_storage.Storer, id int64, addressType, address string) (*Host, error) {
	jsonBytes, err := store.Get(fmt.Sprintf("applications/id/%v/hosts/%v/%v", id, addressType, address))

	h := &Host{}

	err = json.Unmarshal(jsonBytes, h)
	if err != nil {
		return nil, err
	}
	h.store = store

	return h, nil
}

// GetHostByAppIdAndHardwareAddr returns Host struct with app id and hardware address as keys.
func GetHostByAppIdAndHardwareAddr(store resourcedmaster_storage.Storer, id int64, address string) (*Host, error) {
	return getHostByAppIdAndAddress(store, id, "hardware-addr", address)
}

// GetHostByAppIdAndIpAddr returns Host struct with app id and ip address as keys.
func GetHostByAppIdAndIpAddr(store resourcedmaster_storage.Storer, id int64, address string) (*Host, error) {
	return getHostByAppIdAndAddress(store, id, "ip-addr", address)
}

// AllHosts returns a slice of all Host structs.
func AllHosts(store resourcedmaster_storage.Storer, id int64) ([]*Host, error) {
	hostnames, err := store.List(fmt.Sprintf("applications/id/%v/hosts/names", id))
	if err != nil {
		return nil, err
	}

	hosts := make([]*Host, 0)

	for _, hostname := range hostnames {
		host, err := GetHostByAppId(store, id, hostname)
		if err == nil {
			hosts = append(hosts, host)
		}
	}

	return hosts, nil
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

	err = SaveHostByAppIdJson(h.store, h.ApplicationId, h.Name, jsonBytes)
	if err != nil {
		return err
	}

	err = h.SaveByHardwareAddr()
	if err != nil {
		return err
	}

	return h.SaveByIpAddrs()
}

func (h *Host) SaveByHardwareAddr() error {
	if h.NetworkInterfaces != nil {
		jsonBytes, err := json.Marshal(h)
		if err != nil {
			return err
		}

		for _, data := range h.NetworkInterfaces {
			hardwareAddrInterface, ok := data["HardwareAddress"]
			if !ok {
				continue
			}

			hardwareAddr := hardwareAddrInterface.(string)
			SaveHostByAppIdAndHardwareAddrJson(h.store, h.ApplicationId, hardwareAddr, jsonBytes)
		}
	}

	return nil
}

func (h *Host) SaveByIpAddrs() error {
	if h.NetworkInterfaces != nil {
		jsonBytes, err := json.Marshal(h)
		if err != nil {
			return err
		}

		for _, data := range h.NetworkInterfaces {
			ipAddressInterface, ok := data["IPAddresses"]
			if !ok {
				continue
			}

			addrs := ipAddressInterface.([]string)
			for _, addr := range addrs {
				SaveHostByAppIdAndIpAddrJson(h.store, h.ApplicationId, addr, jsonBytes)
			}
		}
	}

	return nil
}

func (h *Host) Delete() error {
	err := DeleteHostByAppId(h.store, h.ApplicationId, h.Name)
	if err != nil {
		return err
	}

	err = h.DeleteByHardwareAddr()
	if err != nil {
		return err
	}

	return h.DeleteByIpAddrs()
}

func (h *Host) DeleteByHardwareAddr() error {
	if h.NetworkInterfaces != nil {
		for _, data := range h.NetworkInterfaces {
			hardwareAddrInterface, ok := data["HardwareAddress"]
			if !ok {
				continue
			}

			hardwareAddr := hardwareAddrInterface.(string)
			DeleteHostByAppIdAndHardwareAddrJson(h.store, h.ApplicationId, hardwareAddr)
		}
	}

	return nil
}

func (h *Host) DeleteByIpAddrs() error {
	if h.NetworkInterfaces != nil {
		for _, data := range h.NetworkInterfaces {
			ipAddressInterface, ok := data["IPAddresses"]
			if !ok {
				continue
			}

			addrs := ipAddressInterface.([]string)
			for _, addr := range addrs {
				DeleteHostByAppIdAndIpAddrJson(h.store, h.ApplicationId, addr)
			}
		}
	}

	return nil
}
