package users

import (
	"encoding/json"
	"errors"
	"fmt"
	resourcedmaster_storage "github.com/resourced/resourced-master/storage"
	"golang.org/x/crypto/bcrypt"
	"time"
)

// NewUser returns struct User with basic level permission.
func NewUser(store resourcedmaster_storage.Storer, name, password string) (*User, error) {
	u := &User{}
	u.store = store
	u.Name = name
	u.Level = "basic"
	u.Enabled = true
	u.CreatedUnixNano = time.Now().UnixNano()
	u.Id = u.CreatedUnixNano

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 5)
	if err != nil {
		return nil, err
	}

	u.HashedPassword = string(hashedPassword)

	return u, nil
}

// NewUser returns struct User with admin level permission.
func NewAdminUser(store resourcedmaster_storage.Storer, name, password string) (*User, error) {
	u, err := NewUser(store, name, password)
	if err != nil {
		return nil, err
	}
	u.Level = "admin"

	return u, nil
}

// SaveByName saves user data in JSON format with name as key.
func SaveByName(store resourcedmaster_storage.Storer, name string, data []byte) error {
	return store.Update("/users/name/"+name, data)
}

// SaveById saves user data in JSON format with Id as key.
func SaveById(store resourcedmaster_storage.Storer, id int64, data []byte) error {
	return store.Update(fmt.Sprintf("/users/id/%s", id), data)
}

// GetByName returns User struct with name as key.
func GetByName(store resourcedmaster_storage.Storer, name string) (*User, error) {
	jsonBytes, err := store.Get("/users/name/" + name)
	if err != nil {
		return nil, err
	}

	u := &User{}

	err = json.Unmarshal(jsonBytes, u)
	if err != nil {
		return nil, err
	}

	u.store = store

	return u, nil
}

// DeleteByName deletes user with name as key.
func DeleteByName(store resourcedmaster_storage.Storer, name string) error {
	u, err := GetByName(store, name)
	if err != nil {
		return err
	}

	err = store.Delete("/users/name/" + name)
	if err != nil {
		return err
	}

	err = store.Delete(fmt.Sprintf("/users/id/%s", u.Id))
	if err != nil {
		return err
	}

	return nil
}

// Login returns User struct after validating password.
func Login(store resourcedmaster_storage.Storer, name, password string) (*User, error) {
	u, err := GetByName(store, name)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(password))
	if err != nil {
		return nil, err
	}

	return u, nil
}

type User struct {
	Id              int64
	Name            string
	HashedPassword  string
	Level           string
	Enabled         bool
	CreatedUnixNano int64
	store           resourcedmaster_storage.Storer
}

// validateBeforeSave checks various conditions before saving.
func (u *User) validateBeforeSave() error {
	if u.Id <= 0 {
		return errors.New("Id must not be empty.")
	}
	if u.Name == "" {
		return errors.New("Name must not be empty.")
	}
	return nil
}

// Save saves user in JSON format.
func (u *User) Save() error {
	err := u.validateBeforeSave()
	if err != nil {
		return err
	}

	jsonBytes, err := json.Marshal(u)
	if err != nil {
		return err
	}

	err = SaveByName(u.store, u.Name, jsonBytes)
	if err != nil {
		return err
	}

	err = SaveById(u.store, u.Id, jsonBytes)
	if err != nil {
		return err
	}

	return nil
}

func (u *User) Delete() error {
	return DeleteByName(u.store, u.Name)
}
