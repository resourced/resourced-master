package application

import (
	"fmt"
	"net/http"

	"github.com/carbocation/interpose"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/justinas/alice"
	_ "github.com/lib/pq"
	"github.com/mattes/migrate/migrate"
	"github.com/resourced/resourced-master/handlers"
	"github.com/resourced/resourced-master/libenv"
	"github.com/resourced/resourced-master/libunix"
	"github.com/resourced/resourced-master/middlewares"
	"github.com/resourced/resourced-master/wstrafficker"
)

// New is the constructor for Application struct.
func New() (*Application, error) {
	u, err := libunix.CurrentUser()
	if err != nil {
		return nil, err
	}

	dsn := libenv.EnvWithDefault("RESOURCED_MASTER_DSN", fmt.Sprintf("postgres://%v@localhost:5432/resourced-master?sslmode=disable", u))

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// As a user, you must provide your own secret
	// But make sure you keep using the same one, otherwise sessions will expire.
	cookieStoreSecret := libenv.EnvWithDefault("RESOURCED_MASTER_COOKIE_SECRET", "T0PS3CR3T")

	app := &Application{}
	app.Addr = libenv.EnvWithDefault("RESOURCED_MASTER_ADDR", ":55655")
	app.dsn = dsn
	app.db = db
	app.cookieStore = sessions.NewCookieStore([]byte(cookieStoreSecret))
	app.WSTraffickers = wstrafficker.NewWSTraffickers()

	return app, err
}

// Application is the application object that runs HTTP server.
type Application struct {
	Addr          string
	dsn           string
	db            *sqlx.DB
	cookieStore   *sessions.CookieStore
	WSTraffickers *wstrafficker.WSTraffickers
}

func (app *Application) MiddlewareStruct() (*interpose.Middleware, error) {
	middle := interpose.New()
	middle.Use(middlewares.SetAddr(app.Addr))
	middle.Use(middlewares.SetDB(app.db))
	middle.Use(middlewares.SetCookieStore(app.cookieStore))
	middle.Use(middlewares.SetWSTraffickers(app.WSTraffickers))

	middle.UseHandler(app.mux())

	return middle, nil
}

func (app *Application) mux() *mux.Router {
	MustLogin := middlewares.MustLogin
	MustLoginApi := middlewares.MustLoginApi
	SetClusters := middlewares.SetClusters

	router := mux.NewRouter()

	router.HandleFunc("/signup", handlers.GetSignup).Methods("GET")
	router.HandleFunc("/signup", handlers.PostSignup).Methods("POST")
	router.HandleFunc("/login", handlers.GetLogin).Methods("GET")
	router.HandleFunc("/login", handlers.PostLogin).Methods("POST")
	router.HandleFunc("/logout", handlers.GetLogout).Methods("GET")

	router.Handle("/", alice.New(MustLogin, SetClusters).ThenFunc(handlers.GetHosts)).Methods("GET")

	router.Handle("/metadata", alice.New(MustLogin, SetClusters).ThenFunc(handlers.GetMetadata)).Methods("GET")
	router.Handle("/metadata", alice.New(MustLogin, SetClusters).ThenFunc(handlers.PostMetadata)).Methods("POST")

	router.Handle("/tasks", alice.New(MustLogin, SetClusters).ThenFunc(handlers.GetTasks)).Methods("GET")

	router.Handle("/users/{id:[0-9]+}", alice.New(MustLogin).ThenFunc(handlers.PostPutDeleteUsersID)).Methods("POST", "PUT", "DELETE")

	router.Handle("/clusters", alice.New(MustLogin, SetClusters).ThenFunc(handlers.GetClusters)).Methods("GET")
	router.Handle("/clusters", alice.New(MustLogin).ThenFunc(handlers.PostClusters)).Methods("POST")

	router.Handle("/clusters/current", alice.New(MustLogin, SetClusters).ThenFunc(handlers.PostClustersCurrent)).Methods("POST")

	router.Handle("/clusters/{id:[0-9]+}/access-tokens", alice.New(MustLogin).ThenFunc(handlers.PostAccessTokens)).Methods("POST")

	router.Handle("/access-tokens/{id:[0-9]+}/level", alice.New(MustLogin).ThenFunc(handlers.PostAccessTokensLevel)).Methods("POST")
	router.Handle("/access-tokens/{id:[0-9]+}/enabled", alice.New(MustLogin).ThenFunc(handlers.PostAccessTokensEnabled)).Methods("POST")

	router.Handle("/saved-queries", alice.New(MustLogin).ThenFunc(handlers.PostSavedQueries)).Methods("POST")
	router.Handle("/saved-queries/{id:[0-9]+}", alice.New(MustLogin).ThenFunc(handlers.PostPutDeleteSavedQueriesID)).Methods("POST", "PUT", "DELETE")

	router.HandleFunc("/api/ws/access-tokens/{id}", handlers.ApiWSAccessToken)

	router.Handle("/api/hosts", alice.New(MustLoginApi).ThenFunc(handlers.GetApiHosts)).Methods("GET")
	router.Handle("/api/hosts", alice.New(MustLoginApi).ThenFunc(handlers.PostApiHosts)).Methods("POST")

	router.Handle("/api/metadata", alice.New(MustLoginApi).ThenFunc(handlers.GetApiMetadata)).Methods("GET")
	router.Handle(`/api/metadata/{key:[\w\/\-\_]+}`, alice.New(MustLoginApi).ThenFunc(handlers.PostApiMetadataKey)).Methods("POST")
	router.Handle(`/api/metadata/{key:[\w\/\-\_]+}`, alice.New(MustLoginApi).ThenFunc(handlers.DeleteApiMetadataKey)).Methods("DELETE")
	router.Handle(`/api/metadata/{key:[\w\/\-\_]+}`, alice.New(MustLoginApi).ThenFunc(handlers.GetApiMetadataKey)).Methods("GET")

	router.Handle(`/api/executors`, alice.New(MustLoginApi).ThenFunc(handlers.GetApiExecutors)).Methods("GET")
	router.Handle(`/api/executors/{hostname:[\w\-\_]+}`, alice.New(MustLoginApi).ThenFunc(handlers.PostApiExecutorsByHostname)).Methods("POST")

	// Path of static files must be last!
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("static")))

	return router
}

func (app *Application) MigrateUp() (err []error, ok bool) {
	return migrate.UpSync(app.dsn, "./migrations")
}
