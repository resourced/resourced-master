package users

import (
	"github.com/resourced/resourced-master/libenv"
	resourcedmaster_storage "github.com/resourced/resourced-master/storage"
	"testing"
)

func s3StorageForTest(t *testing.T) *resourcedmaster_storage.S3 {
	env := "test"
	s3AccessKey := libenv.EnvWithDefault("RESOURCED_MASTER_S3_ACCESS_KEY", "")
	s3SecretKey := libenv.EnvWithDefault("RESOURCED_MASTER_S3_SECRET_KEY", "")
	s3Region := "us-east-1"
	s3Bucket := "resourcedmaster-test"

	if s3AccessKey == "" || s3SecretKey == "" {
		t.Fatal("You must set RESOURCED_MASTER_S3_ACCESS_KEY & RESOURCED_MASTER_S3_SECRET_KEY environments to run these tests.")
	}

	return resourcedmaster_storage.NewS3(env, s3AccessKey, s3SecretKey, s3Region, s3Bucket)
}

func TestValidateBeforeSave(t *testing.T) {
	u := &User{}

	err := u.validateBeforeSave()
	if err == nil {
		t.Error("validateBeforeSave should return error because Id is empty.")
	}

	u.Id = 1

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
	if user.AccessToken == "" {
		t.Errorf("user.AccessToken should not be empty. user.AccessToken: %v", user.AccessToken)
	}
}

func TestSaveLoginDelete(t *testing.T) {
	store := s3StorageForTest(t)

	user, err := NewUser(store, "bob", "abc123")
	if err != nil {
		t.Errorf("Creating user struct should work. Error: %v", err)
	}

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

	userFromStorage2, err := LoginByAccessToken(store, user.AccessToken)
	if err != nil {
		t.Errorf("Login with the correct accessToken should work. Error: %v", err)
	}

	if (user.Id != userFromStorage2.Id) || (user.Name != userFromStorage2.Name) {
		t.Errorf("Login returns the wrong user. userFromStorage2.Id: %v, userFromStorage2.Name: %v", userFromStorage2.Id, userFromStorage2.Name)
	}

	err = user.Delete()
	if err != nil {
		t.Errorf("Deleting user should work. Error: %v", err)
	}

	_, err = GetByName(store, "bob")
	if err == nil {
		t.Errorf("Getting user by name should not work because user is deleted.")
	}
}
