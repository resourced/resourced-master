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

// NewHandlerInstruments creates channels for recording latencies.
func (app *Application) NewHandlerInstruments() map[string]chan int64 {
	instruments := make(map[string]chan int64)
	for _, key := range []string{"GetLogin", "GetHosts", "GetGraphs", "GetGraphsID", "GetLogs", "GetChecks", "GetApiLogs"} {
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

// Mux routes HTTP requests to their appropriate handlers
func (app *Application) Mux() *chi.Mux {
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
	r.Use(middlewares.SetLogger("outLogger", app.OutLogger))
	r.Use(middlewares.SetLogger("errLogger", app.ErrLogger))

	if !app.GeneralConfig.JustAPI {
		r.Get("/signup", handlers.GetSignup)
		r.Post("/signup", tollbooth.LimitFuncHandler(signupLimiter, handlers.PostSignup).(http.HandlerFunc))

		r.Get("/login", stopwatch.LatencyFuncHandler(app.getHandlerInstrument("GetLogin"), []string{"GET"}, handlers.GetLogin).(http.HandlerFunc))
		r.Post("/login", handlers.PostLogin)

		r.Route("/", func(r chi.Router) {
			r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember, middlewares.SetAccessTokens)
			r.Get("/", stopwatch.LatencyFuncHandler(app.getHandlerInstrument("GetHosts"), []string{"GET"}, handlers.GetHosts).(http.HandlerFunc))
		})

		r.Route("/hosts", func(r chi.Router) {
			r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember, middlewares.SetAccessTokens)
			r.Get("/", stopwatch.LatencyFuncHandler(app.getHandlerInstrument("GetHosts"), []string{"GET"}, handlers.GetHosts).(http.HandlerFunc))

			r.Route("/:id", func(r chi.Router) {
				r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember, middlewares.SetAccessTokens)
				r.Get("/", handlers.GetHostsID)

				r.Post("/master-tags", handlers.PostHostsIDMasterTags)
			})
		})

		r.Route("/saved-queries", func(r chi.Router) {
			r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember)
			r.Post("/", handlers.PostSavedQueries)

			r.Route("/:id", func(r chi.Router) {
				r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember)
				r.Post("/", handlers.PostPutDeleteSavedQueriesID)
				r.Delete("/", handlers.PostPutDeleteSavedQueriesID)
			})
		})

		r.Route("/graphs", func(r chi.Router) {
			r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember, middlewares.SetAccessTokens)
			r.Get("/", stopwatch.LatencyFuncHandler(app.getHandlerInstrument("GetGraphs"), []string{"GET"}, handlers.GetGraphs).(http.HandlerFunc))
			r.Post("/", handlers.PostGraphs)

			r.Route("/:id", func(r chi.Router) {
				r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember, middlewares.SetAccessTokens)
				r.Get("/", stopwatch.LatencyFuncHandler(app.getHandlerInstrument("GetGraphsID"), []string{"GET"}, handlers.GetPostPutDeleteGraphsID).(http.HandlerFunc))
				r.Post("/", handlers.GetPostPutDeleteGraphsID)
				r.Put("/", handlers.GetPostPutDeleteGraphsID)
				r.Delete("/", handlers.GetPostPutDeleteGraphsID)
			})
		})

		r.Route("/logs", func(r chi.Router) {
			r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember, middlewares.SetAccessTokens)
			r.Get("/", stopwatch.LatencyFuncHandler(app.getHandlerInstrument("GetLogs"), []string{"GET"}, handlers.GetLogs).(http.HandlerFunc))
		})

		r.Route("/checks", func(r chi.Router) {
			r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember, middlewares.SetAccessTokens)
			r.Get("/", stopwatch.LatencyFuncHandler(app.getHandlerInstrument("GetChecks"), []string{"GET"}, handlers.GetChecks).(http.HandlerFunc))
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

					r.Route("/:triggerID", func(r chi.Router) {
						r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember)
						r.Post("/", handlers.PostPutDeleteCheckTriggerID)
						r.Put("/", handlers.PostPutDeleteCheckTriggerID)
						r.Delete("/", handlers.PostPutDeleteCheckTriggerID)
					})
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

			r.Route("/:clusterID", func(r chi.Router) {
				r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember)
				r.Post("/", handlers.PostPutDeleteClusterID)
				r.Put("/", handlers.PostPutDeleteClusterID)
				r.Delete("/", handlers.PostPutDeleteClusterID)

				r.Route("/current", func(r chi.Router) {
					r.Post("/", handlers.PostClusterIDCurrent)
				})

				r.Post("/access-tokens", handlers.PostAccessTokens)
				r.Post("/users", handlers.PostPutDeleteClusterIDUsers)
				r.Put("/users", handlers.PostPutDeleteClusterIDUsers)
				r.Delete("/users", handlers.PostPutDeleteClusterIDUsers)

				r.Route("/metrics", func(r chi.Router) {
					r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember)
					r.Post("/", handlers.PostMetrics)

					r.Route("/:metricID", func(r chi.Router) {
						r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember)
						r.Post("/", handlers.PostPutDeleteMetricID)
						r.Put("/", handlers.PostPutDeleteMetricID)
						r.Delete("/", handlers.PostPutDeleteMetricID)
					})
				})
			})
		})

		r.Route("/access-tokens/:id", func(r chi.Router) {
			r.Use(CSRF, middlewares.MustLogin, middlewares.SetClusters, middlewares.MustBeMember)
			r.Post("/level", handlers.PostAccessTokensLevel)
			r.Post("/enabled", handlers.PostAccessTokensEnabled)
			r.Post("/delete", handlers.PostAccessTokensDelete)
		})
	}

	r.Route("/api", func(r chi.Router) {
		r.Route("/hosts", func(r chi.Router) {
			r.Use(middlewares.MustLoginApi)
			r.Get("/", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiHosts).(http.HandlerFunc))
			r.Post("/", handlers.PostApiHosts)
		})

		r.Route("/graphs/:id", func(r chi.Router) {
			r.Use(middlewares.MustLoginApi)
			r.Put("/metrics", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.PutApiGraphsIDMetrics).(http.HandlerFunc))
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
			r.Get("/", tollbooth.LimitHandler(
				generalAPILimiter,
				stopwatch.LatencyFuncHandler(
					app.getHandlerInstrument("GetApiLogs"),
					[]string{"GET"}, handlers.GetApiLogs,
				),
			).(http.HandlerFunc))

			r.Post("/", handlers.PostApiLogs)
		})

		r.Route("/checks/:id", func(r chi.Router) {
			r.Use(middlewares.MustLoginApi)
			r.Get("/results", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiCheckIDResults).(http.HandlerFunc))
		})

		r.Route("/metadata", func(r chi.Router) {
			r.Use(middlewares.MustLoginApi)
			r.Get("/", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiMetadata).(http.HandlerFunc))

			r.Route("/:key", func(r chi.Router) {
				r.Use(middlewares.MustLoginApi)
				r.Get("/", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.GetApiMetadataKey).(http.HandlerFunc))
				r.Post("/", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.PostApiMetadataKey).(http.HandlerFunc))
				r.Delete("/", tollbooth.LimitFuncHandler(generalAPILimiter, handlers.DeleteApiMetadataKey).(http.HandlerFunc))
			})
		})
	})

	// Path to /static files
	if !app.GeneralConfig.JustAPI {
		workDir, _ := os.Getwd()
		r.FileServer("/static", http.Dir(filepath.Join(workDir, "static")))
	}

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
		Server:  &http.Server{Addr: app.GeneralConfig.Addr, Handler: app.Mux()},
	}

	return srv, nil
}
