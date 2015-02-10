// Package dao provides Data Access Objects that perform CRUD operations on given storage.
package dao

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/resourced/resourced-master/libstring"
	resourcedmaster_storage "github.com/resourced/resourced-master/storage"
	"golang.org/x/crypto/bcrypt"
	"io"
	"strconv"
	"time"
)

// NewUser returns struct User with basic level permission.
func NewUser(store resourcedmaster_storage.Storer, name, password string) (*User, error) {
	u := &User{}
	u.store = store
	u.Name = name
	u.Level = "basic"
	u.Enabled = true
	u.Type = "human"
	u.Id = strconv.FormatInt(time.Now().UnixNano(), 10)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 5)
	if err != nil {
		return nil, err
	}
	u.HashedPassword = string(hashedPassword)

	accessToken, err := libstring.GeneratePassword(32)
	if err != nil {
		return nil, err
	}
	u.Token = accessToken

	return u, nil
}

// NewAccessTokenUser returns struct User with basic level permission.
func NewAccessTokenUser(store resourcedmaster_storage.Storer, app *Application) (*User, error) {
	accessToken, err := libstring.GeneratePassword(32)
	if err != nil {
		return nil, err
	}

	u, err := NewUser(store, accessToken, accessToken)
	if err != nil {
		return nil, err
	}
	u.Token = accessToken
	u.Type = "token"
	u.ApplicationId = app.Id

	return u, nil
}

// NewUserGivenJson returns struct User.
func NewUserGivenJson(store resourcedmaster_storage.Storer, jsonBody io.ReadCloser) (*User, error) {
	var userArgs map[string]interface{}

	err := json.NewDecoder(jsonBody).Decode(&userArgs)
	if err != nil {
		return nil, err
	}

	if _, ok := userArgs["Name"]; !ok {
		return nil, errors.New("Name key does not exist.")
	}
	if _, ok := userArgs["Password"]; !ok {
		return nil, errors.New("Password key does not exist.")
	}

	u, err := NewUser(store, userArgs["Name"].(string), userArgs["Password"].(string))
	if err != nil {
		return nil, err
	}

	if _, ok := userArgs["Level"]; ok {
		u.Level = userArgs["Level"].(string)
	}
	if _, ok := userArgs["Enabled"]; ok {
		u.Enabled = userArgs["Enabled"].(bool)
	}

	return u, nil
}

// UpdateUserByNameGivenJson returns struct User.
func UpdateUserByNameGivenJson(store resourcedmaster_storage.Storer, name string, allowLevelUpdate bool, jsonBody io.ReadCloser) (*User, error) {
	var userArgs map[string]interface{}

	err := json.NewDecoder(jsonBody).Decode(&userArgs)
	if err != nil {
		return nil, err
	}

	u, err := GetUserByName(store, name)
	if err != nil {
		return nil, err
	}

	if _, ok := userArgs["Name"]; ok {
		u.Name = userArgs["Name"].(string)
	}
	if _, ok := userArgs["Level"]; ok && allowLevelUpdate {
		u.Level = userArgs["Level"].(string)
	}
	if _, ok := userArgs["Enabled"]; ok {
		u.Enabled = userArgs["Enabled"].(bool)
	}

	if _, ok := userArgs["Password"]; ok {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userArgs["Password"].(string)), 5)
		if err != nil {
			return nil, err
		}
		u.HashedPassword = string(hashedPassword)
	}

	err = u.Save()
	if err != nil {
		return nil, err
	}

	return u, nil
}

// AllUsers returns a slice of all User structs.
func AllUsers(store resourcedmaster_storage.Storer) ([]*User, error) {
	nameList, err := store.List("/users/name")
	if err != nil {
		return nil, err
	}

	users := make([]*User, 0)

	for _, name := range nameList {
		u, err := GetUserByName(store, name)
		if err == nil {
			users = append(users, u)
		}
	}

	return users, nil
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

// UpdateUserTokenByName returns struct User.
func UpdateUserTokenByName(store resourcedmaster_storage.Storer, name string) (*User, error) {
	u, err := GetUserByName(store, name)
	if err != nil {
		return nil, err
	}

	accessToken, err := libstring.GeneratePassword(32)
	if err != nil {
		return nil, err
	}
	u.Token = accessToken

	err = u.Save()
	if err != nil {
		return nil, err
	}

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

// saveByName saves user data in JSON format with name as key.
func saveByName(store resourcedmaster_storage.Storer, name string, data []byte) error {
	return store.Update(fmt.Sprintf("/users/name/%v", name), data)
}

type User struct {
	Id             string
	ApplicationId  string
	Name           string
	HashedPassword string
	Level          string
	Type           string
	Token          string
	Enabled        bool
	store          resourcedmaster_storage.Storer
}

// validateBeforeSave checks various conditions before saving.
func (u *User) validateBeforeSave() error {
	if u.Id == "" {
		return errors.New("Id must not be empty.")
	}
	if u.ApplicationId == "" && u.Level != "staff" {
		return errors.New("ApplicationId must not be empty.")
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

	err = saveByName(u.store, u.Name, jsonBytes)
	if err != nil {
		return err
	}

	err = CommonSaveById(u.store, "users", u.Id, jsonBytes)
	if err != nil {
		return err
	}

	if u.ApplicationId != "" && u.Token != "" {
		app, err := GetApplicationById(u.store, u.ApplicationId)
		if err != nil {
			return err
		}

		jsonBytes, err := json.Marshal(app)
		if err != nil {
			return err
		}

		err = SaveApplicationByAccessToken(u.store, u.Token, jsonBytes)
	}

	return err
}

// Delete user
func (u *User) Delete() error {
	err := DeleteUserByName(u.store, u.Name)

	if u.ApplicationId != "" && u.Token != "" {
		err = DeleteApplicationByAccessToken(u.store, u.Token)
	}

	return err
}
