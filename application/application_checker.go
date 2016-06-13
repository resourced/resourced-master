package application

import (
	"github.com/Sirupsen/logrus"

	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libtime"
)

// CheckAndRunTriggers pulls list of all checks, distributed evenly across N master daemons,
// evaluates the checks and run triggers when conditions are met.
func (app *Application) CheckAndRunTriggers(refetchChecksChan <-chan bool) {
	checkRowsChan := make(chan []*dal.CheckRow)

	// Fetch Checks data, split by number of daemons, every time there's a value in refetchChecksChan
	go func() {
		select {
		case refetchChecks := <-refetchChecksChan:
			if refetchChecks {
				daemons := make([]string, 0)

				for hostAndPort, _ := range app.Peers.All() {
					daemons = append(daemons, hostAndPort)
				}

				groupedCheckRows, _ := dal.NewCheck(app.DBConfig.Core).AllSplitToDaemons(nil, daemons)
				checkRowsChan <- groupedCheckRows[app.FullAddr()]
			}
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
						clusterRow, err := dal.NewCluster(app.DBConfig.Core).GetByID(nil, checkRow.ClusterID)
						if err != nil {
							logrus.WithFields(logrus.Fields{
								"Method":    "Cluster.GetByID",
								"ClusterID": checkRow.ClusterID,
								"CheckID":   checkRow.ID,
							}).Error(err)

							libtime.SleepString(checkRow.Interval)
							return
						}

						deletedFrom := clusterRow.GetDeletedFromUNIXTimestampForInsert("ts_checks")

						err = dal.NewTSCheck(app.DBConfig.TSCheck).Create(nil, checkRow.ClusterID, checkRow.ID, finalResult, expressionResults, deletedFrom)
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
