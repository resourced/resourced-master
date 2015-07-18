package main

import (
	"encoding/gob"
	"fmt"
	"github.com/carbocation/interpose"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mattes/migrate/migrate"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/handlers"
	"github.com/resourced/resourced-master/libenv"
	"github.com/resourced/resourced-master/libunix"
	"github.com/resourced/resourced-master/middlewares"
	"github.com/stretchr/graceful"
	"net/http"
	"os"
	"time"
)

func registerToGob() {
	gob.Register(&dal.UserRow{})
}

// NewResourcedMaster is the constructor for ResourcedMaster struct.
func NewResourcedMaster() (*ResourcedMaster, error) {
	registerToGob()

	u, err := libunix.CurrentUser()
	if err != nil {
		return nil, err
	}

	dsn := libenv.EnvWithDefault("RESOURCED_MASTER_DSN", fmt.Sprintf("postgres://%v@localhost:5432/resourced-master?sslmode=disable", u))

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	cookieStoreSecret := libenv.EnvWithDefault("RESOURCED_MASTER_COOKIE_SECRET", "T0PS3CR3T")

	rm := &ResourcedMaster{}
	rm.dsn = dsn
	rm.db = db
	rm.cookieStore = sessions.NewCookieStore([]byte(cookieStoreSecret))

	return rm, err
}

// ResourcedMaster is the application object that runs HTTP server.
type ResourcedMaster struct {
	dsn         string
	db          *sqlx.DB
	cookieStore *sessions.CookieStore
}

func (rm *ResourcedMaster) middlewareStruct() (*interpose.Middleware, error) {
	middle := interpose.New()
	middle.Use(middlewares.SetDB(rm.db))
	middle.Use(middlewares.SetCookieStore(rm.cookieStore))
	middle.Use(middlewares.SetWSConnections(rm.cookieStore))

	middle.UseHandler(rm.mux())

	return middle, nil
}

func (rm *ResourcedMaster) mux() *mux.Router {
	MustLogin := middlewares.MustLogin
	MustLoginApi := middlewares.MustLoginApi

	router := mux.NewRouter()

	router.Handle("/", MustLogin(http.HandlerFunc(handlers.GetHosts))).Methods("GET")

	router.HandleFunc("/signup", handlers.GetSignup).Methods("GET")
	router.HandleFunc("/signup", handlers.PostSignup).Methods("POST")
	router.HandleFunc("/login", handlers.GetLogin).Methods("GET")
	router.HandleFunc("/login", handlers.PostLogin).Methods("POST")
	router.HandleFunc("/logout", handlers.GetLogout).Methods("GET")

	router.Handle("/users/{id:[0-9]+}", MustLogin(http.HandlerFunc(handlers.PostPutDeleteUsersID))).Methods("POST", "PUT", "DELETE")

	router.Handle("/access-tokens", MustLogin(http.HandlerFunc(handlers.GetAccessTokens))).Methods("GET")
	router.Handle("/access-tokens", MustLogin(http.HandlerFunc(handlers.PostAccessTokens))).Methods("POST")

	router.Handle("/access-tokens/{id:[0-9]+}/level", MustLogin(http.HandlerFunc(handlers.PostAccessTokensLevel))).Methods("POST")
	router.Handle("/access-tokens/{id:[0-9]+}/enabled", MustLogin(http.HandlerFunc(handlers.PostAccessTokensEnabled))).Methods("POST")

	router.Handle("/saved-queries", MustLogin(http.HandlerFunc(handlers.PostSavedQueries))).Methods("POST")
	router.Handle("/saved-queries/{id:[0-9]+}", MustLogin(http.HandlerFunc(handlers.PostPutDeleteSavedQueriesID))).Methods("POST", "PUT", "DELETE")

	router.HandleFunc("/api/ws", handlers.ApiWS)

	router.Handle("/api/hosts", MustLoginApi(http.HandlerFunc(handlers.GetApiHosts))).Methods("GET")
	router.Handle("/api/hosts", MustLoginApi(http.HandlerFunc(handlers.PostApiHosts))).Methods("POST")

	// Path of static files must be last!
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("static")))

	return router
}

func (rm *ResourcedMaster) migrateUp() (err []error, ok bool) {
	return migrate.UpSync(rm.dsn, "./migrations")
}

func main() {
	app, err := NewResourcedMaster()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	// Migrate up
	errs, ok := app.migrateUp()
	if !ok {
		for _, err := range errs {
			println(err.Error())
		}
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
