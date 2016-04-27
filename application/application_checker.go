package application

import (
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libtime"
)

func (app *Application) CheckAll() {
	checkRowsChan := make(chan []*dal.CheckRow)

	// Fetch daemons and checks data every 5 minutes
	go func() {
		for {
			daemonHostnames, _ := dal.NewDaemon(app.DBConfig.Core).AllHostnames(nil)
			groupedCheckRows, _ := dal.NewCheck(app.DBConfig.Core).AllSplitToDaemons(nil, daemonHostnames)
			checkRowsChan <- groupedCheckRows[app.Hostname]

			libtime.SleepString(app.GeneralConfig.Watchers.ListFetchInterval)
		}
	}()

	go func() {
		for {
			select {
			case checkRows := <-checkRowsChan:
				for _, checkRow := range checkRows {
					go func(checkRow *dal.CheckRow) {
						expressionResults, finalResult, err := checkRow.EvalExpressions(app.DBConfig.Core, app.DBConfig.TSMetric, app.DBConfig.TSLog)

						if err != nil || expressionResults == nil || len(expressionResults) == 0 {
							libtime.SleepString(checkRow.Interval)
							return
						}

						println(finalResult)

						libtime.SleepString(checkRow.Interval)
					}(checkRow)
				}
			}
		}
	}()
}
