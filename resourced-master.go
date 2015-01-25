package main

import (
	"github.com/carbocation/interpose"
	gorilla_mux "github.com/gorilla/mux"
	resourcedmaster_dao "github.com/resourced/resourced-master/dao"
	resourcedmaster_handlers "github.com/resourced/resourced-master/handlers"
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
	users, err := resourcedmaster_dao.AllUsers(store)
	if err != nil {
		return nil, err
	}

	middle := interpose.New()
	middle.Use(resourcedmaster_middlewares.SetStore(store))
	middle.Use(resourcedmaster_middlewares.AccessTokenAuth(users))

	return middle, nil
}

func mux() *gorilla_mux.Router {
	router := gorilla_mux.NewRouter()

	// Admin level access
	router.HandleFunc("/api/users", resourcedmaster_handlers.PostApiUser).Methods("POST")
	router.HandleFunc("/api/users", resourcedmaster_handlers.GetApiUser).Methods("GET")

	router.HandleFunc("/api/users/{name}", resourcedmaster_handlers.GetApiUserName).Methods("GET")
	router.HandleFunc("/api/users/{name}", resourcedmaster_handlers.PutApiUserName).Methods("PUT")
	router.HandleFunc("/api/users/{name}", resourcedmaster_handlers.DeleteApiUserName).Methods("DELETE")

	router.HandleFunc("/api/users/{name}/access-token", resourcedmaster_handlers.PutApiUserNameAccessToken).Methods("PUT")
	router.HandleFunc("/api/app/{id:[0-9]+}/access-token", resourcedmaster_handlers.PostApiApplicationIdAccessToken).Methods("POST")
	router.HandleFunc("/api/app/{id:[0-9]+}/access-token/:token", resourcedmaster_handlers.DeleteApiApplicationIdAccessToken).Methods("DELETE")

	// Basic level access
	router.HandleFunc("/", resourcedmaster_handlers.GetRoot).Methods("GET")
	router.HandleFunc("/api", resourcedmaster_handlers.GetApi).Methods("GET")
	router.HandleFunc("/api/app", resourcedmaster_handlers.GetApiApp).Methods("GET")

	router.HandleFunc("/api/app/{id:[0-9]+}/{reader-or-writer}/{path}", resourcedmaster_handlers.PostApiAppIdReaderWriter).Methods("POST")

	return router
}

func main() {
	store, _ := storage()

	middle, _ := middlewareStruct(store)

	middle.UseHandler(mux())

	serverAddress := libenv.EnvWithDefault("RESOURCED_MASTER_ADDR", ":55655")
	http.ListenAndServe(serverAddress, middle)
}
