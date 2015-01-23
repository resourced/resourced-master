package dao

import (
	"fmt"
	"testing"
	"time"
)

func TestValidateBeforeSave(t *testing.T) {
	u := &User{}

	err := u.validateBeforeSave()
	if err == nil {
		t.Error("validateBeforeSave should return error because Id is empty.")
	}

	u.Id = 1

	err = u.validateBeforeSave()
	if err == nil {
		t.Error("validateBeforeSave should return error because ApplicationId us empty.")
	}

	u.ApplicationId = 1

	err = u.validateBeforeSave()
	if err == nil {
		t.Error("validateBeforeSave should return error because Name is empty.")
	}

	u.Name = "Bob"

	err = u.validateBeforeSave()
	if err != nil {
		t.Errorf("validateBeforeSave should not return error because all conditions are met. Error: %v", err)
	}
}

func TestNewUser(t *testing.T) {
	store := s3StorageForTest(t)

	user, err := NewUser(store, "bob", "abc123")
	if err != nil {
		t.Errorf("Creating user struct should work. Error: %v", err)
	}

	if user.Id <= 0 {
		t.Errorf("user.Id should not be empty. user.Id: %v", user.Id)
	}
	if user.CreatedUnixNano != user.Id {
		t.Errorf("user.Id == user.CreatedUnixNano. user.Id: %v, user.CreatedUnixNano: %v", user.Id, user.CreatedUnixNano)
	}
	if user.HashedPassword == "" {
		t.Errorf("user.HashedPassword should not be empty. user.HashedPassword: %v", user.HashedPassword)
	}
}

func TestSaveLoginDeleteBasicUser(t *testing.T) {
	store := s3StorageForTest(t)

	app, _ := NewApplication(store, fmt.Sprintf("application-for-test-%v", time.Now().UnixNano()))
	app.Save()

	user, err := NewUser(store, "bob", "abc123")
	if err != nil {
		t.Errorf("Creating user struct should work. Error: %v", err)
	}

	user.ApplicationId = app.Id

	err = user.Save()
	if err != nil {
		t.Errorf("Saving user struct should work. Error: %v", err)
	}

	userFromStorage, err := Login(store, "bob", "abc123")
	if err != nil {
		t.Errorf("Login with the correct password should work. Error: %v", err)
	}

	if (user.Id != userFromStorage.Id) || (user.Name != userFromStorage.Name) {
		t.Errorf("Login returns the wrong user. userFromStorage.Id: %v, userFromStorage.Name: %v", userFromStorage.Id, userFromStorage.Name)
	}

	err = user.Delete()
	if err != nil {
		t.Errorf("Deleting user should work. Error: %v", err)
	}

	_, err = GetUserByName(store, "bob")
	if err == nil {
		t.Errorf("Getting user by name should not work because user is deleted.")
	}

	app.Delete()
}

func TestSaveDeleteAccessTokenUser(t *testing.T) {
	store := s3StorageForTest(t)

	app, _ := NewApplication(store, fmt.Sprintf("application-for-test-%v", time.Now().UnixNano()))
	err := app.Save()
	if err != nil {
		t.Errorf("Saving app struct should work. Error: %v", err)
	}

	user, err := NewAccessTokenUser(store, app)
	if err != nil {
		t.Errorf("Creating user struct should work. Error: %v", err)
	}

	err = user.Save()
	if err != nil {
		t.Errorf("Saving user struct should work. Error: %v", err)
	}

	userFromStorage, err := GetUserByName(store, user.Name)
	if err != nil {
		t.Errorf("Getting user by name should work. Error: %v", err)
	}

	if (user.Id != userFromStorage.Id) || (user.Name != userFromStorage.Name) {
		t.Errorf("Login returns the wrong user. userFromStorage.Id: %v, userFromStorage.Name: %v", userFromStorage.Id, userFromStorage.Name)
	}

	users, err := AllUsers(store)
	if err != nil {
		t.Errorf("Failed to get all users. Error: %v", err)
	}

	foundId := false
	for _, u := range users {
		if u.Id == userFromStorage.Id {
			foundId = true
			break
		}
	}
	if !foundId {
		t.Error("AllUsers did not return everything.")
	}

	err = user.Delete()
	if err != nil {
		t.Errorf("Deleting user should work. Error: %v", err)
	}

	_, err = GetUserByName(store, user.Name)
	if err == nil {
		t.Errorf("Getting user by name should not work because user is deleted.")
	}

	app.Delete()
}

func TestNewUserGivenJson(t *testing.T) {

}

func TestUpdateUserByNameGivenJson(t *testing.T) {

}

func TestUpdateUserTokenByName(t *testing.T) {

}
