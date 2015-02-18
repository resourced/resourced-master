package storage

import (
	"github.com/resourced/resourced-master/libenv"
	"os"
	"testing"
)

func s3StorageForTest(t *testing.T) *S3 {
	env := "test"
	accessKey := libenv.EnvWithDefault("RESOURCED_MASTER_ACCESS_KEY", "")
	secretKey := libenv.EnvWithDefault("RESOURCED_MASTER_SECRET_KEY", "")
	s3Region := "us-east-1"
	s3Bucket := "resourcedmaster-test"

	if accessKey == "" || secretKey == "" {
		t.Fatal("You must set RESOURCED_MASTER_ACCESS_KEY & RESOURCED_MASTER_SECRET_KEY environments to run these tests.")
	}

	return NewS3(env, accessKey, secretKey, s3Region, s3Bucket)
}

func TestS3RootWithDefaultEnvironment(t *testing.T) {
	os.Setenv("RESOURCED_MASTER_ENV", "test")
	os.Setenv("RESOURCED_MASTER_STORAGE_TYPE", "s3")

	storage := NewStorage()

	if storage.GetRoot() != "resourcedmaster-test" {
		t.Errorf("Root of S3 storage should be located at resourcedmaster-test. storage.GetRoot(): %v", storage.GetRoot())
	}
}

func TestS3CreateGetDelete(t *testing.T) {
	storage := s3StorageForTest(t)

	err := storage.Create("/hello", []byte(`{"Data": "Hello World"}`))
	if err != nil {
		t.Errorf("Create should not fail. Error: %v, Bucket url: %v", err, storage.Bucket.URL("/hello"))
	}

	data, err := storage.Get("/hello")
	if err != nil {
		t.Errorf("Get should not fail. Error: %v", err)
	}
	if string(data) != `{"Data": "Hello World"}` {
		t.Errorf("Get should not fail.")
	}

	err = storage.Delete("/hello")
	if err != nil {
		t.Errorf("Delete should not fail. Error: %v", err)
	}
}
