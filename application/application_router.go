package application

import (
	"net/http"
	"time"

	"github.com/didip/stopwatch"
	"github.com/didip/tollbooth"
	"github.com/gorilla/mux"

	"github.com/resourced/resourced-master/handlers"
	"github.com/resourced/resourced-master/middlewares"
)

func (app *Application) NewHandlerInstruments() map[string]chan int64 {
	instruments := make(map[string]chan int64)
	for _, key := range []string{"GetHosts", "GetLogs", "GetLogsExecutors"} {
		instruments[key] = make(chan int64)
	}
	return instruments
}

// mux returns an instance of HTTP router with all the predefined rules.
func (app *Application) mux() *mux.Router {
	SetAccessTokens := middlewares.SetAccessTokens
	MinimumMiddlewareChain := middlewares.MinimumMiddlewareChain
	MinimumAPIMiddlewareChain := middlewares.MinimumAPIMiddlewareChain

	useHTTPS := app.GeneralConfig.HTTPS.CertFile != "" && app.GeneralConfig.HTTPS.KeyFile != ""
	CSRF := middlewares.CSRFMiddleware(useHTTPS, app.GeneralConfig.CookieSecret)

	generalAPILimiter := tollbooth.NewLimiter(int64(app.GeneralConfig.RateLimiters.GeneralAPI), time.Second)
	signupLimiter := tollbooth.NewLimiter(int64(app.GeneralConfig.RateLimiters.PostSignup), time.Second)

	router := mux.NewRouter()

	router.HandleFunc("/signup", handlers.GetSignup).Methods("GET")
	router.Handle("/signup", tollbooth.LimitFuncHandler(signupLimiter, handlers.PostSignup)).Methods("POST")

	router.HandleFunc("/login", handlers.GetLogin).Methods("GET")
	router.HandleFunc("/login", handlers.PostLogin).Methods("POST")

	router.Handle("/", MinimumMiddlewareChain(CSRF).Append(SetAccessTokens).Then(
		stopwatch.LatencyFuncHandler(app.HandlerInstruments["GetHosts"], []string{"GET"}, handlers.GetHosts),
	)).Methods("GET")

	router.Handle("/saved-queries", MinimumMiddlewareChain(CSRF).ThenFunc(handlers.PostSavedQueries)).Methods("POST")
	router.Handle("/saved-queries/{id:[0-9]+}", MinimumMiddlewareChain(CSRF).ThenFunc(handlers.PostPutDeleteSavedQueriesID)).Methods("POST", "PUT", "DELETE")

	router.Handle("/graphs", MinimumMiddlewareChain(CSRF).Append(SetAccessTokens).ThenFunc(handlers.GetGraphs)).Methods("GET")
	router.Handle("/graphs", MinimumMiddlewareChain(CSRF).ThenFunc(handlers.PostGraphs)).Methods("POST")
	router.Handle("/graphs/{id:[0-9]+}", MinimumMiddlewareChain(CSRF).Append(SetAccessTokens).ThenFunc(handlers.GetPostPutDeleteGraphsID)).Methods("GET", "POST", "PUT", "DELETE")

	router.Handle("/logs", MinimumMiddlewareChain(CSRF).Append(SetAccessTokens).Then(
		stopwatch.LatencyFuncHandler(app.HandlerInstruments["GetLogs"], []string{"GET"}, handlers.GetLogs),
	)).Methods("GET")

	router.Handle("/logs/executors", MinimumMiddlewareChain(CSRF).Append(SetAccessTokens).Then(
		stopwatch.LatencyFuncHandler(app.HandlerInstruments["GetLogsExecutors"], []string{"GET"}, handlers.GetLogsExecutors),
	)).Methods("GET")

	router.Handle("/checks", MinimumMiddlewareChain(CSRF).Append(SetAccessTokens).ThenFunc(handlers.GetChecks)).Methods("GET")
	router.Handle("/checks", MinimumMiddlewareChain(CSRF).ThenFunc(handlers.PostChecks)).Methods("POST")
	router.Handle("/checks/{id:[0-9]+}", MinimumMiddlewareChain(CSRF).ThenFunc(handlers.PostPutDeleteCheckID)).Methods("POST", "PUT", "DELETE")
	router.Handle("/checks/{id:[0-9]+}/silence", MinimumMiddlewareChain(CSRF).ThenFunc(handlers.PostCheckIDSilence)).Methods("POST")

	router.Handle("/checks/{checkid:[0-9]+}/triggers", MinimumMiddlewareChain(CSRF).ThenFunc(handlers.PostChecksTriggers)).Methods("POST")
	router.Handle("/checks/{checkid:[0-9]+}/triggers/{id:[0-9]+}", MinimumMiddlewareChain(CSRF).ThenFunc(handlers.PostPutDeleteCheckTriggerID)).Methods("POST", "PUT", "DELETE")

	router.Handle("/users/{id:[0-9]+}", MinimumMiddlewareChain(CSRF).ThenFunc(handlers.PostPutDeleteUsersID)).Methods("POST", "PUT", "DELETE")

	router.HandleFunc("/users/email-verification/{token}", handlers.GetUsersEmailVerificationToken).Methods("GET")

	router.Handle("/clusters", MinimumMiddlewareChain(CSRF).ThenFunc(handlers.GetClusters)).Methods("GET")
	router.Handle("/clusters", MinimumMiddlewareChain(CSRF).ThenFunc(handlers.PostClusters)).Methods("POST")
	router.Handle("/clusters/{id:[0-9]+}", MinimumMiddlewareChain(CSRF).ThenFunc(handlers.PostPutDeleteClusterID)).Methods("POST", "PUT", "DELETE")

	router.Handle("/clusters/current", MinimumMiddlewareChain(CSRF).ThenFunc(handlers.PostClustersCurrent)).Methods("POST")
	router.Handle("/clusters/{id:[0-9]+}/access-tokens", MinimumMiddlewareChain(CSRF).ThenFunc(handlers.PostAccessTokens)).Methods("POST")
	router.Handle("/clusters/{id:[0-9]+}/users", MinimumMiddlewareChain(CSRF).ThenFunc(handlers.PostPutDeleteClusterIDUsers)).Methods("POST", "PUT", "DELETE")

	router.Handle("/clusters/{clusterid:[0-9]+}/metrics", MinimumMiddlewareChain(CSRF).ThenFunc(handlers.PostMetrics)).Methods("POST")
	router.Handle("/clusters/{clusterid:[0-9]+}/metrics/{id:[0-9]+}", MinimumMiddlewareChain(CSRF).ThenFunc(handlers.PostPutDeleteMetricID)).Methods("POST", "PUT", "DELETE")

	router.Handle("/access-tokens/{id:[0-9]+}/level", MinimumMiddlewareChain(CSRF).ThenFunc(handlers.PostAccessTokensLevel)).Methods("POST")
	router.Handle("/access-tokens/{id:[0-9]+}/enabled", MinimumMiddlewareChain(CSRF).ThenFunc(handlers.PostAccessTokensEnabled)).Methods("POST")

	router.Handle("/api/hosts", MinimumAPIMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiHosts))).Methods("GET")
	router.Handle("/api/hosts", MinimumAPIMiddlewareChain().ThenFunc(handlers.PostApiHosts)).Methods("POST")

	router.Handle("/api/graphs/{id:[0-9]+}/metrics", MinimumAPIMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.PutApiGraphsIDMetrics))).Methods("PUT")

	router.Handle("/api/metrics/{id:[0-9]+}/hosts/{host}", MinimumAPIMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiTSMetricsByHost))).Methods("GET")
	router.Handle("/api/metrics/{id:[0-9]+}/hosts/{host}/15min", MinimumAPIMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiTSMetricsByHost15Min))).Methods("GET")

	router.Handle("/api/metrics/{id:[0-9]+}", MinimumAPIMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiTSMetrics))).Methods("GET")
	router.Handle("/api/metrics/{id:[0-9]+}/15min", MinimumAPIMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiTSMetrics15Min))).Methods("GET")

	router.Handle(`/api/events`, MinimumAPIMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.PostApiEvents))).Methods("POST")
	router.Handle(`/api/events/{id:[0-9]+}`, MinimumAPIMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.DeleteApiEventsID))).Methods("DELETE")
	router.Handle(`/api/events/line`, MinimumAPIMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiEventsLine))).Methods("GET")
	router.Handle(`/api/events/band`, MinimumAPIMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiEventsBand))).Methods("GET")

	router.Handle(`/api/executors`, MinimumAPIMiddlewareChain().ThenFunc(handlers.PostApiExecutors)).Methods("POST")

	router.Handle(`/api/logs`, MinimumAPIMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiLogs))).Methods("GET")
	router.Handle(`/api/logs`, MinimumAPIMiddlewareChain().ThenFunc(handlers.PostApiLogs)).Methods("POST")

	router.Handle(`/api/logs/executors`, MinimumAPIMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiLogsExecutors))).Methods("GET")

	router.Handle("/api/metadata", MinimumAPIMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiMetadata))).Methods("GET")
	router.Handle(`/api/metadata/{key}`, MinimumAPIMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.PostApiMetadataKey))).Methods("POST")
	router.Handle(`/api/metadata/{key}`, MinimumAPIMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.DeleteApiMetadataKey))).Methods("DELETE")
	router.Handle(`/api/metadata/{key}`, MinimumAPIMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiMetadataKey))).Methods("GET")

	// Path of static files must be last!
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("static")))

	return router
}
