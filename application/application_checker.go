package application

import (
	"github.com/Sirupsen/logrus"

	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libtime"
)

// CheckAndRunTriggers pulls list of all checks, distributed evenly across N master daemons,
// evaluates the checks and run triggers when conditions are met.
func (app *Application) CheckAndRunTriggers() {
	checkRowsChan := make(chan []*dal.CheckRow)

	// Fetch daemons and checks data every 5 minutes
	go func() {
		for {
			daemonHostnames, _ := dal.NewDaemon(app.DBConfig.Core).AllHostnames(nil)
			groupedCheckRows, _ := dal.NewCheck(app.DBConfig.Core).AllSplitToDaemons(nil, daemonHostnames)
			checkRowsChan <- groupedCheckRows[app.Hostname]

			libtime.SleepString(app.GeneralConfig.Checks.ListFetchInterval)
		}
	}()

	go func() {
		for {
			select {
			case checkRows := <-checkRowsChan:
				for _, checkRow := range checkRows {
					go func(checkRow *dal.CheckRow) {
						// 1. Evaluate all expressions in a check.
						expressionResults, finalResult, err := checkRow.EvalExpressions(app.DBConfig.Core, app.DBConfig.TSMetric, app.DBConfig.TSLog)
						if err != nil {
							logrus.WithFields(logrus.Fields{
								"Method":    "checkRow.EvalExpressions",
								"ClusterID": checkRow.ClusterID,
								"CheckID":   checkRow.ID,
							}).Error(err)
						}

						if err != nil || expressionResults == nil || len(expressionResults) == 0 {
							libtime.SleepString(checkRow.Interval)
							return
						}

						// 2. Store the check result.
						err = dal.NewTSCheck(app.DBConfig.TSCheck).Create(nil, checkRow.ClusterID, checkRow.ID, finalResult, expressionResults)
						if err != nil {
							logrus.WithFields(logrus.Fields{
								"Method":    "TSCheck.Create",
								"ClusterID": checkRow.ClusterID,
								"CheckID":   checkRow.ID,
								"Result":    finalResult,
							}).Error(err)

							libtime.SleepString(checkRow.Interval)
							return
						}

						// 3. Run check's triggers.
						err = checkRow.RunTriggers(app.GeneralConfig, app.DBConfig.TSCheck, app.Mailers["GeneralConfig.Checks"])
						if err != nil {
							logrus.WithFields(logrus.Fields{
								"Method":    "checkRow.RunTriggers",
								"ClusterID": checkRow.ClusterID,
								"CheckID":   checkRow.ID,
							}).Error(err)

							libtime.SleepString(checkRow.Interval)
							return
						}

						libtime.SleepString(checkRow.Interval)
					}(checkRow)
				}
			}
		}
	}()
}
