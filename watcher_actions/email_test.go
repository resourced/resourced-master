package watcher_actions

import (
	"testing"
)

func TestSend(t *testing.T) {
	settings := make(map[string]interface{})
	settings["From"] = "bob@example.com"
	settings["Subject"] = "Testing"
	settings["Host"] = "smtp.gmail.com"
	settings["Port"] = 465
	settings["Username"] = "bob@example.com"
	settings["Password"] = ""

	action := &Email{}
	err := action.SetSettings(settings)
	if err != nil {
		t.Fatalf("Settings must be valid to send email. Error: %v", err)
	}

	err = action.Send("bob@example.com")
	if err != nil {
		t.Fatalf("Failed to send email. Error: %v", err)
	}
}
