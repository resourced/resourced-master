package application

import (
	"fmt"
	"github.com/carbocation/interpose"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mattes/migrate/migrate"
	"github.com/resourced/resourced-master/handlers"
	"github.com/resourced/resourced-master/libenv"
	"github.com/resourced/resourced-master/libunix"
	"github.com/resourced/resourced-master/middlewares"
	"github.com/resourced/resourced-master/wstrafficker"
	"net/http"
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

	router := mux.NewRouter()

	router.Handle("/", MustLogin(http.HandlerFunc(handlers.GetHosts))).Methods("GET")
	router.Handle("/tasks", MustLogin(http.HandlerFunc(handlers.GetTasks))).Methods("GET")

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

	router.HandleFunc("/api/ws/access-tokens/{id}", handlers.ApiWSAccessToken)

	router.Handle("/api/hosts", MustLoginApi(http.HandlerFunc(handlers.GetApiHosts))).Methods("GET")
	router.Handle("/api/hosts", MustLoginApi(http.HandlerFunc(handlers.PostApiHosts))).Methods("POST")

	// Path of static files must be last!
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("static")))

	return router
}

func (app *Application) MigrateUp() (err []error, ok bool) {
	return migrate.UpSync(app.dsn, "./migrations")
}
