package wstrafficker

import (
	"testing"
)

func TestConstructor(t *testing.T) {
	trafficker := NewWSTrafficker(nil)

	if trafficker.Chans.Send == nil {
		t.Errorf("Send channel should not be nil")
	}
	if trafficker.Chans.Receive == nil {
		t.Errorf("Receive channel should not be nil")
	}
}
