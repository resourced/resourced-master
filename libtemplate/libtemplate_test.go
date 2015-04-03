package libtemplate

import (
	"github.com/GeertJohan/go.rice"
	"testing"
)

func TestGetFromRicebox(t *testing.T) {
	gorice := &rice.Config{
		LocateOrder: []rice.LocateMethod{rice.LocateEmbedded, rice.LocateAppended, rice.LocateFS},
	}

	box, err := gorice.FindBox("../templates")
	if err != nil {
		t.Errorf("Unable to get ricebox. Error: %v", err)
	}

	_, err = GetFromRicebox(box, false, "dashboard.html.tmpl", "hosts/list.html.tmpl")
	if err != nil {
		t.Errorf("Unable to get template struct from ricebox. Error: %v", err)
	}
}
