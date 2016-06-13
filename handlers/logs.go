package handlers

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/context"
	"github.com/gorilla/csrf"
	"github.com/jmoiron/sqlx"

	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func GetLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	qParams := r.URL.Query()

	toString := qParams.Get("To")
	if toString == "" {
		toString = qParams.Get("to")
	}
	to, err := strconv.ParseInt(toString, 10, 64)

	fromString := qParams.Get("From")
	if fromString == "" {
		fromString = qParams.Get("from")
	}
	from, err := strconv.ParseInt(fromString, 10, 64)

	query := qParams.Get("q")

	currentUser := context.Get(r, "currentUser").(*dal.UserRow)

	currentCluster := context.Get(r, "currentCluster").(*dal.ClusterRow)

	db := context.Get(r, "db.Core").(*sqlx.DB)

	tsLogsDB := context.Get(r, "db.TSLog").(*sqlx.DB)

	// -----------------------------------
	// Create channels to receive SQL rows
	// -----------------------------------
	tsLogsChan := make(chan *dal.TSLogRowsWithError)
	defer close(tsLogsChan)

	savedQueriesChan := make(chan *dal.SavedQueryRowsWithError)
	defer close(savedQueriesChan)

	// --------------------------
	// Fetch SQL rows in parallel
	// --------------------------
	go func(currentCluster *dal.ClusterRow, query string) {
		var tsLogs []*dal.TSLogRow
		var err error

		deletedFrom := currentCluster.GetDeletedFromUNIXTimestampForSelect("ts_logs")

		if fromString == "" && toString == "" {
			tsLogs, err = dal.NewTSLog(tsLogsDB).AllByClusterIDLastRowIntervalAndQuery(nil, currentCluster.ID, "15 minute", query, deletedFrom)

			if len(tsLogs) > 0 {
				if from == 0 {
					from = tsLogs[0].Created.UTC().Unix()
				}
				if to == 0 {
					to = tsLogs[len(tsLogs)-1].Created.UTC().Unix()
				}
			}

		} else {
			tsLogs, err = dal.NewTSLog(tsLogsDB).AllByClusterIDRangeAndQuery(nil, currentCluster.ID, from, to, query, deletedFrom)
		}

		tsLogRowsWithError := &dal.TSLogRowsWithError{}
		tsLogRowsWithError.TSLogRows = tsLogs
		tsLogRowsWithError.Error = err
		tsLogsChan <- tsLogRowsWithError
	}(currentCluster, query)

	go func(currentCluster *dal.ClusterRow) {
		savedQueriesWithError := &dal.SavedQueryRowsWithError{}
		savedQueriesWithError.SavedQueries, savedQueriesWithError.Error = dal.NewSavedQuery(db).AllByClusterIDAndType(nil, currentCluster.ID, "logs")
		savedQueriesChan <- savedQueriesWithError
	}(currentCluster)

	// -----------------------------------
	// Wait for channels to return results
	// -----------------------------------
	tsLogsWithError := <-tsLogsChan
	if tsLogsWithError.Error != nil && tsLogsWithError.Error.Error() != "sql: no rows in result set" {
		libhttp.HandleErrorJson(w, tsLogsWithError.Error)
		return
	}

	savedQueriesWithError := <-savedQueriesChan
	if savedQueriesWithError.Error != nil && savedQueriesWithError.Error.Error() != "sql: no rows in result set" {
		libhttp.HandleErrorJson(w, savedQueriesWithError.Error)
		return
	}

	accessToken, err := getAccessToken(w, r, "read")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	data := struct {
		CSRFToken      string
		Addr           string
		CurrentUser    *dal.UserRow
		AccessToken    *dal.AccessTokenRow
		Clusters       []*dal.ClusterRow
		CurrentCluster *dal.ClusterRow
		Logs           []*dal.TSLogRow
		SavedQueries   []*dal.SavedQueryRow
		From           int64
		To             int64
	}{
		csrf.Token(r),
		context.Get(r, "addr").(string),
		currentUser,
		accessToken,
		context.Get(r, "clusters").([]*dal.ClusterRow),
		currentCluster,
		tsLogsWithError.TSLogRows,
		savedQueriesWithError.SavedQueries,
		from,
		to,
	}

	var tmpl *template.Template

	currentUserPermission := currentCluster.GetPermissionByUserID(currentUser.ID)
	if currentUserPermission == "read" {
		tmpl, err = template.ParseFiles("templates/dashboard.html.tmpl", "templates/logs/list-readonly.html.tmpl")
	} else {
		tmpl, err = template.ParseFiles("templates/dashboard.html.tmpl", "templates/logs/list.html.tmpl")
	}
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	tmpl.Execute(w, data)
}

func GetLogsExecutors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	qParams := r.URL.Query()
	to := time.Now().UTC().Unix()
	from := to - 900

	var err error

	toString := qParams.Get("To")
	if toString == "" {
		toString = qParams.Get("to")
	}
	if toString != "" {
		to, err = strconv.ParseInt(toString, 10, 64)
		if err != nil {
			to = time.Now().UTC().Unix()
		}
	}

	fromString := qParams.Get("From")
	if fromString == "" {
		fromString = qParams.Get("from")
	}
	if fromString != "" {
		from, err = strconv.ParseInt(fromString, 10, 64)
		if err != nil {
			// default is 15 minutes range
			from = to - 900
		}
	}

	query := qParams.Get("q")

	currentUser := context.Get(r, "currentUser").(*dal.UserRow)

	currentCluster := context.Get(r, "currentCluster").(*dal.ClusterRow)

	db := context.Get(r, "db.Core").(*sqlx.DB)

	tsExecutorLogsDB := context.Get(r, "db.TSExecutorLog").(*sqlx.DB)

	// -----------------------------------
	// Create channels to receive SQL rows
	// -----------------------------------
	tsLogsChan := make(chan *dal.TSExecutorLogRowsWithError)
	defer close(tsLogsChan)

	savedQueriesChan := make(chan *dal.SavedQueryRowsWithError)
	defer close(savedQueriesChan)

	// --------------------------
	// Fetch SQL rows in parallel
	// --------------------------
	go func(currentCluster *dal.ClusterRow, from, to int64, query string) {
		deletedFrom := currentCluster.GetDeletedFromUNIXTimestampForSelect("ts_executor_logs")

		tsLogRowsWithError := &dal.TSExecutorLogRowsWithError{}
		tsLogRowsWithError.TSExecutorLogRows, tsLogRowsWithError.Error = dal.NewTSExecutorLog(tsExecutorLogsDB).AllByClusterIDRangeAndQuery(nil, currentCluster.ID, from, to, query, deletedFrom)
		tsLogRowsWithError.Error = err
		tsLogsChan <- tsLogRowsWithError
	}(currentCluster, from, to, query)

	go func(currentCluster *dal.ClusterRow) {
		savedQueriesWithError := &dal.SavedQueryRowsWithError{}
		savedQueriesWithError.SavedQueries, savedQueriesWithError.Error = dal.NewSavedQuery(db).AllByClusterIDAndType(nil, currentCluster.ID, "executor_logs")
		savedQueriesChan <- savedQueriesWithError
	}(currentCluster)

	// -----------------------------------
	// Wait for channels to return results
	// -----------------------------------
	tsLogsWithError := <-tsLogsChan
	if tsLogsWithError.Error != nil && tsLogsWithError.Error.Error() != "sql: no rows in result set" {
		libhttp.HandleErrorJson(w, tsLogsWithError.Error)
		return
	}

	savedQueriesWithError := <-savedQueriesChan
	if savedQueriesWithError.Error != nil && savedQueriesWithError.Error.Error() != "sql: no rows in result set" {
		libhttp.HandleErrorJson(w, savedQueriesWithError.Error)
		return
	}

	accessToken, err := getAccessToken(w, r, "read")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	data := struct {
		CSRFToken      string
		Addr           string
		CurrentUser    *dal.UserRow
		AccessToken    *dal.AccessTokenRow
		Clusters       []*dal.ClusterRow
		CurrentCluster *dal.ClusterRow
		Logs           []*dal.TSExecutorLogRow
		SavedQueries   []*dal.SavedQueryRow
		From           int64
		To             int64
	}{
		csrf.Token(r),
		context.Get(r, "addr").(string),
		currentUser,
		accessToken,
		context.Get(r, "clusters").([]*dal.ClusterRow),
		currentCluster,
		tsLogsWithError.TSExecutorLogRows,
		savedQueriesWithError.SavedQueries,
		from,
		to,
	}

	var tmpl *template.Template

	currentUserPermission := currentCluster.GetPermissionByUserID(currentUser.ID)
	if currentUserPermission == "read" {
		tmpl, err = template.ParseFiles("templates/dashboard.html.tmpl", "templates/logs/executor-list-readonly.html.tmpl")
	} else {
		tmpl, err = template.ParseFiles("templates/dashboard.html.tmpl", "templates/logs/executor-list.html.tmpl")
	}
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	tmpl.Execute(w, data)
}

func PostApiLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := context.Get(r, "db.Core").(*sqlx.DB)

	accessTokenRow := context.Get(r, "accessToken").(*dal.AccessTokenRow)

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	clusterRow, err := dal.NewCluster(db).GetByID(nil, accessTokenRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	deletedFrom := clusterRow.GetDeletedFromUNIXTimestampForInsert("ts_logs")

	tsLogsDB := context.Get(r, "db.TSLog").(*sqlx.DB)

	go func() {
		err = dal.NewTSLog(tsLogsDB).CreateFromJSON(nil, accessTokenRow.ClusterID, dataJson, deletedFrom)
		if err != nil {
			logrus.Error(err)
		}
	}()

	w.Write([]byte(`{"Message": "Success"}`))
}

func GetApiLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := context.Get(r, "db.Core").(*sqlx.DB)

	tsLogsDB := context.Get(r, "db.TSLog").(*sqlx.DB)

	accessTokenRow := context.Get(r, "accessToken").(*dal.AccessTokenRow)

	qParams := r.URL.Query()

	toString := qParams.Get("To")
	if toString == "" {
		toString = qParams.Get("to")
	}
	to, err := strconv.ParseInt(toString, 10, 64)

	fromString := qParams.Get("From")
	if fromString == "" {
		fromString = qParams.Get("from")
	}
	from, err := strconv.ParseInt(fromString, 10, 64)

	query := qParams.Get("q")

	clusterRow, err := dal.NewCluster(db).GetByID(nil, accessTokenRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	deletedFrom := clusterRow.GetDeletedFromUNIXTimestampForSelect("ts_logs")

	var tsLogs []*dal.TSLogRow

	if fromString == "" && toString == "" {
		tsLogs, err = dal.NewTSLog(tsLogsDB).AllByClusterIDLastRowIntervalAndQuery(nil, accessTokenRow.ClusterID, "15 minute", query, deletedFrom)
	} else {
		tsLogs, err = dal.NewTSLog(tsLogsDB).AllByClusterIDRangeAndQuery(nil, accessTokenRow.ClusterID, from, to, query, deletedFrom)
	}

	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	rowsJSON, err := json.Marshal(tsLogs)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(rowsJSON)
}

func GetApiLogsExecutors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := context.Get(r, "db.Core").(*sqlx.DB)

	tsExecutorLogsDB := context.Get(r, "db.TSExecutorLog").(*sqlx.DB)

	accessTokenRow := context.Get(r, "accessToken").(*dal.AccessTokenRow)

	qParams := r.URL.Query()

	toString := qParams.Get("To")
	if toString == "" {
		toString = qParams.Get("to")
	}
	to, err := strconv.ParseInt(toString, 10, 64)

	fromString := qParams.Get("From")
	if fromString == "" {
		fromString = qParams.Get("from")
	}
	from, err := strconv.ParseInt(fromString, 10, 64)

	query := qParams.Get("q")

	clusterRow, err := dal.NewCluster(db).GetByID(nil, accessTokenRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	deletedFrom := clusterRow.GetDeletedFromUNIXTimestampForSelect("ts_logs")

	var tsExecutorLogs []*dal.TSExecutorLogRow

	if fromString == "" && toString == "" {
		tsExecutorLogs, err = dal.NewTSExecutorLog(tsExecutorLogsDB).AllByClusterIDLastRowIntervalAndQuery(nil, accessTokenRow.ClusterID, "15 minute", query, deletedFrom)
	} else {
		tsExecutorLogs, err = dal.NewTSExecutorLog(tsExecutorLogsDB).AllByClusterIDRangeAndQuery(nil, accessTokenRow.ClusterID, from, to, query, deletedFrom)
	}

	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	rowsJSON, err := json.Marshal(tsExecutorLogs)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(rowsJSON)
}
