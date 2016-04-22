package application

import (
	"strings"

	_ "github.com/lib/pq"

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
						var hostRows []*dal.HostRow
						var err error

						if checkRow.HostsQuery != "" {
							hostRows, err = dal.NewHost(app.DBConfig.Core).AllByClusterIDQueryAndUpdatedInterval(nil, checkRow.ClusterID, checkRow.HostsQuery, "5m")

						} else {
							hostnames, err := checkRow.GetHostsList()
							if err == nil && len(hostnames) > 0 {
								hostRows, err = dal.NewHost(app.DBConfig.Core).AllByClusterIDAndHostnames(nil, checkRow.ClusterID, hostnames)
							}
						}

						if hostRows != nil {
							println("Check:")
							println(checkRow.ID)
							println(checkRow.Name)
							println("")

							println("Hosts Length:")
							println(len(hostRows))
							println("")
						}

						if err != nil || hostRows == nil || len(hostRows) == 0 {
							libtime.SleepString(checkRow.Interval)
							return
						}

						expressions, err := checkRow.GetExpressions()
						if err != nil {
							println(err.Error())
							libtime.SleepString(checkRow.Interval)
							return
						}

						println("Expressions Length:")
						println(len(expressions))
						println("")

						expressionResults := make([]dal.CheckExpression, 0)
						var finalResult bool
						var lastExpressionBooleanOperator string

						for expIndex, expression := range expressions {
							if expression.Type == "RawHostData" {
								affectedHosts := 0
								var perHostResult bool

								for _, hostRow := range hostRows {
									var val float64

									for prefix, keyAndValue := range hostRow.DataAsFlatKeyValue() {
										if !strings.HasPrefix(expression.Metric, prefix) {
											continue
										}

										for key, value := range keyAndValue {
											if strings.HasSuffix(expression.Metric, key) {
												val = value.(float64)
												break
											}
										}
									}

									if val < float64(0) {
										continue
									}

									if expression.Operator == ">" {
										println("eval >")
										println(val)
										println(expression.Value)
										println(val > expression.Value)

										perHostResult = val > expression.Value

										println(perHostResult)
										println("")

									} else if expression.Operator == ">=" {
										perHostResult = val >= expression.Value

									} else if expression.Operator == "=" {
										perHostResult = val == expression.Value

									} else if expression.Operator == "<" {
										perHostResult = val < expression.Value

									} else if expression.Operator == "<=" {
										perHostResult = val <= expression.Value
									}

									if perHostResult {
										affectedHosts = affectedHosts + 1
									}
								}

								println("calculating expression.Result")
								println(affectedHosts)
								println(expression.MinHost)

								expression.Result = affectedHosts >= expression.MinHost

							} else if expression.Type == "RelativeHostData" {

							} else if expression.Type == "LogData" {

							} else if expression.Type == "LogData" {

							} else if expression.Type == "Ping" {

							} else if expression.Type == "SSH" {

							} else if expression.Type == "HTTP" {

							} else if expression.Type == "BooleanOperator" {
								lastExpressionBooleanOperator = expression.Operator
							}

							if expIndex == 0 {
								finalResult = expression.Result

							} else {
								if lastExpressionBooleanOperator == "and" {
									finalResult = finalResult && expression.Result

								} else if lastExpressionBooleanOperator == "or" {
									finalResult = finalResult || expression.Result
								}
							}

							expressionResults = append(expressionResults, expression)
						}

						println("Final Result:")
						println(finalResult)

						libtime.SleepString(checkRow.Interval)
					}(checkRow)
				}
			}
		}
	}()
}
