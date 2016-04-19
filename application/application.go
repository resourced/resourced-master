package application

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/carbocation/interpose"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"github.com/marcw/pagerduty"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/dal"
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

func (app *Application) PruneAll() {
	for {
		app.PruneTSWatcherOnce()
		app.PruneTSMetricOnce()
		app.PruneTSEventOnce()
		app.PruneTSExecutorLogOnce()
		app.PruneTSLogOnce()

		libtime.SleepString("24h")
	}
}

func (app *Application) PruneTSWatcherOnce() {
	go func() {
		err := dal.NewTSWatcher(app.DBConfig.TSWatcher).DeleteByDayInterval(nil, app.GeneralConfig.Watchers.DataRetention)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Method":        "TSWatcher.DeleteByDayInterval",
				"DataRetention": app.GeneralConfig.Watchers.DataRetention,
			}).Error(err)
		}
	}()
}

func (app *Application) PruneTSMetricOnce() {
	go func() {
		err := dal.NewTSMetric(app.DBConfig.TSMetric).DeleteByDayInterval(nil, app.GeneralConfig.Metrics.DataRetention)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Method":        "TSMetric.DeleteByDayInterval",
				"DataRetention": app.GeneralConfig.Metrics.DataRetention,
			}).Error(err)
		}
	}()
}

func (app *Application) PruneTSEventOnce() {
	go func() {
		err := dal.NewTSEvent(app.DBConfig.TSEvent).DeleteByDayInterval(nil, app.GeneralConfig.Events.DataRetention)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Method":        "TSEvent.DeleteByDayInterval",
				"DataRetention": app.GeneralConfig.Events.DataRetention,
			}).Error(err)
		}
	}()
}

func (app *Application) PruneTSExecutorLogOnce() {
	go func() {
		err := dal.NewTSExecutorLog(app.DBConfig.TSExecutorLog).DeleteByDayInterval(nil, app.GeneralConfig.ExecutorLogs.DataRetention)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Method":        "TSExecutorLog.DeleteByDayInterval",
				"DataRetention": app.GeneralConfig.ExecutorLogs.DataRetention,
			}).Error(err)
		}
	}()
}

func (app *Application) PruneTSLogOnce() {
	go func() {
		err := dal.NewTSLog(app.DBConfig.TSLog).DeleteByDayInterval(nil, app.GeneralConfig.Logs.DataRetention)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Method":        "TSLog.DeleteByDayInterval",
				"DataRetention": app.GeneralConfig.Logs.DataRetention,
			}).Error(err)
		}
	}()
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
			tsWatcherDataHosts[i] = affectedHost.Hostname
		}

		tsWatcherData := make(map[string]interface{})
		tsWatcherData["hosts"] = tsWatcherDataHosts

		jsonData, err := json.Marshal(tsWatcherData)
		if err != nil {
			return err
		}

		// Write to ts_watchers asynchronously
		go func() {
			err := dal.NewTSWatcher(app.DBConfig.TSWatcher).Create(nil, clusterID, watcherRow.ID, numAffectedHosts, jsonData)
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
		go func() {
			err := dal.NewTSWatcher(app.DBConfig.TSWatcher).Create(nil, clusterID, watcherRow.ID, int64(numAffectedHosts), jsonData)
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

	return nil
}

func (app *Application) RunTrigger(clusterID int64, watcherRow *dal.WatcherRow) error {
	// Don't do anything if watcher is silenced.
	if watcherRow.IsSilenced {
		return nil
	}

	triggerRows, err := dal.NewWatcherTrigger(app.DBConfig.Core).AllByClusterIDAndWatcherID(nil, clusterID, watcherRow.ID)
	if err != nil {
		return err
	}

	for _, triggerRow := range triggerRows {
		tsWatchers, err := dal.NewTSWatcher(app.DBConfig.TSWatcher).AllViolationsByClusterIDWatcherIDAndInterval(nil, clusterID, watcherRow.ID, watcherRow.LowAffectedHosts, triggerRow.CreatedInterval)
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
