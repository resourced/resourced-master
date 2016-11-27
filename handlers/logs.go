package handlers

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"

	"github.com/gorilla/csrf"

	"github.com/resourced/resourced-master/contexthelper"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/models/cassandra"
	"github.com/resourced/resourced-master/models/shared"
	"github.com/resourced/resourced-master/models/shims"
)

func GetLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	currentUser := r.Context().Value("currentUser").(*cassandra.UserRow)

	currentCluster := r.Context().Value("currentCluster").(*cassandra.ClusterRow)

	qParams := r.URL.Query()

	toString := qParams.Get("to")
	fromString := qParams.Get("from")

	// Fetch the last log row if any of the from/to are missing.
	var lastLogRow shared.ICreatedUnix
	var err error

	if fromString == "" || toString == "" {
		lastLogRow, err = shims.NewTSLog(r.Context(), currentCluster.ID).LastByClusterID(currentCluster.ID)
		if err != nil && err.Error() != "sql: no rows in result set" {
			libhttp.HandleErrorHTML(w, err, 500)
			return
		}
	}

	to, err := strconv.ParseInt(toString, 10, 64)
	if err != nil {
		to = lastLogRow.CreatedUnix()
	}

	from, err := strconv.ParseInt(fromString, 10, 64)
	if err != nil {
		from = to - 1800 // 30 minutes
	}

	savedQueries, err := cassandra.NewSavedQuery(r.Context()).AllByClusterIDAndType(currentCluster.ID, "logs")
	if err != nil && err.Error() != "sql: no rows in result set" {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	accessToken, err := getAccessToken(w, r, "read")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	data := struct {
		CSRFToken      string
		Addr           string
		CurrentUser    *cassandra.UserRow
		AccessToken    *cassandra.AccessTokenRow
		Clusters       []*cassandra.ClusterRow
		CurrentCluster *cassandra.ClusterRow
		SavedQueries   []*cassandra.SavedQueryRow
		From           int64
		To             int64
	}{
		csrf.Token(r),
		r.Context().Value("Addr").(string),
		currentUser,
		accessToken,
		r.Context().Value("clusters").([]*cassandra.ClusterRow),
		currentCluster,
		savedQueries,
		from,
		to,
	}

	var tmpl *template.Template

	currentUserPermission := currentCluster.GetLevelByUserID(currentUser.ID)
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

func PostApiLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accessTokenRow := r.Context().Value("accessToken").(*cassandra.AccessTokenRow)

	errLogger, err := contexthelper.GetLogger(r.Context(), "ErrLogger")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	clusterRow, err := cassandra.NewCluster(r.Context()).GetByID(accessTokenRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	shimsTSLog := shims.NewTSLog(r.Context(), accessTokenRow.ClusterID)

	go func() {
		err = shimsTSLog.CreateFromJSON(accessTokenRow.ClusterID, dataJson, clusterRow.GetDeletedFromUNIXTimestampForInsert("ts_logs"), clusterRow.GetTTLDurationForInsert("ts_logs"))
		if err != nil {
			errLogger.Error(err)
		}
	}()

	w.Write([]byte(`{"Message": "Success"}`))
}

func GetApiLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accessTokenRow := r.Context().Value("accessToken").(*cassandra.AccessTokenRow)

	qParams := r.URL.Query()

	toString := qParams.Get("to")
	fromString := qParams.Get("from")

	// Fetch the last log row if any of the from/to are missing.
	var lastLogRow shared.ICreatedUnix
	var err error

	tsLog := shims.NewTSLog(r.Context(), accessTokenRow.ClusterID)

	if fromString == "" || toString == "" {
		lastLogRow, err = tsLog.LastByClusterID(accessTokenRow.ID)
		if err != nil && err.Error() != "sql: no rows in result set" {
			libhttp.HandleErrorJson(w, err)
			return
		}
	}

	to, err := strconv.ParseInt(toString, 10, 64)
	if err != nil {
		to = lastLogRow.CreatedUnix()
	}

	from, err := strconv.ParseInt(fromString, 10, 64)
	if err != nil {
		from = to - 1800 // 30 minutes
	}

	clusterRow, err := cassandra.NewCluster(r.Context()).GetByID(accessTokenRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tsLogs, err := tsLog.AllByClusterIDRangeAndQuery(
		accessTokenRow.ClusterID,
		int64(math.Min(float64(from), float64(to))),
		int64(math.Max(float64(from), float64(to))),
		qParams.Get("q"),
		clusterRow.GetDeletedFromUNIXTimestampForSelect("ts_logs"))

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
