package dao

import (
	"encoding/json"
	"errors"
	"github.com/resourced/resourced-master/libstring"
	resourcedmaster_storage "github.com/resourced/resourced-master/storage"
	"time"
)

// NewAccessToken is constructor for NewAccessToken struct.
func NewAccessToken(store resourcedmaster_storage.Storer, app *Application) (*AccessToken, error) {
	a := &AccessToken{}
	a.store = store
	a.Token = ""
	a.Level = "basic"
	a.Enabled = true
	a.CreatedUnixNano = time.Now().UnixNano()
	a.Id = a.CreatedUnixNano
	a.ApplicationId = app.Id

	accessToken, err := libstring.GeneratePassword(32)
	if err != nil {
		return nil, err
	}
	a.Token = accessToken

	return a, nil
}

type AccessToken struct {
	Id              int64
	ApplicationId   int64
	Token           string
	Level           string
	Enabled         bool
	CreatedUnixNano int64
	store           resourcedmaster_storage.Storer
}

// validateBeforeSave checks various conditions before saving.
func (a *AccessToken) validateBeforeSave() error {
	if a.Id <= 0 {
		return errors.New("Id must not be empty.")
	}
	if a.Token == "" {
		return errors.New("Token must not be empty.")
	}
	return nil
}

// Save application in JSON format.
func (a *AccessToken) Save() error {
	err := a.validateBeforeSave()
	if err != nil {
		return err
	}

	jsonBytes, err := json.Marshal(a)
	if err != nil {
		return err
	}

	err = CommonSaveById(a.store, "access-tokens", a.Id, jsonBytes)
	if err != nil {
		return err
	}

	return nil
}

func (a *AccessToken) Delete() error {
	err := CommonDeleteById(a.store, "access-tokens", a.Id)
	if err != nil {
		return err
	}

	err = a.store.Delete("/applications/access-token/" + a.Token)
	if err != nil {
		return err
	}
	return nil
}
