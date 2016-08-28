package application

import (
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/resourced/resourced-master/dal"
)

// CheckAndRunTriggers pulls list of all checks, distributed evenly across N master daemons,
// evaluates the checks and run triggers when conditions are met.
func (app *Application) CheckAndRunTriggers() {
	checkRowsChan := make(chan []*dal.CheckRow)

	// Fetch Checks data, split by number of daemons, every time there's a value in app.RefetchChecksChan
	go func() {
		for refetchChecks := range app.RefetchChecksChan {
			if refetchChecks {
				daemons := make([]string, 0)

				for hostAndPort, _ := range app.Peers.Items() {
					daemons = append(daemons, hostAndPort)
				}

				groupedCheckRows, _ := dal.NewCheck(app.DBConfig.Core).AllSplitToDaemons(nil, daemons)
				checkRowsChan <- groupedCheckRows[app.FullAddr()]
			}
		}
	}()

	go func() {
		for checkRows := range checkRowsChan {
			for _, checkRow := range checkRows {
				go func(checkRow *dal.CheckRow) {
					checkDuration, err := time.ParseDuration(checkRow.Interval)
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"ClusterID": checkRow.ClusterID,
							"CheckID":   checkRow.ID,
							"Error":     err,
						}).Error("Failed to parse checkRow.Interval")
						return
					}

					for range time.Tick(checkDuration) {
						// 1. Evaluate all expressions in a check.
						expressionResults, finalResult, err := checkRow.EvalExpressions(app.DBConfig)
						if err != nil {
							logrus.WithFields(logrus.Fields{
								"Method":    "checkRow.EvalExpressions",
								"ClusterID": checkRow.ClusterID,
								"CheckID":   checkRow.ID,
							}).Error(err)
						}

						if err != nil || expressionResults == nil || len(expressionResults) == 0 {
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
							return
						}

						deletedFrom := clusterRow.GetDeletedFromUNIXTimestampForInsert("ts_checks")

						err = dal.NewTSCheck(app.DBConfig.GetTSCheck(checkRow.ClusterID)).Create(nil, checkRow.ClusterID, checkRow.ID, finalResult, expressionResults, deletedFrom)
						if err != nil {
							logrus.WithFields(logrus.Fields{
								"Method":    "TSCheck.Create",
								"ClusterID": checkRow.ClusterID,
								"CheckID":   checkRow.ID,
								"Result":    finalResult,
							}).Error(err)
							return
						}

						// 3. Run check's triggers.
						err = checkRow.RunTriggers(app.GeneralConfig, app.DBConfig.Core, app.DBConfig.GetTSCheck(checkRow.ClusterID), app.Mailers["GeneralConfig.Checks"])
						if err != nil {
							logrus.WithFields(logrus.Fields{
								"Method":    "checkRow.RunTriggers",
								"ClusterID": checkRow.ClusterID,
								"CheckID":   checkRow.ID,
							}).Error(err)
							return
						}
					}
				}(checkRow)
			}
		}
	}()
}
