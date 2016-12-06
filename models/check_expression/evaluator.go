package check_expression

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/resourced/resourced-master/models/cassandra"
	"github.com/resourced/resourced-master/models/pg"
)

type CheckExpressionEvaluator struct {
	AppContext context.Context
}

// EvalExpressions reduces the result of expression into a single true/false.
// 1st value: List of all CheckExpression containing results.
// 2nd value: The value of all expressions.
// 3rd value: Error
func (evaluator *CheckExpressionEvaluator) EvalExpressions(checkRow *pg.CheckRow) ([]pg.CheckExpression, bool, error) {
	var hostData map[string][]*cassandra.HostDataRow
	var err error

	if checkRow.HostsQuery != "" {
		hostData, err = cassandra.NewHostData(evaluator.AppContext).AllByClusterIDQueryAndUpdatedInterval(checkRow.ClusterID, checkRow.HostsQuery, "5m")

	} else {
		hostnames, err := checkRow.GetHostsList()
		if err == nil && len(hostnames) > 0 {
			hostData, err = cassandra.NewHostData(evaluator.AppContext).AllByClusterIDAndHostnames(checkRow.ClusterID, hostnames)
		}
	}

	if err != nil {
		return nil, false, err
	}

	expressions, err := checkRow.GetExpressions()
	if err != nil {
		return nil, false, err
	}

	expressionResults := make([]pg.CheckExpression, 0)
	var finalResult bool
	var lastExpressionBooleanOperator string

	for expIndex, expression := range expressions {
		if expression.Type == "RawHostData" {
			expression = evaluator.EvalRawHostDataExpression(checkRow, hostData, expression)

		} else if expression.Type == "RelativeHostData" {
			expression = evaluator.EvalRelativeHostDataExpression(checkRow, hostData, expression)

		} else if expression.Type == "LogData" {
			expression = evaluator.EvalLogDataExpression(checkRow, hostData, expression)

		} else if expression.Type == "Ping" {
			expression = evaluator.EvalPingExpression(checkRow, hostData, expression)

		} else if expression.Type == "SSH" {
			expression = evaluator.EvalSSHExpression(checkRow, hostData, expression)

		} else if expression.Type == "HTTP" || expression.Type == "HTTPS" {
			expression = evaluator.EvalHTTPExpression(checkRow, hostData, expression)

		} else if expression.Type == "BooleanOperator" {
			lastExpressionBooleanOperator = expression.Operator
		}

		if expIndex == 0 {
			finalResult = expression.Result.Value

		} else {
			if lastExpressionBooleanOperator == "and" {
				finalResult = finalResult && expression.Result.Value

			} else if lastExpressionBooleanOperator == "or" {
				finalResult = finalResult || expression.Result.Value
			}
		}

		expressionResults = append(expressionResults, expression)
	}

	return expressionResults, finalResult, nil
}

func (evaluator *CheckExpressionEvaluator) EvalRawHostDataExpression(checkRow *pg.CheckRow, hostData map[string][]*cassandra.HostDataRow, expression pg.CheckExpression) pg.CheckExpression {
	if hostData == nil || len(hostData) <= 0 {
		expression.Result.Value = true
		expression.Result.Message = "There are no hosts to check"
		return expression
	}

	affectedHosts := 0
	badHostnames := make([]string, 0)
	goodHostnames := make([]string, 0)

	var perHostResult bool

	for hostname, hostData := range hostData {
		var val float64

		for _, hostDatum := range hostData {
			if !strings.HasPrefix(expression.Metric, hostDatum.Key) {
				continue
			}

			if strings.HasSuffix(expression.Metric, hostDatum.Key) {
				valueFloat64, err := strconv.ParseFloat(hostDatum.Value, 64)
				if err == nil {
					val = valueFloat64
				}
				break
			}
		}

		// If a Host does not contain a particular metric,
		// We assume that there's something wrong with it.
		if val < float64(0) {
			perHostResult = true
		}

		if expression.Operator == ">" {
			perHostResult = val > expression.Value

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
			badHostnames = append(badHostnames, hostname)

		} else {
			goodHostnames = append(goodHostnames, hostname)
		}
	}

	expression.Result.Value = affectedHosts >= expression.MinHost
	expression.Result.BadHostnames = badHostnames
	expression.Result.GoodHostnames = goodHostnames

	return expression
}

func (evaluator *CheckExpressionEvaluator) EvalRelativeHostDataExpression(checkRow *pg.CheckRow, hostData map[string][]*cassandra.HostDataRow, expression pg.CheckExpression) pg.CheckExpression {
	if hostData == nil || len(hostData) <= 0 {
		expression.Result.Value = true
		expression.Result.Message = "There are no hosts to check"
		return expression
	}

	affectedHosts := 0
	badHostnames := make([]string, 0)
	goodHostnames := make([]string, 0)

	var perHostResult bool

	for hostname, hostData := range hostData {
		metric, err := cassandra.NewMetric(evaluator.AppContext).GetByClusterIDAndKey(checkRow.ClusterID, expression.Metric)
		if err != nil {
			// If we are unable to pull metric metadata,
			// We assume that there's something wrong with it.
			if strings.Contains(err.Error(), "no rows in result set") {
				perHostResult = true
				affectedHosts = affectedHosts + 1
			}
			continue
		}

		tsMetric := cassandra.NewTSMetric(evaluator.AppContext)

		aggregateData, err := tsMetric.GetAggregateXMinutesByMetricIDAndHostname(checkRow.ClusterID, metric.ID, expression.PrevRange, hostname)
		if err != nil {
			// If a Host does not contain historical data of a particular metric,
			// We assume that there's something wrong with it.
			if strings.Contains(err.Error(), "no rows in result set") {
				perHostResult = true
				affectedHosts = affectedHosts + 1
			}
			continue
		}

		var val float64

		for _, hostDatum := range hostData {
			if !strings.HasPrefix(expression.Metric, hostDatum.Key) {
				continue
			}

			if strings.HasSuffix(expression.Metric, hostDatum.Key) {
				valueFloat64, err := strconv.ParseFloat(hostDatum.Value, 64)
				if err == nil {
					val = valueFloat64
				}
				break
			}
		}

		// If a Host does not contain a particular metric,
		// We assume that there's something wrong with it.
		if val < float64(0) {
			perHostResult = true
		}

		var prevVal float64

		if expression.PrevAggr == "avg" {
			prevVal = aggregateData.Avg
		} else if expression.PrevAggr == "max" {
			prevVal = aggregateData.Max
		} else if expression.PrevAggr == "min" {
			prevVal = aggregateData.Min
		} else if expression.PrevAggr == "sum" {
			prevVal = aggregateData.Sum
		}

		if prevVal < float64(0) {
			continue
		}

		valPercentage := (val / prevVal) * float64(100)

		if expression.Operator == ">" {
			perHostResult = valPercentage > expression.Value

		} else if expression.Operator == "<" {
			perHostResult = valPercentage < expression.Value
		}

		if perHostResult {
			affectedHosts = affectedHosts + 1
			badHostnames = append(badHostnames, hostname)
		} else {
			goodHostnames = append(goodHostnames, hostname)
		}
	}

	expression.Result.Value = affectedHosts >= expression.MinHost
	expression.Result.BadHostnames = badHostnames
	expression.Result.GoodHostnames = goodHostnames

	return expression
}

func (evaluator *CheckExpressionEvaluator) EvalLogDataExpression(checkRow *pg.CheckRow, hostData map[string][]*cassandra.HostDataRow, expression pg.CheckExpression) pg.CheckExpression {
	hostnames, err := checkRow.GetHostsList()
	if err != nil {
		expression.Result.Value = false
		return expression
	}

	if len(hostnames) == 0 && hostData != nil && len(hostData) > 0 {
		hostnames = make([]string, len(hostData))

		for hostname, _ := range hostData {
			hostnames = append(hostnames, hostname)
		}
	}

	if hostnames == nil || len(hostnames) <= 0 {
		expression.Result.Value = false
		return expression
	}

	affectedHosts := 0
	badHostnames := make([]string, 0)
	goodHostnames := make([]string, 0)

	clusterRow, err := pg.NewCluster(evaluator.AppContext).GetByID(nil, checkRow.ClusterID)
	if err != nil {
		expression.Result.Value = false
		return expression
	}

	deletedFrom := clusterRow.GetDeletedFromUNIXTimestampForSelect("ts_logs")

	var perHostResult bool

	for _, hostname := range hostnames {
		now := time.Now().UTC()
		from := now.Add(-1 * time.Duration(expression.PrevRange) * time.Minute).UTC().Unix()
		searchQuery := fmt.Sprintf(`logline search "%v"`, expression.Search)

		valInt64, err := pg.NewTSLog(evaluator.AppContext, checkRow.ClusterID).CountByClusterIDFromTimestampHostAndQuery(nil, checkRow.ClusterID, from, hostname, searchQuery, deletedFrom)
		if err != nil {
			continue
		}

		val := float64(valInt64)

		if val < float64(0) {
			continue
		}

		if expression.Operator == ">" {
			perHostResult = val > expression.Value

		} else if expression.Operator == "<" {
			perHostResult = val < expression.Value
		}

		if perHostResult {
			affectedHosts = affectedHosts + 1
			badHostnames = append(badHostnames, hostname)
		} else {
			goodHostnames = append(goodHostnames, hostname)
		}
	}

	expression.Result.Value = affectedHosts >= expression.MinHost
	expression.Result.BadHostnames = badHostnames
	expression.Result.GoodHostnames = goodHostnames

	return expression
}

func (evaluator *CheckExpressionEvaluator) CheckPing(hostname string) (outBytes []byte, err error) {
	return exec.Command("ping", "-c", "1", hostname).CombinedOutput()
}

func (evaluator *CheckExpressionEvaluator) EvalPingExpression(checkRow *pg.CheckRow, hostData map[string][]*cassandra.HostDataRow, expression pg.CheckExpression) pg.CheckExpression {
	hostnames, err := checkRow.GetHostsList()
	if err != nil {
		expression.Result.Value = false
		return expression
	}

	if len(hostnames) == 0 && hostData != nil && len(hostData) > 0 {
		hostnames = make([]string, len(hostData))

		for hostname, _ := range hostData {
			hostnames = append(hostnames, hostname)
		}
	}

	if hostnames == nil || len(hostnames) <= 0 {
		expression.Result.Value = false
		return expression
	}

	affectedHosts := 0
	badHostnames := make([]string, 0)
	goodHostnames := make([]string, 0)

	for _, hostname := range hostnames {
		_, err := evaluator.CheckPing(hostname)
		if err != nil {
			affectedHosts = affectedHosts + 1
			badHostnames = append(badHostnames, hostname)
		} else {
			goodHostnames = append(goodHostnames, hostname)
		}
	}

	expression.Result.Value = affectedHosts >= expression.MinHost
	expression.Result.BadHostnames = badHostnames
	expression.Result.GoodHostnames = goodHostnames

	return expression
}

func (evaluator *CheckExpressionEvaluator) CheckSSH(hostname, port, user string) (outBytes []byte, err error) {
	sshOptions := []string{"-o BatchMode=yes", "-o ConnectTimeout=10"}

	if port != "" {
		sshOptions = append(sshOptions, []string{"-p", port}...)
	}

	userAtHost := hostname

	if user != "" {
		userAtHost = fmt.Sprintf("%v@%v", user, hostname)
	}

	sshOptions = append(sshOptions, userAtHost)

	return exec.Command("ssh", sshOptions...).CombinedOutput()
}

func (evaluator *CheckExpressionEvaluator) EvalSSHExpression(checkRow *pg.CheckRow, hostData map[string][]*cassandra.HostDataRow, expression pg.CheckExpression) pg.CheckExpression {
	hostnames, err := checkRow.GetHostsList()
	if err != nil {
		expression.Result.Value = false
		return expression
	}

	if len(hostnames) == 0 && hostData != nil && len(hostData) > 0 {
		hostnames = make([]string, len(hostData))

		for hostname, _ := range hostData {
			hostnames = append(hostnames, hostname)
		}
	}

	if hostnames == nil || len(hostnames) <= 0 {
		expression.Result.Value = false
		return expression
	}

	affectedHosts := 0
	badHostnames := make([]string, 0)
	goodHostnames := make([]string, 0)

	for _, hostname := range hostnames {
		outputBytes, err := evaluator.CheckSSH(hostname, expression.Port, expression.Username)
		outputString := string(outputBytes)

		if err != nil && !strings.Contains(outputString, "Permission denied") && !strings.Contains(outputString, "Host key verification failed") {
			affectedHosts = affectedHosts + 1
			badHostnames = append(badHostnames, hostname)
		} else {
			goodHostnames = append(goodHostnames, hostname)
		}
	}

	expression.Result.Value = affectedHosts >= expression.MinHost
	expression.Result.BadHostnames = badHostnames
	expression.Result.GoodHostnames = goodHostnames

	return expression
}

func (evaluator *CheckExpressionEvaluator) CheckHTTP(hostname, scheme, port, method, user, pass string, headers map[string]string, body string) (resp *http.Response, err error) {
	url := fmt.Sprintf("%v://%v:%s", scheme, hostname, port)

	client := &http.Client{}

	var req *http.Request

	if body != "" {
		req, err = http.NewRequest(strings.ToUpper(method), url, strings.NewReader(body))

		// Detect if POST body is JSON and set content-type
		if strings.HasPrefix(body, "{") || strings.HasPrefix(body, "[") {
			req.Header.Set("Content-Type", "application/json")
		} else {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}

	} else {
		req, err = http.NewRequest(strings.ToUpper(method), url, nil)
	}

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Method":     "http.NewRequest",
			"URL":        url,
			"HTTPMethod": method,
		}).Error(err)
		return nil, err
	}

	for headerKey, headerVal := range headers {
		req.Header.Add(headerKey, headerVal)
	}

	if user != "" || pass != "" {
		req.SetBasicAuth(user, pass)
	}

	resp, err = client.Do(req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Method":     "http.Client{}.Do",
			"URL":        url,
			"HTTPMethod": method,
		}).Error(err)
		return nil, err

	} else if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}

	return resp, err
}

func (evaluator *CheckExpressionEvaluator) EvalHTTPExpression(checkRow *pg.CheckRow, hostData map[string][]*cassandra.HostDataRow, expression pg.CheckExpression) pg.CheckExpression {
	hostnames, err := checkRow.GetHostsList()
	if err != nil {
		expression.Result.Value = false
		return expression
	}

	if len(hostnames) == 0 && hostData != nil && len(hostData) > 0 {
		hostnames = make([]string, len(hostData))

		for hostname, _ := range hostData {
			hostnames = append(hostnames, hostname)
		}
	}

	if hostnames == nil || len(hostnames) <= 0 {
		expression.Result.Value = false
		return expression
	}

	affectedHosts := 0
	badHostnames := make([]string, 0)
	goodHostnames := make([]string, 0)

	for _, hostname := range hostnames {
		headers := make(map[string]string)

		for _, headersNewLine := range strings.Split(expression.Headers, "\n") {
			for _, kvString := range strings.Split(headersNewLine, ",") {
				if strings.Contains(kvString, "=") {
					kvSlice := strings.Split(kvString, "=")
					if len(kvSlice) >= 2 {
						headers[strings.TrimSpace(kvSlice[0])] = strings.TrimSpace(kvSlice[1])
					}
				}
			}
		}

		resp, err := evaluator.CheckHTTP(hostname, expression.Protocol, expression.Port, expression.HTTPMethod, expression.Username, expression.Password, headers, expression.HTTPBody)
		if err != nil || (resp != nil && resp.StatusCode != 200) {
			affectedHosts = affectedHosts + 1
			badHostnames = append(badHostnames, hostname)
		} else {
			goodHostnames = append(goodHostnames, hostname)
		}
	}

	expression.Result.Value = affectedHosts >= expression.MinHost
	expression.Result.BadHostnames = badHostnames
	expression.Result.GoodHostnames = goodHostnames

	return expression
}
