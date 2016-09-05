package config

// EmailConfig stores all email configuration data.
type EmailConfig struct {
	From          string
	SubjectPrefix string
	Host          string
	Port          int
	Username      string
	Password      string
	Identity      string
}
