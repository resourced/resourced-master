package application

import (
	"net/http"
	"time"

	"github.com/didip/stopwatch"
	"github.com/didip/tollbooth"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"

	"github.com/resourced/resourced-master/handlers"
	"github.com/resourced/resourced-master/middlewares"
)

func (app *Application) NewHandlerInstruments() map[string]chan int64 {
	instruments := make(map[string]chan int64)
	instruments["GetHosts"] = make(chan int64)
	return instruments
}

// mux returns an instance of HTTP router with all the predefined rules.
func (app *Application) mux() *mux.Router {
	MustLogin := middlewares.MustLogin
	MustLoginApi := middlewares.MustLoginApi
	SetClusters := middlewares.SetClusters
	SetAccessTokens := middlewares.SetAccessTokens

	CSRFOptions := csrf.Secure(false)
	if app.GeneralConfig.HTTPS.CertFile != "" {
		CSRFOptions = csrf.Secure(true)
	}
	CSRF := csrf.Protect([]byte(app.GeneralConfig.CookieSecret), CSRFOptions)

	generalAPILimiter := tollbooth.NewLimiter(int64(app.GeneralConfig.RateLimiters.GeneralAPI), time.Second)
	signupLimiter := tollbooth.NewLimiter(int64(app.GeneralConfig.RateLimiters.PostSignup), time.Second)

	router := mux.NewRouter()

	router.HandleFunc("/signup", handlers.GetSignup).Methods("GET")
	router.Handle("/signup", tollbooth.LimitFuncHandler(signupLimiter, handlers.PostSignup)).Methods("POST")

	router.HandleFunc("/login", handlers.GetLogin).Methods("GET")
	router.HandleFunc("/login", handlers.PostLogin).Methods("POST")

	router.Handle("/", alice.New(CSRF, MustLogin, SetClusters, SetAccessTokens).Then(
		stopwatch.LatencyFuncHandler(app.HandlerInstruments["GetHosts"], []string{"GET"}, handlers.GetHosts),
	)).Methods("GET")

	router.Handle("/saved-queries", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.PostSavedQueries)).Methods("POST")
	router.Handle("/saved-queries/{id:[0-9]+}", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.PostPutDeleteSavedQueriesID)).Methods("POST", "PUT", "DELETE")

	router.Handle("/graphs", alice.New(CSRF, MustLogin, SetClusters, SetAccessTokens).ThenFunc(handlers.GetGraphs)).Methods("GET")
	router.Handle("/graphs", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.PostGraphs)).Methods("POST")
	router.Handle("/graphs/{id:[0-9]+}", alice.New(CSRF, MustLogin, SetClusters, SetAccessTokens).ThenFunc(handlers.GetPostPutDeleteGraphsID)).Methods("GET", "POST", "PUT", "DELETE")

	router.Handle("/logs", alice.New(CSRF, MustLogin, SetClusters, SetAccessTokens).ThenFunc(handlers.GetLogs)).Methods("GET")
	router.Handle("/logs/executors", alice.New(CSRF, MustLogin, SetClusters, SetAccessTokens).ThenFunc(handlers.GetLogsExecutors)).Methods("GET")

	router.Handle("/checks", alice.New(CSRF, MustLogin, SetClusters, SetAccessTokens).ThenFunc(handlers.GetChecks)).Methods("GET")
	router.Handle("/checks", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.PostChecks)).Methods("POST")
	router.Handle("/checks/{id:[0-9]+}", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.PostPutDeleteCheckID)).Methods("POST", "PUT", "DELETE")
	router.Handle("/checks/{id:[0-9]+}/silence", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.PostCheckIDSilence)).Methods("POST")

	router.Handle("/checks/{checkid:[0-9]+}/triggers", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.PostChecksTriggers)).Methods("POST")
	router.Handle("/checks/{checkid:[0-9]+}/triggers/{id:[0-9]+}", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.PostPutDeleteCheckTriggerID)).Methods("POST", "PUT", "DELETE")

	router.Handle("/users/{id:[0-9]+}", alice.New(CSRF, MustLogin).ThenFunc(handlers.PostPutDeleteUsersID)).Methods("POST", "PUT", "DELETE")

	router.HandleFunc("/users/email-verification/{token}", handlers.GetUsersEmailVerificationToken).Methods("GET")

	router.Handle("/clusters", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.GetClusters)).Methods("GET")
	router.Handle("/clusters", alice.New(CSRF, MustLogin).ThenFunc(handlers.PostClusters)).Methods("POST")
	router.Handle("/clusters/{id:[0-9]+}", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.PostPutDeleteClusterID)).Methods("POST", "PUT", "DELETE")

	router.Handle("/clusters/current", alice.New(MustLogin, SetClusters).ThenFunc(handlers.PostClustersCurrent)).Methods("POST")
	router.Handle("/clusters/{id:[0-9]+}/access-tokens", alice.New(CSRF, MustLogin).ThenFunc(handlers.PostAccessTokens)).Methods("POST")
	router.Handle("/clusters/{id:[0-9]+}/users", alice.New(CSRF, MustLogin).ThenFunc(handlers.PostPutDeleteClusterIDUsers)).Methods("POST", "PUT", "DELETE")

	router.Handle("/clusters/{clusterid:[0-9]+}/metrics", alice.New(CSRF, MustLogin).ThenFunc(handlers.PostMetrics)).Methods("POST")
	router.Handle("/clusters/{clusterid:[0-9]+}/metrics/{id:[0-9]+}", alice.New(CSRF, MustLogin).ThenFunc(handlers.PostPutDeleteMetricID)).Methods("POST", "PUT", "DELETE")

	router.Handle("/access-tokens/{id:[0-9]+}/level", alice.New(CSRF, MustLogin).ThenFunc(handlers.PostAccessTokensLevel)).Methods("POST")
	router.Handle("/access-tokens/{id:[0-9]+}/enabled", alice.New(CSRF, MustLogin).ThenFunc(handlers.PostAccessTokensEnabled)).Methods("POST")

	router.Handle("/api/hosts", alice.New(MustLoginApi).Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiHosts))).Methods("GET")
	router.Handle("/api/hosts", alice.New(MustLoginApi).ThenFunc(handlers.PostApiHosts)).Methods("POST")

	router.Handle("/api/graphs/{id:[0-9]+}/metrics", alice.New(MustLoginApi).Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.PutApiGraphsIDMetrics))).Methods("PUT")

	router.Handle("/api/metrics/{id:[0-9]+}/hosts/{host}", alice.New(MustLoginApi).Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiTSMetricsByHost))).Methods("GET")
	router.Handle("/api/metrics/{id:[0-9]+}/hosts/{host}/15min", alice.New(MustLoginApi).Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiTSMetricsByHost15Min))).Methods("GET")

	router.Handle("/api/metrics/{id:[0-9]+}", alice.New(MustLoginApi).Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiTSMetrics))).Methods("GET")
	router.Handle("/api/metrics/{id:[0-9]+}/15min", alice.New(MustLoginApi).Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiTSMetrics15Min))).Methods("GET")

	router.Handle(`/api/events`, alice.New(MustLoginApi).Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.PostApiEvents))).Methods("POST")
	router.Handle(`/api/events/{id:[0-9]+}`, alice.New(MustLoginApi).Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.DeleteApiEventsID))).Methods("DELETE")
	router.Handle(`/api/events/line`, alice.New(MustLoginApi).Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiEventsLine))).Methods("GET")
	router.Handle(`/api/events/band`, alice.New(MustLoginApi).Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiEventsBand))).Methods("GET")

	router.Handle(`/api/executors`, alice.New(MustLoginApi).ThenFunc(handlers.PostApiExecutors)).Methods("POST")

	router.Handle(`/api/logs`, alice.New(MustLoginApi).Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiLogs))).Methods("GET")
	router.Handle(`/api/logs`, alice.New(MustLoginApi).ThenFunc(handlers.PostApiLogs)).Methods("POST")

	router.Handle(`/api/logs/executors`, alice.New(MustLoginApi).Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiLogsExecutors))).Methods("GET")

	router.Handle("/api/metadata", alice.New(MustLoginApi).Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiMetadata))).Methods("GET")
	router.Handle(`/api/metadata/{key}`, alice.New(MustLoginApi).Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.PostApiMetadataKey))).Methods("POST")
	router.Handle(`/api/metadata/{key}`, alice.New(MustLoginApi).Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.DeleteApiMetadataKey))).Methods("DELETE")
	router.Handle(`/api/metadata/{key}`, alice.New(MustLoginApi).Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiMetadataKey))).Methods("GET")

	// Path of static files must be last!
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("static")))

	return router
}
