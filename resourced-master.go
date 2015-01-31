package main

import (
	"github.com/carbocation/interpose"
	"github.com/codegangsta/cli"
	gorilla_mux "github.com/gorilla/mux"
	resourcedmaster_dao "github.com/resourced/resourced-master/dao"
	resourcedmaster_handlers "github.com/resourced/resourced-master/handlers"
	"github.com/resourced/resourced-master/libenv"
	resourcedmaster_middlewares "github.com/resourced/resourced-master/middlewares"
	resourcedmaster_storage "github.com/resourced/resourced-master/storage"
	"net/http"
	"os"
)

func NewResourcedMaster() (*ResourcedMaster, error) {
	var err error

	rm := &ResourcedMaster{}

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
						println(err)
						os.Exit(1)
					}

					err = application.Save()
					if err != nil {
						println(err)
						os.Exit(1)
					}
				}
			},
		},
	}

	rm.store, err = rm.storage()

	return rm, err
}

type ResourcedMaster struct {
	store resourcedmaster_storage.Storer
	app   *cli.App
}

func (rm *ResourcedMaster) storage() (resourcedmaster_storage.Storer, error) {
	env := libenv.EnvWithDefault("RESOURCED_MASTER_ENV", "test")
	s3AccessKey := libenv.EnvWithDefault("RESOURCED_MASTER_S3_ACCESS_KEY", "")
	s3SecretKey := libenv.EnvWithDefault("RESOURCED_MASTER_S3_SECRET_KEY", "")
	s3Region := libenv.EnvWithDefault("RESOURCED_MASTER_S3_REGION", "us-east-1")
	s3Bucket := libenv.EnvWithDefault("RESOURCED_MASTER_S3_BUCKET", "")

	return resourcedmaster_storage.NewS3(env, s3AccessKey, s3SecretKey, s3Region, s3Bucket), nil
}

func (rm *ResourcedMaster) middlewareStruct(store resourcedmaster_storage.Storer) (*interpose.Middleware, error) {
	users, err := resourcedmaster_dao.AllUsers(store)
	if err != nil {
		return nil, err
	}

	middle := interpose.New()
	middle.Use(resourcedmaster_middlewares.SetStore(store))
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
	router.HandleFunc("/api/app/{id:[0-9]+}/hosts/hardware-addr/{address}", resourcedmaster_handlers.GetApiAppIdHostsHardwareAddr).Methods("GET")
	router.HandleFunc("/api/app/{id:[0-9]+}/hosts/ip-addr/{address}", resourcedmaster_handlers.GetApiAppIdHostsIpAddr).Methods("GET")

	router.HandleFunc("/api/app/{id:[0-9]+}/{reader-or-writer}/{path}", resourcedmaster_handlers.PostApiAppIdReaderWriter).Methods("POST")

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
