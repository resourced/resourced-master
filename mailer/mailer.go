// Package mailer provides simple abstraction around sending email.
package mailer

import (
	"fmt"
	"net/smtp"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/libsmtp"
)

func New(conf *config.EmailConfig) (*Mailer, error) {
	mailer := &Mailer{}
	if conf == nil {
		return mailer, nil
	}

	mailer.Auth = smtp.PlainAuth(
		conf.Identity,
		conf.Username,
		conf.Password,
		conf.Host)

	mailer.HostAndPort = fmt.Sprintf("%v:%v", conf.Host, conf.Port)

	mailer.From = conf.From
	mailer.SubjectPrefix = conf.SubjectPrefix

	return mailer, nil
}

type Mailer struct {
	Auth          smtp.Auth
	HostAndPort   string
	From          string
	SubjectPrefix string
}

func (m *Mailer) Send(to, subject, body string) error {
	subject = m.SubjectPrefix + " " + subject
	message := libsmtp.BuildMessage(m.From, to, subject, body)
	return smtp.SendMail(m.HostAndPort, m.Auth, m.From, []string{to}, []byte(message))
}
