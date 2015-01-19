package dao

import (
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
