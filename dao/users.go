package dao

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/resourced/resourced-master/libstring"
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

	accessToken, err := libstring.GeneratePassword(32)
	if err != nil {
		return nil, err
	}
	u.AccessToken = accessToken

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

// SaveByAccessToken saves user data in JSON format with accessToken as key.
func SaveByAccessToken(store resourcedmaster_storage.Storer, accessToken string, data []byte) error {
	return store.Update("/users/access-token/"+accessToken, data)
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

// GetByAccessToken returns User struct with name as key.
func GetByAccessToken(store resourcedmaster_storage.Storer, accessToken string) (*User, error) {
	jsonBytes, err := store.Get("/users/access-token/" + accessToken)
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

	err = store.Delete("/users/access-token/" + u.AccessToken)
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

func LoginByAccessToken(store resourcedmaster_storage.Storer, accessToken string) (*User, error) {
	return GetByAccessToken(store, accessToken)
}

type User struct {
	Id              int64
	Name            string
	HashedPassword  string
	Level           string
	Enabled         bool
	CreatedUnixNano int64
	AccessToken     string
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

	err = SaveByAccessToken(u.store, u.AccessToken, jsonBytes)
	if err != nil {
		return err
	}

	return nil
}

// Delete user
func (u *User) Delete() error {
	return DeleteUserByName(u.store, u.Name)
}
