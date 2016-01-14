package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/carbocation/interpose"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/justinas/alice"
	_ "github.com/lib/pq"
	"github.com/marcw/pagerduty"
	"github.com/mattes/migrate/migrate"
	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/handlers"
	"github.com/resourced/resourced-master/libsmtp"
	"github.com/resourced/resourced-master/libstring"
	"github.com/resourced/resourced-master/libtime"
	"github.com/resourced/resourced-master/middlewares"
	"github.com/resourced/resourced-master/wstrafficker"
)

// New is the constructor for Application struct.
func New(configDir string) (*Application, error) {
	generalConfig, err := config.NewGeneralConfig(configDir)
	if err != nil {
		return nil, err
	}

	db, err := sqlx.Connect("postgres", generalConfig.DSN)
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
	app.DB = db
	app.cookieStore = sessions.NewCookieStore([]byte(app.GeneralConfig.CookieSecret))
	app.csrfProtect = csrf.Protect([]byte(app.GeneralConfig.CookieSecret))
	app.WSTraffickers = wstrafficker.NewWSTraffickers()

	return app, err
}

// Application is the application object that runs HTTP server.
type Application struct {
	Hostname      string
	GeneralConfig config.GeneralConfig
	DB            *sqlx.DB
	cookieStore   *sessions.CookieStore
	csrfProtect   func(http.Handler) http.Handler
	WSTraffickers *wstrafficker.WSTraffickers
}

func (app *Application) MiddlewareStruct() (*interpose.Middleware, error) {
	middle := interpose.New()
	middle.Use(middlewares.SetAddr(app.GeneralConfig.Addr))
	middle.Use(middlewares.SetDB(app.DB))
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
	router.Handle(`/metadata/{key}`, alice.New(MustLogin, SetClusters).ThenFunc(handlers.DeleteMetadataKey)).Methods("POST", "DELETE")

	router.Handle("/watchers", alice.New(MustLogin, SetClusters).ThenFunc(handlers.GetWatchers)).Methods("GET")
	router.Handle("/watchers", alice.New(MustLogin, SetClusters).ThenFunc(handlers.PostWatchers)).Methods("POST")
	router.Handle("/watchers/{id:[0-9]+}", alice.New(MustLogin, SetClusters).ThenFunc(handlers.PostPutDeleteWatcherID)).Methods("POST", "PUT", "DELETE")

	router.Handle("/watchers/{watcherid:[0-9]+}/triggers", alice.New(MustLogin, SetClusters).ThenFunc(handlers.PostWatchersTriggers)).Methods("POST")
	router.Handle("/watchers/{watcherid:[0-9]+}/triggers/{id:[0-9]+}", alice.New(MustLogin, SetClusters).ThenFunc(handlers.PostPutDeleteWatcherTriggerID)).Methods("POST", "PUT", "DELETE")

	router.Handle("/users/{id:[0-9]+}", alice.New(MustLogin).ThenFunc(handlers.PostPutDeleteUsersID)).Methods("POST", "PUT", "DELETE")

	router.Handle("/clusters", alice.New(MustLogin, SetClusters).ThenFunc(handlers.GetClusters)).Methods("GET")
	router.Handle("/clusters", alice.New(MustLogin).ThenFunc(handlers.PostClusters)).Methods("POST")

	router.Handle("/clusters/current", alice.New(MustLogin, SetClusters).ThenFunc(handlers.PostClustersCurrent)).Methods("POST")

	router.Handle("/clusters/{id:[0-9]+}/access-tokens", alice.New(MustLogin).ThenFunc(handlers.PostAccessTokens)).Methods("POST")

	router.Handle("/clusters/{id:[0-9]+}/metrics", alice.New(MustLogin).ThenFunc(handlers.PostMetrics)).Methods("POST")

	router.Handle("/access-tokens/{id:[0-9]+}/level", alice.New(MustLogin).ThenFunc(handlers.PostAccessTokensLevel)).Methods("POST")
	router.Handle("/access-tokens/{id:[0-9]+}/enabled", alice.New(MustLogin).ThenFunc(handlers.PostAccessTokensEnabled)).Methods("POST")

	router.Handle("/saved-queries", alice.New(MustLogin).ThenFunc(handlers.PostSavedQueries)).Methods("POST")
	router.Handle("/saved-queries/{id:[0-9]+}", alice.New(MustLogin).ThenFunc(handlers.PostPutDeleteSavedQueriesID)).Methods("POST", "PUT", "DELETE")

	router.HandleFunc("/api/ws/access-tokens/{id}", handlers.ApiWSAccessToken)

	router.Handle("/api/hosts", alice.New(MustLoginApi).ThenFunc(handlers.GetApiHosts)).Methods("GET")
	router.Handle("/api/hosts", alice.New(MustLoginApi).ThenFunc(handlers.PostApiHosts)).Methods("POST")

	router.Handle("/api/metadata", alice.New(MustLoginApi).ThenFunc(handlers.GetApiMetadata)).Methods("GET")
	router.Handle(`/api/metadata/{key}`, alice.New(MustLoginApi).ThenFunc(handlers.PostApiMetadataKey)).Methods("POST")
	router.Handle(`/api/metadata/{key}`, alice.New(MustLoginApi).ThenFunc(handlers.DeleteApiMetadataKey)).Methods("DELETE")
	router.Handle(`/api/metadata/{key}`, alice.New(MustLoginApi).ThenFunc(handlers.GetApiMetadataKey)).Methods("GET")

	// Path of static files must be last!
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("static")))

	return router
}

func (app *Application) MigrateUp() (err []error, ok bool) {
	return migrate.UpSync(app.GeneralConfig.DSN, "./migrations")
}

func (app *Application) WatchAll() {
	watcherRowsChan := make(chan []*dal.WatcherRow)

	// Fetch daemons and watchers data every 5 minutes
	go func() {
		for {
			daemonHostnames, _ := dal.NewDaemon(app.DB).AllHostnames(nil)
			groupedWatcherRows, _ := dal.NewWatcher(app.DB).AllSplitToDaemons(nil, daemonHostnames)
			watcherRowsChan <- groupedWatcherRows[app.Hostname]

			libtime.SleepString(app.GeneralConfig.Watchers.ListFetchInterval)
		}
	}()

	go func() {
		for {
			select {
			case watcherRows := <-watcherRowsChan:
				for _, watcherRow := range watcherRows {
					go func() {
						app.WatchOnce(watcherRow.ClusterID, watcherRow)
						app.RunTrigger(watcherRow.ClusterID, watcherRow)

						libtime.SleepString(watcherRow.CheckInterval)
					}()
				}
			}
		}
	}()
}

func (app *Application) WatchOnce(clusterID int64, watcherRow *dal.WatcherRow) error {
	var err error

	lastUpdated := strings.Replace(watcherRow.HostsLastUpdated, " ago", "", 1)

	affectedHosts, err := dal.NewHost(app.DB).AllByClusterIDQueryAndUpdatedInterval(nil, clusterID, watcherRow.SavedQuery, lastUpdated)
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

		err = dal.NewTSWatcher(app.DB).Create(nil, clusterID, watcherRow.ID, numAffectedHosts, jsonData)
	}

	return err
}

func (app *Application) RunTrigger(clusterID int64, watcherRow *dal.WatcherRow) error {
	violationsCount, err := dal.NewTSWatcher(app.DB).CountViolationsSinceLastGreenMarker(nil)
	if err != nil {
		return err
	}

	// Don't bother doing anything else if there are no new violations.
	if violationsCount <= 0 {
		return nil
	}

	triggerRows, err := dal.NewWatcherTrigger(app.DB).AllByClusterIDAndWatcherID(nil, clusterID, watcherRow.ID)
	if err != nil {
		return err
	}

	lastViolation, err := dal.NewTSWatcher(app.DB).LastViolation(nil)
	if err != nil {
		return err
	}

	for _, triggerRow := range triggerRows {
		if int64(violationsCount) >= triggerRow.LowViolationsCount && int64(violationsCount) <= triggerRow.HighViolationsCount {
			if triggerRow.ActionTransport() == "nothing" {
				// Do nothing
			} else if triggerRow.ActionTransport() == "email" {
				if triggerRow.ActionEmail() == "" {
					continue
				}

				auth := smtp.PlainAuth(
					app.GeneralConfig.Watchers.Email.Identity,
					app.GeneralConfig.Watchers.Email.Username,
					app.GeneralConfig.Watchers.Email.Password,
					app.GeneralConfig.Watchers.Email.Host)

				hostAndPort := fmt.Sprintf("%v:%v", app.GeneralConfig.Watchers.Email.Host, app.GeneralConfig.Watchers.Email.Port)
				from := app.GeneralConfig.Watchers.Email.From
				to := triggerRow.ActionEmail()
				subject := fmt.Sprintf(`%v Watcher(ID: %v): %v, Query: %v`, app.GeneralConfig.Watchers.Email.SubjectPrefix, watcherRow.ID, watcherRow.Name, watcherRow.SavedQuery)

				bodyBytes, err := libstring.PrettyPrintJSON([]byte(lastViolation.Data.String()))
				if err != nil {
					logrus.Error(err)
					continue
				}

				body := string(bodyBytes)
				message := libsmtp.BuildMessage(from, to, subject, body)

				err = smtp.SendMail(hostAndPort, auth, from, []string{to}, []byte(message))
				if err != nil {
					logrus.Error(err)
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

				auth := smtp.PlainAuth(
					app.GeneralConfig.Watchers.Email.Identity,
					app.GeneralConfig.Watchers.Email.Username,
					app.GeneralConfig.Watchers.Email.Password,
					app.GeneralConfig.Watchers.Email.Host)

				hostAndPort := fmt.Sprintf("%v:%v", app.GeneralConfig.Watchers.Email.Host, app.GeneralConfig.Watchers.Email.Port)
				from := app.GeneralConfig.Watchers.Email.From
				to := fmt.Sprintf("%v@%v", flattenPhone, gateway)
				subject := ""
				body := fmt.Sprintf(`%v Watcher(ID: %v): %v, Query: %v, failed %v times`, app.GeneralConfig.Watchers.Email.SubjectPrefix, watcherRow.ID, watcherRow.Name, watcherRow.SavedQuery, violationsCount)

				message := libsmtp.BuildMessage(from, to, subject, body)

				err = smtp.SendMail(hostAndPort, auth, from, []string{to}, []byte(message))
				if err != nil {
					logrus.Error(err)
					continue
				}

			} else if triggerRow.ActionTransport() == "pagerduty" {
				// Create a new PD "trigger" event
				event := pagerduty.NewTriggerEvent(triggerRow.ActionPagerDutyServiceKey(), triggerRow.ActionPagerDutyDescription())

				// Add details to PD event
				err = lastViolation.Data.Unmarshal(&event.Details)
				if err != nil {
					logrus.Error(err)
					continue
				}

				// Add Client to PD event
				event.Client = fmt.Sprintf("ResourceD Master on: %v", app.Hostname)

				// Submit PD event
				pdResponse, _, err := pagerduty.Submit(event)
				if err != nil {
					logrus.Error(err)
				}
				if pdResponse == nil {
					continue
				}

				// Update incident key into watchers_triggers row
				wt := dal.NewWatcherTrigger(app.DB)

				triggerUpdateActionParams := wt.ActionParamsByExistingTrigger(triggerRow)
				triggerUpdateActionParams["PagerDutyIncidentKey"] = pdResponse.IncidentKey

				triggerUpdateActionJSON, err := json.Marshal(triggerUpdateActionParams)
				if err != nil {
					logrus.Error(err)
					continue
				}

				triggerUpdateParams := wt.CreateOrUpdateParameters(triggerRow.ClusterID, triggerRow.WatcherID, triggerRow.LowViolationsCount, triggerRow.HighViolationsCount, triggerUpdateActionJSON)

				_, err = wt.UpdateFromTable(nil, triggerUpdateParams, fmt.Sprintf("id=%v", triggerRow.ID))
				if err != nil {
					logrus.Error(err)
					continue
				}
			}
		}
	}

	return nil
}
