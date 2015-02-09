package main

import (
	"encoding/json"
	"github.com/carbocation/interpose"
	"github.com/codegangsta/cli"
	gorilla_mux "github.com/gorilla/mux"
	resourcedmaster_dao "github.com/resourced/resourced-master/dao"
	resourcedmaster_handlers "github.com/resourced/resourced-master/handlers"
	"github.com/resourced/resourced-master/libenv"
	resourcedmaster_middlewares "github.com/resourced/resourced-master/middlewares"
	resourcedmaster_storage "github.com/resourced/resourced-master/storage"
	"log"
	"net/http"
	"os"
)

func NewResourcedMaster() (*ResourcedMaster, error) {
	var err error

	rm := &ResourcedMaster{}
	rm.Env = libenv.EnvWithDefault("RESOURCED_MASTER_ENV", "development")

	rm.app = cli.NewApp()
	rm.app.Name = "resourced-master"
	rm.app.Usage = "It stores all of your servers facts."
	rm.app.Author = "Didip Kerabat"
	rm.app.Email = "didipk@gmail.com"

	rm.app.Commands = []cli.Command{
		{
			Name:      "server",
			ShortName: "serv",
			Usage:     "Run HTTP server",
			Action: func(c *cli.Context) {
				rm.httpRun()
			},
		},
		{
			Name:      "application",
			ShortName: "app",
			Usage:     "Application CRUD operations",
			Action: func(c *cli.Context) {
				crud := c.Args().First()

				if crud == "create" {
					name := c.Args().Get(1)

					application, err := resourcedmaster_dao.NewApplication(rm.store, name)
					if err != nil {
						log.Fatalf("Failed to create a new application. Error: %v\n", err)
					}

					err = application.Save()
					if err != nil {
						log.Fatalf("Failed to save the new application. Error: %v\n", err)
					}

					jsonBytes, err := json.Marshal(application)
					if err != nil {
						log.Fatalf("Failed to serialize application to JSON. Error: %v\n", err)
					}

					println(string(jsonBytes))
				}
			},
		},
		{
			Name:      "user",
			ShortName: "u",
			Usage:     "User CRUD operations",
			Action: func(c *cli.Context) {
				crud := c.Args().First()

				if crud == "create" {
					level := c.Args().Get(1)
					name := c.Args().Get(2)
					password := c.Args().Get(3)
					appId := c.Args().Get(4)

					user, err := resourcedmaster_dao.NewUser(rm.store, name, password)
					if err != nil {
						log.Fatalf("Failed to create a new user. Error: %v\n", err)
					}

					user.Level = level
					user.ApplicationId = appId

					err = user.Save()
					if err != nil {
						log.Fatalf("Failed to save the new user. Error: %v\n", err)
					}

					jsonBytes, err := json.Marshal(user)
					if err != nil {
						log.Fatalf("Failed to serialize user to JSON. Error: %v\n", err)
					}

					println(string(jsonBytes))
				}
			},
		},
		{
			Name:      "access-token",
			ShortName: "token",
			Usage:     "Access Token CRUD operations",
			Action: func(c *cli.Context) {
				crud := c.Args().First()

				if crud == "create" {
					level := c.Args().Get(1)
					appId := c.Args().Get(2)

					application, err := resourcedmaster_dao.GetApplicationById(rm.store, appId)
					if err != nil {
						log.Fatalf("Failed to get application by id. Error: %v\n", err)
					}

					user, err := resourcedmaster_dao.NewAccessTokenUser(rm.store, application)
					if err != nil {
						log.Fatalf("Failed to create access token. Error: %v\n", err)
					}

					user.Level = level

					err = user.Save()
					if err != nil {
						log.Fatalf("Failed to save access token. Error: %v\n", err)
					}

					println(user.Token)
				}
			},
		},
	}

	rm.store, err = rm.storage()

	return rm, err
}

type ResourcedMaster struct {
	Env   string
	store resourcedmaster_storage.Storer
	app   *cli.App
}

func (rm *ResourcedMaster) storage() (resourcedmaster_storage.Storer, error) {
	s3AccessKey := libenv.EnvWithDefault("RESOURCED_MASTER_S3_ACCESS_KEY", "")
	s3SecretKey := libenv.EnvWithDefault("RESOURCED_MASTER_S3_SECRET_KEY", "")
	s3Region := libenv.EnvWithDefault("RESOURCED_MASTER_S3_REGION", "us-east-1")
	s3Bucket := libenv.EnvWithDefault("RESOURCED_MASTER_S3_BUCKET", "")

	return resourcedmaster_storage.NewS3(rm.Env, s3AccessKey, s3SecretKey, s3Region, s3Bucket), nil
}

func (rm *ResourcedMaster) middlewareStruct(store resourcedmaster_storage.Storer) (*interpose.Middleware, error) {
	users, err := resourcedmaster_dao.AllUsers(store)
	if err != nil {
		return nil, err
	}

	middle := interpose.New()
	middle.Use(resourcedmaster_middlewares.SetStore(store))
	middle.Use(resourcedmaster_middlewares.SetCurrentApplication(store))
	middle.Use(resourcedmaster_middlewares.AccessTokenAuth(users))

	return middle, nil
}

func (rm *ResourcedMaster) mux() *gorilla_mux.Router {
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

	router.HandleFunc("/api/app/{id:[0-9]+}/hosts", resourcedmaster_handlers.GetApiAppIdHosts).Methods("GET")
	router.HandleFunc("/api/app/{id:[0-9]+}/hosts/{name}", resourcedmaster_handlers.GetApiAppIdHostsName).Methods("GET")
	router.HandleFunc("/api/app/{id:[0-9]+}/hosts/{name}", resourcedmaster_handlers.PostApiAppIdHostsName).Methods("POST")

	return router
}

func (rm *ResourcedMaster) httpRun() {
	middle, _ := rm.middlewareStruct(rm.store)

	middle.UseHandler(rm.mux())

	serverAddress := libenv.EnvWithDefault("RESOURCED_MASTER_ADDR", ":55655")
	http.ListenAndServe(serverAddress, middle)
}

func (rm *ResourcedMaster) Run(arguments []string) error {
	return rm.app.Run(arguments)
}

func main() {
	app, err := NewResourcedMaster()
	if err != nil {
		println(err)
		os.Exit(1)
	}

	app.Run(os.Args)
}
