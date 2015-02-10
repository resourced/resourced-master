// Package dao provides Data Access Objects that perform CRUD operations on given storage.
package dao

import (
	"fmt"
	resourcedmaster_storage "github.com/resourced/resourced-master/storage"
)

// CommonSaveById saves dao data in JSON format with Id as key.
func CommonSaveById(store resourcedmaster_storage.Storer, daoType string, id string, data []byte) error {
	return store.Update(fmt.Sprintf("/%s/id/%v/record", daoType, id), data)
}

// CommonDeleteById deletes dao data with Id as key.
func CommonDeleteById(store resourcedmaster_storage.Storer, daoType string, id string) error {
	return store.Delete(fmt.Sprintf("/%s/id/%v/record", daoType, id))
}
