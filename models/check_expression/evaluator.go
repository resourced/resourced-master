package check_expression

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/resourced/resourced-master/models/pg"
	"github.com/resourced/resourced-master/models/shims"
)

type CheckExpressionEvaluator struct {
	AppContext context.Context
}

// EvalExpressions reduces the result of expression into a single true/false.
// 1st value: List of all CheckExpression containing results.
// 2nd value: The value of all expressions.
// 3rd value: Error
func (evaluator *CheckExpressionEvaluator) EvalExpressions(checkRow *pg.CheckRow) ([]pg.CheckExpression, bool, error) {
	var hostRows []*pg.HostRow
	var err error

	host := pg.NewHost(evaluator.AppContext, checkRow.ClusterID)

	if checkRow.HostsQuery != "" {
		hostRows, err = host.AllByClusterIDQueryAndUpdatedInterval(nil, checkRow.ClusterID, checkRow.HostsQuery, "5m")

	} else {
		hostnames, err := checkRow.GetHostsList()
		if err == nil && len(hostnames) > 0 {
			hostRows, err = host.AllByClusterIDAndHostnames(nil, checkRow.ClusterID, hostnames)
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
			expression = evaluator.EvalRawHostDataExpression(checkRow, hostRows, expression)

		} else if expression.Type == "RelativeHostData" {
			expression = evaluator.EvalRelativeHostDataExpression(checkRow, hostRows, expression)

		} else if expression.Type == "LogData" {
			expression = evaluator.EvalLogDataExpression(checkRow, hostRows, expression)

		} else if expression.Type == "Ping" {
			expression = evaluator.EvalPingExpression(checkRow, hostRows, expression)

		} else if expression.Type == "SSH" {
			expression = evaluator.EvalSSHExpression(checkRow, hostRows, expression)

		} else if expression.Type == "HTTP" || expression.Type == "HTTPS" {
			expression = evaluator.EvalHTTPExpression(checkRow, hostRows, expression)

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

func (evaluator *CheckExpressionEvaluator) EvalRawHostDataExpression(checkRow *pg.CheckRow, hostRows []*pg.HostRow, expression pg.CheckExpression) pg.CheckExpression {
	if hostRows == nil || len(hostRows) <= 0 {
		expression.Result.Value = true
		expression.Result.Message = "There are no hosts to check"
		return expression
	}

	affectedHosts := 0
	badHostnames := make([]string, 0)
	goodHostnames := make([]string, 0)

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
			badHostnames = append(badHostnames, hostRow.Hostname)

		} else {
			goodHostnames = append(goodHostnames, hostRow.Hostname)
		}
	}

	expression.Result.Value = affectedHosts >= expression.MinHost
	expression.Result.BadHostnames = badHostnames
	expression.Result.GoodHostnames = goodHostnames

	return expression
}

func (evaluator *CheckExpressionEvaluator) EvalRelativeHostDataExpression(checkRow *pg.CheckRow, hostRows []*pg.HostRow, expression pg.CheckExpression) pg.CheckExpression {
	if hostRows == nil || len(hostRows) <= 0 {
		expression.Result.Value = true
		expression.Result.Message = "There are no hosts to check"
		return expression
	}

	affectedHosts := 0
	badHostnames := make([]string, 0)
	goodHostnames := make([]string, 0)

	var perHostResult bool

	for _, hostRow := range hostRows {
		metric, err := pg.NewMetric(evaluator.AppContext).GetByClusterIDAndKey(nil, checkRow.ClusterID, expression.Metric)
		if err != nil {
			// If we are unable to pull metric metadata,
			// We assume that there's something wrong with it.
			if strings.Contains(err.Error(), "no rows in result set") {
				perHostResult = true
				affectedHosts = affectedHosts + 1
			}
			continue
		}

		shimsTSMetric := shims.NewTSMetric(evaluator.AppContext, hostRow.ClusterID)

		aggregateData, err := shimsTSMetric.GetAggregateXMinutesByMetricIDAndHostname(checkRow.ClusterID, metric.ID, expression.PrevRange, hostRow.Hostname)
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
			badHostnames = append(badHostnames, hostRow.Hostname)
		} else {
			goodHostnames = append(goodHostnames, hostRow.Hostname)
		}
	}

	expression.Result.Value = affectedHosts >= expression.MinHost
	expression.Result.BadHostnames = badHostnames
	expression.Result.GoodHostnames = goodHostnames

	return expression
}

func (evaluator *CheckExpressionEvaluator) EvalLogDataExpression(checkRow *pg.CheckRow, hostRows []*pg.HostRow, expression pg.CheckExpression) pg.CheckExpression {
	hostnames, err := checkRow.GetHostsList()
	if err != nil {
		expression.Result.Value = false
		return expression
	}

	if len(hostnames) == 0 && hostRows != nil && len(hostRows) > 0 {
		hostnames = make([]string, len(hostRows))

		for i, hostRow := range hostRows {
			hostnames[i] = hostRow.Hostname
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

func (evaluator *CheckExpressionEvaluator) EvalPingExpression(checkRow *pg.CheckRow, hostRows []*pg.HostRow, expression pg.CheckExpression) pg.CheckExpression {
	hostnames, err := checkRow.GetHostsList()
	if err != nil {
		expression.Result.Value = false
		return expression
	}

	if len(hostnames) == 0 && hostRows != nil && len(hostRows) > 0 {
		hostnames = make([]string, len(hostRows))

		for i, hostRow := range hostRows {
			hostnames[i] = hostRow.Hostname
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

func (evaluator *CheckExpressionEvaluator) EvalSSHExpression(checkRow *pg.CheckRow, hostRows []*pg.HostRow, expression pg.CheckExpression) pg.CheckExpression {
	hostnames, err := checkRow.GetHostsList()
	if err != nil {
		expression.Result.Value = false
		return expression
	}

	if len(hostnames) == 0 && hostRows != nil && len(hostRows) > 0 {
		hostnames = make([]string, len(hostRows))

		for i, hostRow := range hostRows {
			hostnames[i] = hostRow.Hostname
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

func (evaluator *CheckExpressionEvaluator) EvalHTTPExpression(checkRow *pg.CheckRow, hostRows []*pg.HostRow, expression pg.CheckExpression) pg.CheckExpression {
	hostnames, err := checkRow.GetHostsList()
	if err != nil {
		expression.Result.Value = false
		return expression
	}

	if len(hostnames) == 0 && hostRows != nil && len(hostRows) > 0 {
		hostnames = make([]string, len(hostRows))

		for i, hostRow := range hostRows {
			hostnames[i] = hostRow.Hostname
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
