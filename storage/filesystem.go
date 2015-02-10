package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

// NewFileSystem is constructor for FileSystem
func NewFileSystem(env string) *FileSystem {
	currentUser, _ := user.Current()

	store := &FileSystem{}
	store.Env = env
	store.Root = filepath.Join(currentUser.HomeDir, fmt.Sprintf("resourcedmaster-%v", env))
	return store
}

type FileSystem struct {
	Env  string
	Root string
}

// GetRoot returns root path.
func (fs *FileSystem) GetRoot() string {
	return fs.Root
}

func (fs *FileSystem) CreateOrUpdate(fullpath string, data []byte) error {
	var err error

	mutex := &sync.Mutex{}
	fullpath = path.Join(fs.Root, fullpath)
	basepath := path.Dir(fullpath)

	mutex.Lock()

	if _, err = os.Stat(fullpath); os.IsNotExist(err) {
		// Create parent directory
		err = os.MkdirAll(basepath, 0744)
		if err != nil {
			mutex.Unlock()
			return err
		}

		// Create file
		fileHandler, err := os.Create(fullpath)
		if err != nil {
			mutex.Unlock()
			return err
		}
		defer fileHandler.Close()
	}

	err = ioutil.WriteFile(fullpath, data, 0744)

	mutex.Unlock()

	return err
}

// Create saves JSON data with fullpath as key.
func (fs *FileSystem) Create(fullpath string, data []byte) error {
	return fs.CreateOrUpdate(fullpath, data)
}

// Update saves JSON data with fullpath as key.
func (fs *FileSystem) Update(fullpath string, data []byte) error {
	return fs.CreateOrUpdate(fullpath, data)
}

// Get returns JSON data with fullpath as key.
func (fs *FileSystem) Get(fullpath string) ([]byte, error) {
	if !strings.HasPrefix(fullpath, fs.Root) {
		fullpath = path.Join(fs.Root, fullpath)
	}
	return ioutil.ReadFile(fullpath)
}

// List returns a slice of base paths.
func (fs *FileSystem) List(fullpath string) ([]string, error) {
	if !strings.HasPrefix(fullpath, fs.Root) {
		fullpath = path.Join(fs.Root, fullpath)
	}
	files, err := ioutil.ReadDir(fullpath)
	names := make([]string, len(files))

	for index, f := range files {
		names[index] = f.Name()
	}

	return names, err
}

// Delete removes item with fullpath as key.
func (fs *FileSystem) Delete(fullpath string) error {
	if !strings.HasPrefix(fullpath, fs.Root) {
		fullpath = path.Join(fs.Root, fullpath)
	}
	return os.RemoveAll(fullpath)
}
