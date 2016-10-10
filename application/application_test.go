package application

import (
	"context"
	"os"
	"testing"
)

func TestConstructor(t *testing.T) {
	app, err := New("../tests/config-files")
	if err != nil {
		t.Errorf("Creating a new app should not fail using test config. Error: %v", err)
	}
	hostname, _ := os.Hostname()
	if app.Hostname != hostname {
		t.Errorf("Failed to configure hostname properly")
	}
}
