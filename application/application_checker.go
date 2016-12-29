package application

import (
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/resourced/resourced-master/models/cassandra"
	"github.com/resourced/resourced-master/models/check_expression"
)

// CheckAndRunTriggers pulls list of all checks, distributed evenly across N master daemons,
// evaluates the checks and run triggers when conditions are met.
func (app *Application) CheckAndRunTriggers() {
	checkRowsChan := make(chan []*cassandra.CheckRow)

	// Fetch Checks data, split by number of daemons, every time there's a value in app.RefetchChecksChan
	go func() {
		for refetchChecks := range app.RefetchChecksChan {
			if refetchChecks {
				daemons := make([]string, 0)

				for hostAndPort, _ := range app.Peers.Items() {
					daemons = append(daemons, hostAndPort)
				}

				groupedCheckRows, _ := cassandra.NewCheck(app.GetContext()).AllSplitToDaemons(daemons)
				checkRowsChan <- groupedCheckRows[app.FullAddr()]
			}
		}
	}()

	go func() {
		for checkRows := range checkRowsChan {
			for _, checkRow := range checkRows {
				go func(checkRow *cassandra.CheckRow) {
					checkDuration, err := time.ParseDuration(checkRow.Interval)
					if err != nil {
						app.ErrLogger.WithFields(logrus.Fields{
							"ClusterID": checkRow.ClusterID,
							"CheckID":   checkRow.ID,
							"Error":     err,
						}).Error("Failed to parse checkRow.Interval")
						return
					}

					for range time.Tick(checkDuration) {
						// 1. Evaluate all expressions in a check.
						evaluator := &check_expression.CheckExpressionEvaluator{
							AppContext: app.GetContext(),
						}

						expressionResults, finalResult, err := evaluator.EvalExpressions(checkRow)
						if err != nil {
							app.ErrLogger.WithFields(logrus.Fields{
								"Method":    "checkRow.EvalExpressions",
								"ClusterID": checkRow.ClusterID,
								"CheckID":   checkRow.ID,
							}).Error(err)
						}

						if err != nil || expressionResults == nil || len(expressionResults) == 0 {
							return
						}

						// 2. Store the check result.
						clusterRow, err := cassandra.NewCluster(app.GetContext()).GetByID(checkRow.ClusterID)
						if err != nil {
							app.ErrLogger.WithFields(logrus.Fields{
								"Method":    "Cluster.GetByID",
								"ClusterID": checkRow.ClusterID,
								"CheckID":   checkRow.ID,
							}).Error(err)
							return
						}

						ttl := clusterRow.GetTTLDurationForInsert("ts_checks")

						err = cassandra.NewTSCheck(app.GetContext()).Create(checkRow.ClusterID, checkRow.ID, finalResult, expressionResults, ttl)
						if err != nil {
							app.ErrLogger.WithFields(logrus.Fields{
								"Method":    "TSCheck.Create",
								"ClusterID": checkRow.ClusterID,
								"CheckID":   checkRow.ID,
								"Result":    finalResult,
							}).Error(err)
							return
						}

						// 3. Run check's triggers.
						err = checkRow.RunTriggers(app.GetContext())
						// err = checkRow.RunTriggers(app.GeneralConfig, app.PGDBConfig.Core, app.PGDBConfig.GetTSCheck(checkRow.ClusterID), app.Mailers["GeneralConfig.Checks"])
						if err != nil {
							app.ErrLogger.WithFields(logrus.Fields{
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
