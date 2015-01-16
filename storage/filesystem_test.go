package storage

import (
	"fmt"
	"github.com/resourced/resourced-master/libenv"
	"os"
	"os/user"
	"path/filepath"
	"testing"
)

func TestDefaultStorageType(t *testing.T) {
	storageType := libenv.EnvWithDefault("RESOURCED_MASTER_STORAGE_TYPE", "FileSystem")
	if storageType != "FileSystem" {
		t.Error("Default storageType should equal to FileSystem")
	}
}

func TestRootFileSystemWithDefaultEnvironment(t *testing.T) {
	currentUser, _ := user.Current()
	env := "development"

	storage := NewStorage()

	if storage.GetRoot() != filepath.Join(currentUser.HomeDir, fmt.Sprintf("resourcedmaster-%v", env)) {
		t.Errorf("Root of FileSystem storage should be located at $HOME/resourcedmaster-%v", env)
	}
}

func TestRootFileSystemWithTestEnvironment(t *testing.T) {
	currentUser, _ := user.Current()
	env := "test"

	os.Setenv("RESOURCED_MASTER_ENV", env)

	storage := NewStorage()

	if storage.GetRoot() != filepath.Join(currentUser.HomeDir, fmt.Sprintf("resourcedmaster-%v", env)) {
		t.Errorf("Root of FileSystem storage should be located at $HOME/resourcedmaster-%v", env)
	}
}

func TestFileSystemCreateGetDelete(t *testing.T) {
	os.Setenv("RESOURCED_MASTER_ENV", "test")

	storage := NewStorage()

	err := storage.Create("/dostuff", []byte("dostuff"))
	if err != nil {
		t.Errorf("Create should not fail. Error: %v", err)
	}

	data, err := storage.Get("/dostuff")
	if err != nil {
		t.Errorf("Get should not fail. Error: %v", err)
	}
	if string(data) != "dostuff" {
		t.Errorf("Get should not fail.")
	}

	err = storage.Delete("/dostuff")
	if err != nil {
		t.Errorf("Delete should not fail. Error: %v", err)
	}
}
