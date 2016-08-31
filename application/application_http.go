package application

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/didip/stopwatch"
	"github.com/didip/tollbooth"
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

func (app *Application) mux() *chi.Mux {
	generalAPILimiter := tollbooth.NewLimiter(int64(app.GeneralConfig.RateLimiters.GeneralAPI), time.Second)
	signupLimiter := tollbooth.NewLimiter(int64(app.GeneralConfig.RateLimiters.PostSignup), time.Second)

	useHTTPS := app.GeneralConfig.HTTPS.CertFile != "" && app.GeneralConfig.HTTPS.KeyFile != ""
	CSRF := middlewares.CSRFMiddleware(useHTTPS, app.GeneralConfig.CookieSecret)

	r := chi.NewRouter()

	// Set middlewares which impact every request.
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
		r.Post("/", handlers.PostSavedQueries)

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
			r.Get("/line", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiEventsLine).(http.HandlerFunc))
			r.Get("/band", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiEventsBand).(http.HandlerFunc))
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

// NewHTTPServer returns an instance of HTTP server.
func (app *Application) NewHTTPServer() (*graceful.Server, error) {
	requestTimeout, err := time.ParseDuration(app.GeneralConfig.RequestShutdownTimeout)
	if err != nil {
		return nil, err
	}

	// Create HTTP server
	srv := &graceful.Server{
		Timeout: requestTimeout,
		Server:  &http.Server{Addr: app.GeneralConfig.Addr, Handler: app.mux()},
	}

	return srv, nil
}
