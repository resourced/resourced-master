package dao

import (
	"github.com/resourced/resourced-master/libenv"
	resourcedmaster_storage "github.com/resourced/resourced-master/storage"
	"testing"
)

func s3StorageForTest(t *testing.T) *resourcedmaster_storage.S3 {
	env := "test"
	s3AccessKey := libenv.EnvWithDefault("RESOURCED_MASTER_S3_ACCESS_KEY", "")
	s3SecretKey := libenv.EnvWithDefault("RESOURCED_MASTER_S3_SECRET_KEY", "")
	s3Region := "us-east-1"
	s3Bucket := "resourcedmaster-test"

	if s3AccessKey == "" || s3SecretKey == "" {
		t.Fatal("You must set RESOURCED_MASTER_S3_ACCESS_KEY & RESOURCED_MASTER_S3_SECRET_KEY environments to run these tests.")
	}

	return resourcedmaster_storage.NewS3(env, s3AccessKey, s3SecretKey, s3Region, s3Bucket)
}
