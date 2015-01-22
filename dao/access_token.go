package dao

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/resourced/resourced-master/libstring"
	resourcedmaster_storage "github.com/resourced/resourced-master/storage"
	"strconv"
	"time"
)

// AllAccessTokens returns slice of all AccessToken
func AllAccessTokens(store resourcedmaster_storage.Storer) ([]*AccessToken, error) {
	ids, err := store.List("/access-tokens/id")
	if err != nil {
		return nil, err
	}

	accessTokens := make([]*AccessToken, 0)

	for _, idString := range ids {
		id, err := strconv.ParseInt(idString, 10, 0)
		if err != nil {
			continue
		}

		access, err := GetAccessTokenById(store, id)
		if err != nil {
			continue
		}

		accessTokens = append(accessTokens, access)
	}

	return accessTokens, nil
}

// GetApplicationById returns AccessToken struct with name as key.
func GetAccessTokenById(store resourcedmaster_storage.Storer, id int64) (*AccessToken, error) {
	jsonBytes, err := store.Get(fmt.Sprintf("/access-tokens/id/%v", id))
	if err != nil {
		return nil, err
	}

	a := &AccessToken{}

	err = json.Unmarshal(jsonBytes, a)
	if err != nil {
		return nil, err
	}

	a.store = store

	return a, nil
}

// DeleteAccessTokenById returns error.
func DeleteAccessTokenById(store resourcedmaster_storage.Storer, id int64) error {
	jsonBytes, err := store.Get(fmt.Sprintf("/access-tokens/id/%v", id))
	if err != nil {
		return err
	}

	a := &AccessToken{}

	err = json.Unmarshal(jsonBytes, a)
	if err != nil {
		return err
	}

	a.store = store

	return a.Delete()
}

// DeleteAccessTokenById returns error.
func DeleteAccessTokenByToken(store resourcedmaster_storage.Storer, token string) error {
	accessTokens, err := AllAccessTokens(store)
	if err != nil {
		return err
	}

	for _, accessToken := range accessTokens {
		if accessToken.Token == token {
			return accessToken.Delete()
		}
	}

	return errors.New("Unable to find token.")
}

// NewAccessToken is constructor for NewAccessToken struct.
func NewAccessToken(store resourcedmaster_storage.Storer, app *Application) (*AccessToken, error) {
	a := &AccessToken{}
	a.store = store
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
