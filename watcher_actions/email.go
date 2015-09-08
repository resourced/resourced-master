package watcher_actions

import (
	"github.com/resourced/resourced-master/dal"
	"gopkg.in/gomail.v2"
)

type Email struct {
	From     string
	Subject  string
	Host     string
	Port     int
	Username string
	Password string
	watcher  *dal.WatcherRow
}

func (e *Email) SetSettings(settings map[string]interface{}) error {
	e.From = settings["From"].(string)
	e.Subject = settings["Subject"].(string)
	e.Host = settings["Host"].(string)
	e.Port = settings["Port"].(int)
	e.Username = settings["Username"].(string)
	e.Password = settings["Password"].(string)

	return nil
}

func (e *Email) SetWatcher(watcher *dal.WatcherRow) {
	e.watcher = watcher
}

func (e *Email) Send(to string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", e.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", e.Subject)
	m.SetBody("text/plain", "Hello World!")

	d := gomail.NewPlainDialer(e.Host, e.Port, e.Username, e.Password)

	return d.DialAndSend(m)
}
