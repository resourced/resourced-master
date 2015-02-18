// Package storage provides common interface to multiple storage backend.
// Currently there are: filesystem and S3 backends.
package storage

import (
	"github.com/resourced/resourced-master/libenv"
	"strings"
)

func NewStorage() Storer {
	storageType := libenv.EnvWithDefault("RESOURCED_MASTER_STORAGE_TYPE", "FileSystem")
	env := libenv.EnvWithDefault("RESOURCED_MASTER_ENV", "development")

	if strings.ToLower(storageType) == "filesystem" {
		return NewFileSystem(env)
	}
	if strings.ToLower(storageType) == "s3" {
		accessKey := libenv.EnvWithDefault("RESOURCED_MASTER_ACCESS_KEY", libenv.EnvWithDefault("AWS_ACCESS_KEY_ID", ""))
		secretKey := libenv.EnvWithDefault("RESOURCED_MASTER_SECRET_KEY", libenv.EnvWithDefault("AWS_SECRET_ACCESS_KEY", ""))
		s3Region := libenv.EnvWithDefault("RESOURCED_MASTER_S3_REGION", "us-east-1")
		s3Bucket := libenv.EnvWithDefault("RESOURCED_MASTER_S3_BUCKET", "")

		return NewS3(env, accessKey, secretKey, s3Region, s3Bucket)
	}
	return nil
}

type Storer interface {
	GetRoot() string
	Create(string, []byte) error
	Update(string, []byte) error
	Get(string) ([]byte, error)
	List(string) ([]string, error)
	Delete(string) error
}
