package dao

import (
	"github.com/resourced/resourced-master/libenv"
	resourcedmaster_storage "github.com/resourced/resourced-master/storage"
	"testing"
)

func s3StorageForTest(t *testing.T) *resourcedmaster_storage.S3 {
	env := "test"
	accessKey := libenv.EnvWithDefault("RESOURCED_MASTER_ACCESS_KEY", "")
	secretKey := libenv.EnvWithDefault("RESOURCED_MASTER_SECRET_KEY", "")
	s3Region := "us-east-1"
	s3Bucket := "resourcedmaster-test"

	if accessKey == "" || secretKey == "" {
		t.Fatal("You must set RESOURCED_MASTER_ACCESS_KEY & RESOURCED_MASTER_SECRET_KEY environments to run these tests.")
	}

	return resourcedmaster_storage.NewS3(env, accessKey, secretKey, s3Region, s3Bucket)
}
