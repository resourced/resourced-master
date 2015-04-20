package main

import (
	"encoding/gob"
	"fmt"
	"github.com/carbocation/interpose"
	gorilla_mux "github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	resourcedmaster_dal "github.com/resourced/resourced-master/dal"
	resourcedmaster_handlers "github.com/resourced/resourced-master/handlers"
	"github.com/resourced/resourced-master/libenv"
	"github.com/resourced/resourced-master/libunix"
	resourcedmaster_middlewares "github.com/resourced/resourced-master/middlewares"
	"github.com/stretchr/graceful"
	"net/http"
	"os"
	"time"
)

func registerToGob() {
	gob.Register(&resourcedmaster_dal.UserRow{})
}

// NewResourcedMaster is the constructor for ResourcedMaster struct.
func NewResourcedMaster() (*ResourcedMaster, error) {
	registerToGob()

	u, err := libunix.CurrentUser()
	if err != nil {
		return nil, err
	}

	dbPath := libenv.EnvWithDefault("RESOURCED_MASTER_DSN", fmt.Sprintf("postgres://%v@localhost:5432/resourced-master?sslmode=disable", u))

	db, err := sqlx.Connect("postgres", dbPath)
	if err != nil {
		return nil, err
	}

	cookieStoreSecret := libenv.EnvWithDefault("RESOURCED_MASTER_COOKIE_SECRET", "T0PS3CR3T")

	rm := &ResourcedMaster{}
	rm.db = db
	rm.cookieStore = sessions.NewCookieStore([]byte(cookieStoreSecret))

	return rm, err
}

// ResourcedMaster is the application object that runs HTTP server.
type ResourcedMaster struct {
	db          *sqlx.DB
	cookieStore *sessions.CookieStore
}

func (rm *ResourcedMaster) middlewareStruct() (*interpose.Middleware, error) {
	middle := interpose.New()
	middle.Use(resourcedmaster_middlewares.SetDB(rm.db))
	middle.Use(resourcedmaster_middlewares.SetCookieStore(rm.cookieStore))

	middle.UseHandler(rm.mux())

	return middle, nil
}

func (rm *ResourcedMaster) mux() *gorilla_mux.Router {
	MustLogin := resourcedmaster_middlewares.MustLogin
	MustLoginApi := resourcedmaster_middlewares.MustLoginApi

	router := gorilla_mux.NewRouter()

	router.Handle("/", MustLogin(http.HandlerFunc(resourcedmaster_handlers.GetHosts))).Methods("GET")

	router.HandleFunc("/signup", resourcedmaster_handlers.GetSignup).Methods("GET")
	router.HandleFunc("/signup", resourcedmaster_handlers.PostSignup).Methods("POST")
	router.HandleFunc("/login", resourcedmaster_handlers.GetLogin).Methods("GET")
	router.HandleFunc("/login", resourcedmaster_handlers.PostLogin).Methods("POST")
	router.HandleFunc("/logout", resourcedmaster_handlers.GetLogout).Methods("GET")

	router.Handle("/users/{id:[0-9]+}", MustLogin(http.HandlerFunc(resourcedmaster_handlers.PostPutDeleteUsersID))).Methods("POST", "PUT", "DELETE")

	router.Handle("/access-tokens", MustLogin(http.HandlerFunc(resourcedmaster_handlers.GetAccessTokens))).Methods("GET")
	router.Handle("/access-tokens", MustLogin(http.HandlerFunc(resourcedmaster_handlers.PostAccessTokens))).Methods("POST")

	router.Handle("/access-tokens/{id:[0-9]+}/level", MustLogin(http.HandlerFunc(resourcedmaster_handlers.PostAccessTokensLevel))).Methods("POST")
	router.Handle("/access-tokens/{id:[0-9]+}/enabled", MustLogin(http.HandlerFunc(resourcedmaster_handlers.PostAccessTokensEnabled))).Methods("POST")

	router.Handle("/api/hosts", MustLoginApi(http.HandlerFunc(resourcedmaster_handlers.GetApiHosts))).Methods("GET")
	router.Handle("/api/hosts", MustLoginApi(http.HandlerFunc(resourcedmaster_handlers.PostApiHosts))).Methods("POST")

	// Path of static files must be last!
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("static")))

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
	requestTimeoutString := libenv.EnvWithDefault("RESOURCED_MASTER_REQUEST_TIMEOUT", "1s")

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
