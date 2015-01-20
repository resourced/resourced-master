package main

import (
	"fmt"
	"github.com/carbocation/interpose"
	resourcedmaster_dao "github.com/resourced/resourced-master/dao"
	"github.com/resourced/resourced-master/libenv"
	resourcedmaster_middlewares "github.com/resourced/resourced-master/middlewares"
	resourcedmaster_storage "github.com/resourced/resourced-master/storage"
	"net/http"
)

func storage() (resourcedmaster_storage.Storer, error) {
	env := libenv.EnvWithDefault("RESOURCED_MASTER_ENV", "test")
	s3AccessKey := libenv.EnvWithDefault("RESOURCED_MASTER_S3_ACCESS_KEY", "")
	s3SecretKey := libenv.EnvWithDefault("RESOURCED_MASTER_S3_SECRET_KEY", "")
	s3Region := libenv.EnvWithDefault("RESOURCED_MASTER_S3_REGION", "us-east-1")
	s3Bucket := libenv.EnvWithDefault("RESOURCED_MASTER_S3_BUCKET", "")

	return resourcedmaster_storage.NewS3(env, s3AccessKey, s3SecretKey, s3Region, s3Bucket), nil
}

func middlewareStruct(store resourcedmaster_storage.Storer) (*interpose.Middleware, error) {
	accessTokens, err := resourcedmaster_dao.AllAccessTokens(store)
	if err != nil {
		return nil, err
	}

	middle := interpose.New()
	middle.Use(resourcedmaster_middlewares.AccessTokenAuth(accessTokens))

	return middle, nil
}

func main() {
	store, _ := storage()

	middle, _ := middlewareStruct(store)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Welcome to the home page!")
	})

	middle.UseHandler(mux)

	serverAddress := libenv.EnvWithDefault("RESOURCED_MASTER_ADDR", ":55655")
	http.ListenAndServe(serverAddress, middle)
}
