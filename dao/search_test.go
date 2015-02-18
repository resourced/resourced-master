package dao

import (
	"github.com/resourced/resourced-master/libenv"
	"testing"
)

func searchForTest(t *testing.T) *Search {
	accessKey := libenv.EnvWithDefault("RESOURCED_MASTER_ACCESS_KEY", "")
	secretKey := libenv.EnvWithDefault("RESOURCED_MASTER_SECRET_KEY", "")
	cloudsearchRegion := "us-east-1"

	if accessKey == "" || secretKey == "" {
		t.Fatal("You must set RESOURCED_MASTER_ACCESS_KEY & RESOURCED_MASTER_SECRET_KEY environments to run these tests.")
	}

	return NewSearch(accessKey, secretKey, "", cloudsearchRegion)
}

func TestNewSearch(t *testing.T) {
	s := searchForTest(t)

	if s.credentials == nil {
		t.Errorf("s.credentials should not be empty. s.credentials: %v", s.credentials)
	}
	if s.cloudsearch == nil {
		t.Errorf("s.cloudsearch should not be empty. s.cloudsearch: %v", s.cloudsearch)
	}
}
