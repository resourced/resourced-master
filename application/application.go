package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/carbocation/interpose"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/justinas/alice"
	_ "github.com/lib/pq"
	"github.com/marcw/pagerduty"
	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/handlers"
	"github.com/resourced/resourced-master/libstring"
	"github.com/resourced/resourced-master/libtime"
	"github.com/resourced/resourced-master/mailer"
	"github.com/resourced/resourced-master/middlewares"
)

// New is the constructor for Application struct.
func New(configDir string) (*Application, error) {
	generalConfig, err := config.NewGeneralConfig(configDir)
	if err != nil {
		return nil, err
	}

	dbConfig, err := config.NewDBConfig(generalConfig)
	if err != nil {
		return nil, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	app := &Application{}
	app.Hostname = hostname
	app.GeneralConfig = generalConfig
	app.DBConfig = dbConfig
	app.cookieStore = sessions.NewCookieStore([]byte(app.GeneralConfig.CookieSecret))
	app.Mailers = make(map[string]*mailer.Mailer)

	if app.GeneralConfig.Email != nil {
		mailer, err := mailer.New(app.GeneralConfig.Email)
		if err != nil {
			return nil, err
		}
		app.Mailers["GeneralConfig"] = mailer
	}

	if app.GeneralConfig.Watchers.Email != nil {
		mailer, err := mailer.New(app.GeneralConfig.Watchers.Email)
		if err != nil {
			return nil, err
		}
		app.Mailers["GeneralConfig.Watchers"] = mailer
	}

	return app, err
}

// Application is the application object that runs HTTP server.
type Application struct {
	Hostname      string
	GeneralConfig config.GeneralConfig
	DBConfig      *config.DBConfig
	cookieStore   *sessions.CookieStore
	Mailers       map[string]*mailer.Mailer
}

func (app *Application) MiddlewareStruct() (*interpose.Middleware, error) {
	middle := interpose.New()
	middle.Use(middlewares.SetAddr(app.GeneralConfig.Addr))
	middle.Use(middlewares.SetVIPAddr(app.GeneralConfig.VIPAddr))
	middle.Use(middlewares.SetVIPProtocol(app.GeneralConfig.VIPProtocol))
	middle.Use(middlewares.SetDBs(app.DBConfig))
	middle.Use(middlewares.SetCookieStore(app.cookieStore))
	middle.Use(middlewares.SetMailers(app.Mailers))

	middle.UseHandler(app.mux())

	return middle, nil
}

func (app *Application) mux() *mux.Router {
	MustLogin := middlewares.MustLogin
	MustLoginApi := middlewares.MustLoginApi
	SetClusters := middlewares.SetClusters

	CSRFOptions := csrf.Secure(false)
	if app.GeneralConfig.HTTPS.CertFile != "" {
		CSRFOptions = csrf.Secure(true)
	}
	CSRF := csrf.Protect([]byte(app.GeneralConfig.CookieSecret), CSRFOptions)

	router := mux.NewRouter()

	router.HandleFunc("/signup", handlers.GetSignup).Methods("GET")
	router.HandleFunc("/signup", handlers.PostSignup).Methods("POST")
	router.HandleFunc("/login", handlers.GetLogin).Methods("GET")
	router.HandleFunc("/login", handlers.PostLogin).Methods("POST")

	router.Handle("/", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.GetHosts)).Methods("GET")

	router.Handle("/saved-queries", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.PostSavedQueries)).Methods("POST")
	router.Handle("/saved-queries/{id:[0-9]+}", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.PostPutDeleteSavedQueriesID)).Methods("POST", "PUT", "DELETE")

	router.Handle("/graphs", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.GetGraphs)).Methods("GET")
	router.Handle("/graphs", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.PostGraphs)).Methods("POST")
	router.Handle("/graphs/{id:[0-9]+}", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.GetPostPutDeleteGraphsID)).Methods("GET", "POST", "PUT", "DELETE")

	router.Handle("/watchers", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.GetWatchers)).Methods("GET")
	router.Handle("/watchers", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.PostWatchers)).Methods("POST")
	router.Handle("/watchers/{id:[0-9]+}", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.PostPutDeleteWatcherID)).Methods("POST", "PUT", "DELETE")
	router.Handle("/watchers/{id:[0-9]+}/silence", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.PostWatcherIDSilence)).Methods("POST")

	router.Handle("/watchers/active", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.GetWatchersActive)).Methods("GET")
	router.Handle("/watchers/active", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.PostWatchersActive)).Methods("POST")
	router.Handle("/watchers/active/{id:[0-9]+}", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.PostPutDeleteWatcherActiveID)).Methods("POST", "PUT", "DELETE")

	router.Handle("/watchers/{watcherid:[0-9]+}/triggers", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.PostWatchersTriggers)).Methods("POST")
	router.Handle("/watchers/{watcherid:[0-9]+}/triggers/{id:[0-9]+}", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.PostPutDeleteWatcherTriggerID)).Methods("POST", "PUT", "DELETE")

	router.Handle("/users/{id:[0-9]+}", alice.New(CSRF, MustLogin).ThenFunc(handlers.PostPutDeleteUsersID)).Methods("POST", "PUT", "DELETE")

	router.HandleFunc("/users/email-verification/{token}", handlers.GetUsersEmailVerificationToken).Methods("GET")

	router.Handle("/clusters", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.GetClusters)).Methods("GET")
	router.Handle("/clusters", alice.New(CSRF, MustLogin).ThenFunc(handlers.PostClusters)).Methods("POST")
	router.Handle("/clusters/{id:[0-9]+}", alice.New(CSRF, MustLogin, SetClusters).ThenFunc(handlers.PostPutDeleteClusterID)).Methods("POST", "PUT", "DELETE")

	router.Handle("/clusters/current", alice.New(MustLogin, SetClusters).ThenFunc(handlers.PostClustersCurrent)).Methods("POST")
	router.Handle("/clusters/{id:[0-9]+}/access-tokens", alice.New(CSRF, MustLogin).ThenFunc(handlers.PostAccessTokens)).Methods("POST")

	router.Handle("/clusters/{clusterid:[0-9]+}/metrics", alice.New(CSRF, MustLogin).ThenFunc(handlers.PostMetrics)).Methods("POST")
	router.Handle("/clusters/{clusterid:[0-9]+}/metrics/{id:[0-9]+}", alice.New(CSRF, MustLogin).ThenFunc(handlers.PostPutDeleteMetricID)).Methods("POST", "PUT", "DELETE")

	router.Handle("/access-tokens/{id:[0-9]+}/level", alice.New(CSRF, MustLogin).ThenFunc(handlers.PostAccessTokensLevel)).Methods("POST")
	router.Handle("/access-tokens/{id:[0-9]+}/enabled", alice.New(CSRF, MustLogin).ThenFunc(handlers.PostAccessTokensEnabled)).Methods("POST")

	router.Handle("/api/hosts", alice.New(MustLoginApi).ThenFunc(handlers.GetApiHosts)).Methods("GET")
	router.Handle("/api/hosts", alice.New(MustLoginApi).ThenFunc(handlers.PostApiHosts)).Methods("POST")

	router.Handle("/api/metrics/{id:[0-9]+}/hosts/{host}", alice.New(MustLoginApi).ThenFunc(handlers.GetApiTSMetricsByHost)).Methods("GET")
	router.Handle("/api/metrics/{id:[0-9]+}/hosts/{host}/15min", alice.New(MustLoginApi).ThenFunc(handlers.GetApiTSMetricsByHost15Min)).Methods("GET")

	router.Handle("/api/metrics/{id:[0-9]+}", alice.New(MustLoginApi).ThenFunc(handlers.GetApiTSMetrics)).Methods("GET")
	router.Handle("/api/metrics/{id:[0-9]+}/15min", alice.New(MustLoginApi).ThenFunc(handlers.GetApiTSMetrics15Min)).Methods("GET")

	router.Handle(`/api/events`, alice.New(MustLoginApi).ThenFunc(handlers.PostApiEvents)).Methods("POST")
	router.Handle(`/api/events/{id:[0-9]+}`, alice.New(MustLoginApi).ThenFunc(handlers.DeleteApiEventsID)).Methods("DELETE")
	router.Handle(`/api/events/line`, alice.New(MustLoginApi).ThenFunc(handlers.GetApiEventsLine)).Methods("GET")
	router.Handle(`/api/events/band`, alice.New(MustLoginApi).ThenFunc(handlers.GetApiEventsBand)).Methods("GET")

	router.Handle(`/api/executors`, alice.New(MustLoginApi).ThenFunc(handlers.PostApiExecutors)).Methods("POST")
	router.Handle(`/api/logs`, alice.New(MustLoginApi).ThenFunc(handlers.PostApiLogs)).Methods("POST")

	router.Handle("/api/metadata", alice.New(MustLoginApi).ThenFunc(handlers.GetApiMetadata)).Methods("GET")
	router.Handle(`/api/metadata/{key}`, alice.New(MustLoginApi).ThenFunc(handlers.PostApiMetadataKey)).Methods("POST")
	router.Handle(`/api/metadata/{key}`, alice.New(MustLoginApi).ThenFunc(handlers.DeleteApiMetadataKey)).Methods("DELETE")
	router.Handle(`/api/metadata/{key}`, alice.New(MustLoginApi).ThenFunc(handlers.GetApiMetadataKey)).Methods("GET")

	// Path of static files must be last!
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("static")))

	return router
}

func (app *Application) PruneAll() {
	for {
		app.PruneTSWatchersOnce()
		app.PruneTSMetricsOnce()
		app.PruneTSEventsOnce()
		app.PruneTSExecutorLogsOnce()
		app.PruneTSLogsOnce()

		libtime.SleepString("24h")
	}
}

func (app *Application) PruneTSWatchersOnce() {
	for _, db := range app.DBConfig.TSWatchers.DBs {
		go func() {
			err := dal.NewTSWatcher(db).DeleteByDayInterval(nil, app.GeneralConfig.Watchers.DataRetention)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"Method":        "TSWatcher.DeleteByDayInterval",
					"DataRetention": app.GeneralConfig.Watchers.DataRetention,
				}).Error(err)
			}
		}()
	}
}

func (app *Application) PruneTSMetricsOnce() {
	for _, db := range app.DBConfig.TSMetrics.DBs {
		go func() {
			err := dal.NewTSMetric(db).DeleteByDayInterval(nil, app.GeneralConfig.Metrics.DataRetention)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"Method":        "TSMetric.DeleteByDayInterval",
					"DataRetention": app.GeneralConfig.Metrics.DataRetention,
				}).Error(err)
			}
		}()
	}
}

func (app *Application) PruneTSEventsOnce() {
	for _, db := range app.DBConfig.TSEvents.DBs {
		go func() {
			err := dal.NewTSEvent(db).DeleteByDayInterval(nil, app.GeneralConfig.Events.DataRetention)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"Method":        "TSEvent.DeleteByDayInterval",
					"DataRetention": app.GeneralConfig.Events.DataRetention,
				}).Error(err)
			}
		}()
	}
}

func (app *Application) PruneTSExecutorLogsOnce() {
	for _, db := range app.DBConfig.TSExecutorLogs.DBs {
		go func() {
			err := dal.NewTSExecutorLog(db).DeleteByDayInterval(nil, app.GeneralConfig.ExecutorLogs.DataRetention)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"Method":        "TSExecutorLog.DeleteByDayInterval",
					"DataRetention": app.GeneralConfig.ExecutorLogs.DataRetention,
				}).Error(err)
			}
		}()
	}
}

func (app *Application) PruneTSLogsOnce() {
	for _, db := range app.DBConfig.TSLogs.DBs {
		go func() {
			err := dal.NewTSLog(db).DeleteByDayInterval(nil, app.GeneralConfig.Logs.DataRetention)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"Method":        "TSLog.DeleteByDayInterval",
					"DataRetention": app.GeneralConfig.Logs.DataRetention,
				}).Error(err)
			}
		}()
	}
}

func (app *Application) WatchAll() {
	watcherRowsChan := make(chan []*dal.WatcherRow)

	// Fetch daemons and watchers data every 5 minutes
	go func() {
		for {
			daemonHostnames, _ := dal.NewDaemon(app.DBConfig.Core).AllHostnames(nil)
			groupedWatcherRows, _ := dal.NewWatcher(app.DBConfig.Core).AllSplitToDaemons(nil, daemonHostnames)
			watcherRowsChan <- groupedWatcherRows[app.Hostname]

			libtime.SleepString(app.GeneralConfig.Watchers.ListFetchInterval)
		}
	}()

	go func() {
		for {
			select {
			case watcherRows := <-watcherRowsChan:
				for _, watcherRow := range watcherRows {
					go func(watcherRow *dal.WatcherRow) {
						if watcherRow.IsPassive() {
							// Passive watching and triggering
							err := app.PassiveWatchOnce(watcherRow.ClusterID, watcherRow)
							if err != nil {
								logrus.WithFields(logrus.Fields{
									"Method":    "Application.PassiveWatchOnce",
									"ClusterID": watcherRow.ClusterID,
									"WatcherID": watcherRow.ID,
								}).Error(err)
							}

							err = app.RunTrigger(watcherRow.ClusterID, watcherRow)
							if err != nil {
								logrus.WithFields(logrus.Fields{
									"Method":    "Application.RunTrigger",
									"ClusterID": watcherRow.ClusterID,
									"WatcherID": watcherRow.ID,
								}).Error(err)
							}
						} else {
							// Active watching and triggering
							err := app.ActiveWatchOnce(watcherRow.ClusterID, watcherRow)
							if err != nil {
								logrus.WithFields(logrus.Fields{
									"Method":    "Application.ActiveWatchOnce",
									"ClusterID": watcherRow.ClusterID,
									"WatcherID": watcherRow.ID,
								}).Error(err)
							}
						}

						libtime.SleepString(watcherRow.CheckInterval)
					}(watcherRow)
				}
			}
		}
	}()
}

func (app *Application) PassiveWatchOnce(clusterID int64, watcherRow *dal.WatcherRow) error {
	affectedHosts, err := dal.NewHost(app.DBConfig.Core).AllByClusterIDQueryAndUpdatedInterval(nil, clusterID, watcherRow.SavedQuery.String, watcherRow.HostsLastUpdatedForPostgres())
	if err != nil {
		return err
	}

	numAffectedHosts := int64(len(affectedHosts))

	if numAffectedHosts == 0 || numAffectedHosts >= watcherRow.LowAffectedHosts {
		tsWatcherDataHosts := make([]string, numAffectedHosts)
		for i, affectedHost := range affectedHosts {
			tsWatcherDataHosts[i] = affectedHost.Name
		}

		tsWatcherData := make(map[string]interface{})
		tsWatcherData["hosts"] = tsWatcherDataHosts

		jsonData, err := json.Marshal(tsWatcherData)
		if err != nil {
			return err
		}

		// Write to ts_watchers asynchronously
		dbs := app.DBConfig.TSWatchers.PickMultipleForWrites()
		for _, db := range dbs {
			go func() {
				err := dal.NewTSWatcher(db).Create(nil, clusterID, watcherRow.ID, numAffectedHosts, jsonData)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"Method":          "TSWatcher.Create",
						"ClusterID":       clusterID,
						"WatcherID":       watcherRow.ID,
						"NumAffectedHost": numAffectedHosts,
					}).Error(err)
				}
			}()
		}
	}

	return err
}

func (app *Application) ActiveWatchOnce(clusterID int64, watcherRow *dal.WatcherRow) error {
	// Don't do anything if watcher is silenced.
	if watcherRow.IsSilenced {
		return nil
	}

	hostnames := dal.NewWatcher(app.DBConfig.Core).HostsToPerformActiveChecks(clusterID, watcherRow)

	// If there are no hostnames, then we are not checking anything
	if len(hostnames) == 0 {
		return nil
	}

	numAffectedHosts := 0
	tsWatcherDataHosts := make([]string, 0)

	var errCollectorMutex sync.Mutex

	for _, hostname := range hostnames {
		if watcherRow.Command() == "ping" {
			go func(hostname string) {
				outBytes, err := watcherRow.PerformActiveCheckPing(hostname)
				if err != nil {
					errCollectorMutex.Lock()
					numAffectedHosts = numAffectedHosts + 1
					tsWatcherDataHosts = append(tsWatcherDataHosts, hostname)
					errCollectorMutex.Unlock()

					logrus.WithFields(logrus.Fields{
						"Method":   "watcherRow.PerformActiveCheckPing",
						"Hostname": hostname,
						"Output":   string(outBytes),
					}).Error(err)
				}
			}(hostname)

		} else if watcherRow.Command() == "ssh" {
			go func(hostname string) {
				outBytes, err := watcherRow.PerformActiveCheckSSH(hostname)
				outString := string(outBytes)

				// We only care about SSH connectivity
				if err != nil && !strings.Contains(outString, "Permission denied") && !strings.Contains(outString, "Host key verification failed") {
					errCollectorMutex.Lock()
					numAffectedHosts = numAffectedHosts + 1
					tsWatcherDataHosts = append(tsWatcherDataHosts, hostname)
					errCollectorMutex.Unlock()

					logrus.WithFields(logrus.Fields{
						"Method":   "watcherRow.PerformActiveCheckSSH",
						"Hostname": hostname,
						"Output":   outString,
					}).Error(err)
				}
			}(hostname)

		} else if watcherRow.Command() == "http" {
			go func(hostname string) {
				resp, err := watcherRow.PerformActiveCheckHTTP(hostname)

				if err != nil || (resp != nil && resp.StatusCode != watcherRow.HTTPCode()) {
					errCollectorMutex.Lock()
					numAffectedHosts = numAffectedHosts + 1
					tsWatcherDataHosts = append(tsWatcherDataHosts, hostname)
					errCollectorMutex.Unlock()

					logrus.WithFields(logrus.Fields{
						"Method":             "watcherRow.PerformActiveCheckHTTP",
						"HTTPMethod":         watcherRow.HTTPMethod(),
						"StatusCode":         resp.StatusCode,
						"ExpectedStatusCode": watcherRow.HTTPCode(),
					}).Error(err)
				}
			}(hostname)
		}
	}

	if numAffectedHosts == 0 || int64(numAffectedHosts) >= watcherRow.LowAffectedHosts {
		tsWatcherData := make(map[string]interface{})
		tsWatcherData["hosts"] = tsWatcherDataHosts

		jsonData, err := json.Marshal(tsWatcherData)
		if err != nil {
			return err
		}

		// Write to ts_watchers asynchronously
		dbs := app.DBConfig.TSWatchers.PickMultipleForWrites()
		for _, db := range dbs {
			go func() {
				err := dal.NewTSWatcher(db).Create(nil, clusterID, watcherRow.ID, int64(numAffectedHosts), jsonData)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"Method":          "TSWatcher.Create",
						"ClusterID":       clusterID,
						"WatcherID":       watcherRow.ID,
						"NumAffectedHost": numAffectedHosts,
					}).Error(err)
				}
			}()
		}
	}

	return nil
}

func (app *Application) RunTrigger(clusterID int64, watcherRow *dal.WatcherRow) error {
	// Don't do anything if watcher is silenced.
	if watcherRow.IsSilenced {
		return nil
	}

	tsWatcherDB := app.DBConfig.TSWatchers.PickRandom()

	triggerRows, err := dal.NewWatcherTrigger(app.DBConfig.Core).AllByClusterIDAndWatcherID(nil, clusterID, watcherRow.ID)
	if err != nil {
		return err
	}

	for _, triggerRow := range triggerRows {
		tsWatchers, err := dal.NewTSWatcher(tsWatcherDB).AllViolationsByClusterIDWatcherIDAndInterval(nil, clusterID, watcherRow.ID, watcherRow.LowAffectedHosts, triggerRow.CreatedInterval)
		if err != nil {
			return err
		}
		if len(tsWatchers) == 0 {
			continue
		}

		lastViolation := tsWatchers[0]
		violationsCount := len(tsWatchers)

		if int64(violationsCount) >= triggerRow.LowViolationsCount && int64(violationsCount) <= triggerRow.HighViolationsCount {
			if triggerRow.ActionTransport() == "nothing" {
				// Do nothing

			} else if triggerRow.ActionTransport() == "email" {
				if triggerRow.ActionEmail() == "" {
					continue
				}

				to := triggerRow.ActionEmail()
				subject := fmt.Sprintf(`Watcher(ID: %v): %v, Query: %v`, watcherRow.ID, watcherRow.Name, watcherRow.SavedQuery.String)
				body := ""

				if lastViolation != nil {
					bodyBytes, err := libstring.PrettyPrintJSON([]byte(lastViolation.Data.String()))
					if err != nil {
						continue
					}

					body = string(bodyBytes)
				}

				mailr, ok := app.Mailers["GeneralConfig.Watchers"]
				if !ok {
					continue
				}

				err = mailr.Send(to, subject, body)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"Method":          "mailer.Send",
						"ActionTransport": triggerRow.ActionTransport(),
						"HostAndPort":     mailr.HostAndPort,
						"From":            mailr.From,
						"To":              to,
						"Subject":         subject,
					}).Error(err)
					continue
				}

			} else if triggerRow.ActionTransport() == "sms" {
				carrier := strings.ToLower(triggerRow.ActionSMSCarrier())

				gateway, ok := app.GeneralConfig.Watchers.SMSEmailGateway[carrier]
				if !ok {
					logrus.Warningf("Unable to lookup SMS Gateway for carrier: %v", carrier)
					continue
				}

				flattenPhone := libstring.FlattenPhone(triggerRow.ActionSMSPhone())
				if len(flattenPhone) != 10 {
					logrus.Warningf("Length of phone number is not 10. Flatten phone number: %v. Length: %v", flattenPhone, len(flattenPhone))
					continue
				}

				to := fmt.Sprintf("%v@%v", flattenPhone, gateway)
				subject := ""
				body := fmt.Sprintf(`%v Watcher(ID: %v): %v, Query: %v, failed %v times`, app.GeneralConfig.Watchers.Email.SubjectPrefix, watcherRow.ID, watcherRow.Name, watcherRow.SavedQuery, violationsCount)

				mailr, ok := app.Mailers["GeneralConfig.Watchers"]
				if !ok {
					continue
				}

				err = mailr.Send(to, subject, body)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"Method":          "smtp.SendMail",
						"ActionTransport": triggerRow.ActionTransport(),
						"HostAndPort":     mailr.HostAndPort,
						"From":            mailr.From,
						"To":              to,
						"Subject":         subject,
					}).Error(err)
					continue
				}

			} else if triggerRow.ActionTransport() == "pagerduty" {
				err = app.RunTriggerPagerDuty(triggerRow, lastViolation)
				if err != nil {
					logrus.Error(err)
					continue
				}
			}
		}
	}

	return nil
}

func (app *Application) RunTriggerPagerDuty(triggerRow *dal.WatcherTriggerRow, lastViolation *dal.TSWatcherRow) (err error) {
	// Create a new PD "trigger" event
	event := pagerduty.NewTriggerEvent(triggerRow.ActionPagerDutyServiceKey(), triggerRow.ActionPagerDutyDescription())

	// Add details to PD event
	if lastViolation != nil {
		err = lastViolation.Data.Unmarshal(&event.Details)
		if err != nil {
			return err
		}
	}

	// Add Client to PD event
	event.Client = fmt.Sprintf("ResourceD Master on: %v", app.Hostname)

	// Submit PD event
	pdResponse, _, err := pagerduty.Submit(event)
	if err != nil {
		return err
	}
	if pdResponse == nil {
		return nil
	}

	// Update incident key into watchers_triggers row
	wt := dal.NewWatcherTrigger(app.DBConfig.Core)

	triggerUpdateActionParams := wt.ActionParamsByExistingTrigger(triggerRow)
	triggerUpdateActionParams["PagerDutyIncidentKey"] = pdResponse.IncidentKey

	triggerUpdateActionJSON, err := json.Marshal(triggerUpdateActionParams)
	if err != nil {
		return err
	}

	triggerUpdateParams := wt.CreateOrUpdateParameters(triggerRow.ClusterID, triggerRow.WatcherID, triggerRow.LowViolationsCount, triggerRow.HighViolationsCount, triggerRow.CreatedInterval, triggerUpdateActionJSON)

	_, err = wt.UpdateFromTable(nil, triggerUpdateParams, fmt.Sprintf("id=%v", triggerRow.ID))
	if err != nil {
		return err
	}

	return err
}
