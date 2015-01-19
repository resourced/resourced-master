package dao

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

// GetUserByName returns User struct with name as key.
func GetUserByName(store resourcedmaster_storage.Storer, name string) (*User, error) {
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

// DeleteUserByName deletes user with name as key.
func DeleteUserByName(store resourcedmaster_storage.Storer, name string) error {
	u, err := GetUserByName(store, name)
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
	u, err := GetUserByName(store, name)
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

// Save user in JSON format.
func (u *User) Save() error {
	err := u.validateBeforeSave()
	if err != nil {
		return err
	}

	jsonBytes, err := json.Marshal(u)
	if err != nil {
		return err
	}

	err = CommonSaveByName(u.store, "users", u.Name, jsonBytes)
	if err != nil {
		return err
	}

	err = CommonSaveById(u.store, "users", u.Id, jsonBytes)
	if err != nil {
		return err
	}

	return nil
}

// Delete user
func (u *User) Delete() error {
	return DeleteUserByName(u.store, u.Name)
}
