package application

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/carbocation/interpose"
	"github.com/didip/stopwatch"
	"github.com/didip/tollbooth"
	"github.com/gorilla/mux"
	"github.com/pressly/chi"
	"gopkg.in/tylerb/graceful.v1"

	"github.com/resourced/resourced-master/handlers"
	"github.com/resourced/resourced-master/middlewares"
)

func (app *Application) newHandlerInstruments() map[string]chan int64 {
	instruments := make(map[string]chan int64)
	for _, key := range []string{"GetHosts", "GetLogs", "GetLogsExecutors"} {
		instruments[key] = make(chan int64)
	}
	return instruments
}

func (app *Application) getHandlerInstrument(key string) chan int64 {
	var instrument chan int64

	app.RLock()
	instrument = app.HandlerInstruments[key]
	app.RUnlock()

	return instrument
}

func (app *Application) mux2() *chi.Mux {
	generalAPILimiter := tollbooth.NewLimiter(int64(app.GeneralConfig.RateLimiters.GeneralAPI), time.Second)
	signupLimiter := tollbooth.NewLimiter(int64(app.GeneralConfig.RateLimiters.PostSignup), time.Second)

	useHTTPS := app.GeneralConfig.HTTPS.CertFile != "" && app.GeneralConfig.HTTPS.KeyFile != ""
	CSRF := middlewares.CSRFMiddleware(useHTTPS, app.GeneralConfig.CookieSecret)

	r := chi.NewRouter()

	// Set various info for every request.
	r.Use(middlewares.SetAddr(app.GeneralConfig.Addr))
	r.Use(middlewares.SetVIPAddr(app.GeneralConfig.VIPAddr))
	r.Use(middlewares.SetVIPProtocol(app.GeneralConfig.VIPProtocol))
	r.Use(middlewares.SetDBs(app.DBConfig))
	r.Use(middlewares.SetCookieStore(app.cookieStore))
	r.Use(middlewares.SetMailers(app.Mailers))
	r.Use(middlewares.SetMessageBus(app.MessageBus))

	r.Get("/signup", handlers.GetSignup)
	r.Post("/signup", tollbooth.LimitFuncHandler(signupLimiter, handlers.PostSignup).(http.HandlerFunc))

	r.Get("/login", handlers.GetLogin)
	r.Post("/login", handlers.PostLogin)

	r.Route("/", func(r chi.Router) {
		r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember, middlewares.SetAccessTokens)
		r.Get("/", stopwatch.LatencyFuncHandler(app.getHandlerInstrument("GetHosts"), []string{"GET"}, handlers.GetHosts).(http.HandlerFunc))
	})

	r.Route("/saved-queries", func(r chi.Router) {
		r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember)
		r.Post("/:id", handlers.PostPutDeleteSavedQueriesID)
		r.Put("/:id", handlers.PostPutDeleteSavedQueriesID)
		r.Delete("/:id", handlers.PostPutDeleteSavedQueriesID)
	})

	r.Route("/graphs", func(r chi.Router) {
		r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember, middlewares.SetAccessTokens)
		r.Get("/", handlers.GetGraphs)
		r.Post("/", handlers.PostGraphs)
		r.Get("/:id", handlers.GetPostPutDeleteGraphsID)
		r.Post("/:id", handlers.GetPostPutDeleteGraphsID)
		r.Put("/:id", handlers.GetPostPutDeleteGraphsID)
		r.Delete("/:id", handlers.GetPostPutDeleteGraphsID)
	})

	r.Route("/logs", func(r chi.Router) {
		r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember, middlewares.SetAccessTokens)
		r.Get("/", stopwatch.LatencyFuncHandler(app.getHandlerInstrument("GetLogs"), []string{"GET"}, handlers.GetLogs).(http.HandlerFunc))
		r.Get("/executors", stopwatch.LatencyFuncHandler(app.getHandlerInstrument("GetLogsExecutors"), []string{"GET"}, handlers.GetLogsExecutors).(http.HandlerFunc))
	})

	r.Route("/checks", func(r chi.Router) {
		r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember, middlewares.SetAccessTokens)
		r.Get("/", handlers.GetChecks)
		r.Post("/", handlers.PostChecks)

		r.Route("/:checkID", func(r chi.Router) {
			r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember)
			r.Post("/", handlers.PostPutDeleteCheckID)
			r.Put("/", handlers.PostPutDeleteCheckID)
			r.Delete("/", handlers.PostPutDeleteCheckID)
			r.Post("/silence", handlers.PostCheckIDSilence)

			r.Route("/triggers", func(r chi.Router) {
				r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember)
				r.Post("/", handlers.PostChecksTriggers)
				r.Post("/:triggerID", handlers.PostPutDeleteCheckTriggerID)
				r.Put("/:triggerID", handlers.PostPutDeleteCheckTriggerID)
				r.Delete("/:triggerID", handlers.PostPutDeleteCheckTriggerID)
			})
		})
	})

	r.Route("/users", func(r chi.Router) {
		r.Route("/:id", func(r chi.Router) {
			r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember)
			r.Post("/", handlers.PostPutDeleteUsersID)
			r.Put("/", handlers.PostPutDeleteUsersID)
			r.Delete("/", handlers.PostPutDeleteUsersID)
		})

		r.Get("/email-verification/:token", handlers.GetUsersEmailVerificationToken)
	})

	r.Route("/clusters", func(r chi.Router) {
		r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember)
		r.Get("/", handlers.GetClusters)
		r.Post("/", handlers.PostClusters)
		r.Post("/current", handlers.PostClustersCurrent)

		r.Route("/:clusterID", func(r chi.Router) {
			r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember)
			r.Post("/", handlers.PostPutDeleteClusterID)
			r.Put("/", handlers.PostPutDeleteClusterID)
			r.Delete("/", handlers.PostPutDeleteClusterID)

			r.Post("/access-tokens", handlers.PostAccessTokens)
			r.Post("/users", handlers.PostPutDeleteClusterIDUsers)
			r.Put("/users", handlers.PostPutDeleteClusterIDUsers)
			r.Delete("/users", handlers.PostPutDeleteClusterIDUsers)

			r.Route("/metrics", func(r chi.Router) {
				r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember)
				r.Post("/", handlers.PostMetrics)
				r.Post("/:metricID", handlers.PostPutDeleteMetricID)
				r.Put("/:metricID", handlers.PostPutDeleteMetricID)
				r.Delete("/:metricID", handlers.PostPutDeleteMetricID)
			})
		})
	})

	r.Route("/access-tokens", func(r chi.Router) {
		r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember)
		r.Post("/:id/level", handlers.PostAccessTokensLevel)
		r.Post("/:id/enabled", handlers.PostAccessTokensEnabled)
		r.Post("/:id/delete", handlers.PostAccessTokensDelete)
	})

	r.Route("/api", func(r chi.Router) {
		r.Route("/hosts", func(r chi.Router) {
			r.Use(middlewares.MustLoginApi)
			r.Get("/", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiHosts).(http.HandlerFunc))
			r.Post("/", handlers.PostApiHosts)
		})

		r.Route("/graphs", func(r chi.Router) {
			r.Use(middlewares.MustLoginApi)
			r.Put("/:id/metrics", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.PutApiGraphsIDMetrics).(http.HandlerFunc))
		})

		r.Route("/metrics", func(r chi.Router) {
			r.Route("/streams", func(r chi.Router) {
				r.Use(middlewares.MustLoginApiStream)
				r.Handle("/", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.ApiMetricStreams))
			})

			r.Route("/:id/streams", func(r chi.Router) {
				r.Use(middlewares.MustLoginApiStream)
				r.Handle("/", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.ApiMetricIDStreams))
			})

			r.Route("/:id/hosts/:host/streams", func(r chi.Router) {
				r.Use(middlewares.MustLoginApiStream)
				r.Handle("/", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.ApiMetricIDStreams))
			})

			r.Route("/:id", func(r chi.Router) {
				r.Use(middlewares.MustLoginApi)
				r.Get("/", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiTSMetrics).(http.HandlerFunc))
				r.Get("/15min", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiTSMetrics15Min).(http.HandlerFunc))

				r.Get("/hosts/:host", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiTSMetricsByHost).(http.HandlerFunc))
				r.Get("/hosts/:host/15min", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiTSMetricsByHost15Min).(http.HandlerFunc))
			})
		})

		r.Route("/events", func(r chi.Router) {
			r.Use(middlewares.MustLoginApi)
			r.Post("/", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.PostApiEvents).(http.HandlerFunc))
			r.Post("/line", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiEventsLine).(http.HandlerFunc))
			r.Post("/band", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiEventsBand).(http.HandlerFunc))
			r.Delete("/:id", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.DeleteApiEventsID).(http.HandlerFunc))
		})

		r.Route("/logs", func(r chi.Router) {
			r.Use(middlewares.MustLoginApi)
			r.Get("/", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiLogs).(http.HandlerFunc))
			r.Post("/", handlers.PostApiLogs)
			r.Get("/executors", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiLogsExecutors).(http.HandlerFunc))
		})

		r.Route("/executors", func(r chi.Router) {
			r.Use(middlewares.MustLoginApi)
			r.Post("/", handlers.PostApiExecutors)
		})

		r.Route("/checks", func(r chi.Router) {
			r.Use(middlewares.MustLoginApi)
			r.Get("/:id/results", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiCheckIDResults).(http.HandlerFunc))
		})

		r.Route("/metadata", func(r chi.Router) {
			r.Use(middlewares.MustLoginApi)
			r.Get("/", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiMetadata).(http.HandlerFunc))
			r.Get("/:key", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiMetadataKey).(http.HandlerFunc))
			r.Post("/:key", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.PostApiMetadataKey).(http.HandlerFunc))
			r.Delete("/:key", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.DeleteApiMetadataKey).(http.HandlerFunc))
		})
	})

	// Path to /static files
	workDir, _ := os.Getwd()
	r.FileServer("/static", http.Dir(filepath.Join(workDir, "static")))

	return r
}

// NOTE: Do not delete this until every handler is converted!
// mux returns an instance of HTTP router with all the predefined rules.
func (app *Application) mux() *mux.Router {
	SetAccessTokens := middlewares.SetAccessTokens
	MinimumMiddlewareChain := middlewares.MinimumMiddlewareChain
	MinimumAPIMiddlewareChain := middlewares.MinimumAPIMiddlewareChain
	MinimumAPIStreamMiddlewareChain := middlewares.MinimumAPIStreamMiddlewareChain

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
	router.Handle("/access-tokens/{id:[0-9]+}/delete", MinimumMiddlewareChain(CSRF).ThenFunc(handlers.PostAccessTokensDelete)).Methods("POST")

	router.Handle("/api/hosts", MinimumAPIMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiHosts))).Methods("GET")
	router.Handle("/api/hosts", MinimumAPIMiddlewareChain().ThenFunc(handlers.PostApiHosts)).Methods("POST")

	router.Handle("/api/graphs/{id:[0-9]+}/metrics", MinimumAPIMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.PutApiGraphsIDMetrics))).Methods("PUT")

	router.Handle("/api/metrics/streams", MinimumAPIStreamMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.ApiMetricStreams)))
	router.Handle("/api/metrics/{id:[0-9]+}/streams", MinimumAPIStreamMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.ApiMetricIDStreams)))
	router.Handle("/api/metrics/{id:[0-9]+}/hosts/{host}/streams", MinimumAPIStreamMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.ApiMetricIDStreams)))

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

	router.Handle("/api/checks/{id:[0-9]+}/results", MinimumAPIMiddlewareChain().ThenFunc(handlers.GetApiCheckIDResults)).Methods("GET")

	router.Handle("/api/metadata", MinimumAPIMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiMetadata))).Methods("GET")
	router.Handle(`/api/metadata/{key}`, MinimumAPIMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.PostApiMetadataKey))).Methods("POST")
	router.Handle(`/api/metadata/{key}`, MinimumAPIMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.DeleteApiMetadataKey))).Methods("DELETE")
	router.Handle(`/api/metadata/{key}`, MinimumAPIMiddlewareChain().Then(tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiMetadataKey))).Methods("GET")

	// Path of static files must be last!
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("static")))

	return router
}

// MiddlewareStruct configures all the middlewares that are in-use for all request handlers.
func (app *Application) MiddlewareStruct() (*interpose.Middleware, error) {
	middle := interpose.New()
	middle.Use(middlewares.SetAddr(app.GeneralConfig.Addr))
	middle.Use(middlewares.SetVIPAddr(app.GeneralConfig.VIPAddr))
	middle.Use(middlewares.SetVIPProtocol(app.GeneralConfig.VIPProtocol))
	middle.Use(middlewares.SetDBs(app.DBConfig))
	middle.Use(middlewares.SetCookieStore(app.cookieStore))
	middle.Use(middlewares.SetMailers(app.Mailers))
	middle.Use(middlewares.SetMessageBus(app.MessageBus))

	middle.UseHandler(app.mux())

	return middle, nil
}

// NewHTTPServer returns an instance of HTTP server.
func (app *Application) NewHTTPServer() (*graceful.Server, error) {
	// Create HTTP middlewares
	// middle, err := app.MiddlewareStruct()
	// if err != nil {
	// 	return nil, err
	// }

	requestTimeout, err := time.ParseDuration(app.GeneralConfig.RequestShutdownTimeout)
	if err != nil {
		return nil, err
	}

	// Create HTTP server
	srv := &graceful.Server{
		Timeout: requestTimeout,
		Server:  &http.Server{Addr: app.GeneralConfig.Addr, Handler: app.mux2()},
	}

	return srv, nil
}
