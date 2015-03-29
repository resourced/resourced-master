package main

import (
	"encoding/gob"
	"fmt"
	"github.com/GeertJohan/go.rice"
	"github.com/carbocation/interpose"
	gorilla_mux "github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	resourcedmaster_dal "github.com/resourced/resourced-master/dal"
	resourcedmaster_handlers "github.com/resourced/resourced-master/handlers"
	"github.com/resourced/resourced-master/libenv"
	resourcedmaster_middlewares "github.com/resourced/resourced-master/middlewares"
	"github.com/stretchr/graceful"
	"net/http"
	"os"
	"os/user"
	"time"
)

func registerToGob() {
	gob.Register(&resourcedmaster_dal.UserRow{})
}

// NewResourcedMaster is the constructor for riceBoxes struct.
func NewriceBoxes(gorice *rice.Config) (map[string]*rice.Box, error) {
	riceBoxes := make(map[string]*rice.Box)

	for _, boxName := range []string{"templates", "static"} {
		box, err := gorice.FindBox(boxName)
		if err != nil {
			return riceBoxes, err
		}

		riceBoxes[boxName] = box
	}

	return riceBoxes, nil
}

// NewResourcedMaster is the constructor for ResourcedMaster struct.
func NewResourcedMaster() (*ResourcedMaster, error) {
	registerToGob()

	u, err := user.Current()
	if err != nil {
		return nil, err
	}

	dbPath := libenv.EnvWithDefault("RESOURCED_MASTER_DSN", fmt.Sprintf("postgres://%v@localhost:5432/resourced-master?sslmode=disable", u.Username))

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

	cookieStoreSecret := libenv.EnvWithDefault("RESOURCED_MASTER_COOKIE_SECRET", "T0PS3CR3T")

	rm := &ResourcedMaster{}
	rm.db = db
	rm.riceBoxes = riceBoxes
	rm.cookieStore = sessions.NewCookieStore([]byte(cookieStoreSecret))

	return rm, err
}

// ResourcedMaster is the application object that runs HTTP server.
type ResourcedMaster struct {
	db          *sqlx.DB
	riceBoxes   map[string]*rice.Box
	cookieStore *sessions.CookieStore
}

func (rm *ResourcedMaster) middlewareStruct() (*interpose.Middleware, error) {
	// users, err := resourcedmaster_dal.NewUser(rm.db).AllUsers(nil)
	// if err != nil {
	// 	return nil, err
	// }

	middle := interpose.New()
	middle.Use(resourcedmaster_middlewares.SetDB(rm.db))
	middle.Use(resourcedmaster_middlewares.SetRiceBoxes(rm.riceBoxes))
	middle.Use(resourcedmaster_middlewares.SetCookieStore(rm.cookieStore))
	middle.Use(resourcedmaster_middlewares.SetCurrentApplication(rm.db))
	middle.Use(resourcedmaster_middlewares.CheckUserSession())
	// middle.Use(resourcedmaster_middlewares.AccessTokenAuth(users))

	middle.UseHandler(rm.mux())

	return middle, nil
}

func (rm *ResourcedMaster) mux() *gorilla_mux.Router {
	router := gorilla_mux.NewRouter()

	router.HandleFunc("/", resourcedmaster_handlers.GetDashboard).Methods("GET")

	router.HandleFunc("/signup", resourcedmaster_handlers.GetSignup).Methods("GET")
	router.HandleFunc("/signup", resourcedmaster_handlers.PostSignup).Methods("POST")
	router.HandleFunc("/login", resourcedmaster_handlers.GetLogin).Methods("GET")
	router.HandleFunc("/login", resourcedmaster_handlers.PostLogin).Methods("POST")
	router.HandleFunc("/logout", resourcedmaster_handlers.GetLogout).Methods("GET")

	router.HandleFunc("/applications/create", resourcedmaster_handlers.GetApplicationsCreate).Methods("GET")
	router.HandleFunc("/applications", resourcedmaster_handlers.PostApplications).Methods("POST")

	// Put static files path last!
	router.PathPrefix("/").Handler(http.FileServer(rm.riceBoxes["static"].HTTPBox()))

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
	certFile := libenv.EnvWithDefault("RESOURCED_MASTER_CERT_FILE", "")
	keyFile := libenv.EnvWithDefault("RESOURCED_MASTER_KEY_FILE", "")
	requestTimeoutString := libenv.EnvWithDefault("RESOURCED_MASTER_REQUEST_TIMEOUT", "7s")

	requestTimeout, err := time.ParseDuration(requestTimeoutString)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	srv := &graceful.Server{
		Timeout: requestTimeout,
		Server:  &http.Server{Addr: serverAddress, Handler: middle},
	}

	if certFile != "" && keyFile != "" {
		srv.ListenAndServeTLS(certFile, keyFile)
	} else {
		srv.ListenAndServe()
	}
}
