package dao

import (
	"fmt"
	resourcedmaster_storage "github.com/resourced/resourced-master/storage"
)

// CommonSaveByName saves dao data in JSON format with name as key.
func CommonSaveByName(store resourcedmaster_storage.Storer, daoType, name string, data []byte) error {
	return store.Update(fmt.Sprintf("/%s/name/%s", daoType, name), data)
}

// CommonSaveById saves dao data in JSON format with Id as key.
func CommonSaveById(store resourcedmaster_storage.Storer, daoType string, id int64, data []byte) error {
	return store.Update(fmt.Sprintf("/%s/id/%v", daoType, id), data)
}

// CommonDeleteById deletes dao data with Id as key.
func CommonDeleteById(store resourcedmaster_storage.Storer, daoType string, id int64) error {
	return store.Delete(fmt.Sprintf("/%s/id/%v", daoType, id))
}
