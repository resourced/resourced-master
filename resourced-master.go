package main

import (
	"fmt"
	"github.com/GeertJohan/go.rice"
	"github.com/carbocation/interpose"
	gorilla_mux "github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	resourcedmaster_dal "github.com/resourced/resourced-master/dal"
	resourcedmaster_handlers "github.com/resourced/resourced-master/handlers"
	"github.com/resourced/resourced-master/libenv"
	resourcedmaster_middlewares "github.com/resourced/resourced-master/middlewares"
	"net/http"
	"os"
	"os/user"
)

// NewResourcedMaster is the constructor for riceBoxes struct.
func NewriceBoxes(gorice *rice.Config) (map[string]*rice.Box, error) {
	templatesBox, err := gorice.FindBox("templates")
	if err != nil {
		return nil, err
	}

	riceBoxes := make(map[string]*rice.Box)
	riceBoxes["templates"] = templatesBox

	return riceBoxes, nil
}

// NewResourcedMaster is the constructor for ResourcedMaster struct.
func NewResourcedMaster() (*ResourcedMaster, error) {
	u, err := user.Current()
	if err != nil {
		return nil, err
	}

	dbPath := libenv.EnvWithDefault("RESOURCED_DB", fmt.Sprintf("postgres://%v@localhost:5432/resourced-master?sslmode=disable", u.Username))

	db, err := sqlx.Connect("postgres", dbPath)
	if err != nil {
		return nil, err
	}

	gorice := &rice.Config{
		LocateOrder: []rice.LocateMethod{rice.LocateEmbedded, rice.LocateAppended, rice.LocateFS},
	}

	riceBoxes, err := NewriceBoxes(gorice)
	if err != nil {
		return nil, err
	}

	rm := &ResourcedMaster{}
	rm.db = db
	rm.riceBoxes = riceBoxes

	return rm, err
}

// ResourcedMaster is the application object that runs HTTP server.
type ResourcedMaster struct {
	db        *sqlx.DB
	riceBoxes map[string]*rice.Box
}

func (rm *ResourcedMaster) middlewareStruct() (*interpose.Middleware, error) {
	users, err := resourcedmaster_dal.NewUser(rm.db).AllUsers(nil)
	if err != nil {
		return nil, err
	}

	middle := interpose.New()
	middle.Use(resourcedmaster_middlewares.SetDB(rm.db))
	middle.Use(resourcedmaster_middlewares.SetRiceBoxes(rm.riceBoxes))
	middle.Use(resourcedmaster_middlewares.SetCurrentApplication(rm.db))
	middle.Use(resourcedmaster_middlewares.AccessTokenAuth(users))

	middle.UseHandler(rm.mux())

	return middle, nil
}

func (rm *ResourcedMaster) mux() *gorilla_mux.Router {
	router := gorilla_mux.NewRouter()

	router.HandleFunc("/", resourcedmaster_handlers.GetDashboard).Methods("GET")

	router.HandleFunc("/signup", resourcedmaster_handlers.GetSignup).Methods("GET")
	router.HandleFunc("/login", resourcedmaster_handlers.GetLogin).Methods("GET")

	return router
}

func main() {
	app, err := NewResourcedMaster()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	middle, err := app.middlewareStruct()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	serverAddress := libenv.EnvWithDefault("RESOURCED_MASTER_ADDR", ":55655")
	http.ListenAndServe(serverAddress, middle)
}
