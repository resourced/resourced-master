package handlers

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/csrf"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/models/pg"
)

func GetLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	currentUser := r.Context().Value("currentUser").(*pg.UserRow)

	currentCluster := r.Context().Value("currentCluster").(*pg.ClusterRow)

	dbs := r.Context().Value("dbs").(*config.PGDBConfig)

	qParams := r.URL.Query()

	// Parse To/to get param
	toString := qParams.Get("To")
	if toString == "" {
		toString = qParams.Get("to")
	}

	// Parse From/from get param
	fromString := qParams.Get("From")
	if fromString == "" {
		fromString = qParams.Get("from")
	}

	// Fetch the last log row if any of the from/to are missing.
	var lastLogRow *pg.TSLogRow
	var err error

	if fromString == "" || toString == "" {
		lastLogRow, err = pg.NewTSLog(dbs.GetTSLog(currentCluster.ID)).LastByClusterID(nil, currentCluster.ID)
		if err != nil && err.Error() != "sql: no rows in result set" {
			libhttp.HandleErrorJson(w, err)
			return
		}
	}

	to, err := strconv.ParseInt(toString, 10, 64)
	if err != nil {
		to = lastLogRow.Created.Unix()
	}

	from, err := strconv.ParseInt(fromString, 10, 64)
	if err != nil {
		from = lastLogRow.Created.Add(-30 * time.Minute).Unix()
	}

	savedQueries, err := pg.NewSavedQuery(dbs.Core).AllByClusterIDAndType(nil, currentCluster.ID, "logs")
	if err != nil && err.Error() != "sql: no rows in result set" {
		libhttp.HandleErrorJson(w, err)
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
		CurrentUser    *pg.UserRow
		AccessToken    *pg.AccessTokenRow
		Clusters       []*pg.ClusterRow
		CurrentCluster *pg.ClusterRow
		SavedQueries   []*pg.SavedQueryRow
		From           int64
		To             int64
	}{
		csrf.Token(r),
		r.Context().Value("addr").(string),
		currentUser,
		accessToken,
		r.Context().Value("clusters").([]*pg.ClusterRow),
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

	dbs := r.Context().Value("dbs").(*config.PGDBConfig)

	accessTokenRow := r.Context().Value("accessToken").(*pg.AccessTokenRow)

	errLogger := r.Context().Value("errLogger").(*logrus.Logger)

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	clusterRow, err := pg.NewCluster(dbs.Core).GetByID(nil, accessTokenRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	deletedFrom := clusterRow.GetDeletedFromUNIXTimestampForInsert("ts_logs")

	go func() {
		err = pg.NewTSLog(dbs.GetTSLog(accessTokenRow.ClusterID)).CreateFromJSON(nil, accessTokenRow.ClusterID, dataJson, deletedFrom)
		if err != nil {
			errLogger.Error(err)
		}
	}()

	w.Write([]byte(`{"Message": "Success"}`))
}

func GetApiLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbs := r.Context().Value("dbs").(*config.PGDBConfig)

	accessTokenRow := r.Context().Value("accessToken").(*pg.AccessTokenRow)

	qParams := r.URL.Query()

	// Parse To/to get param
	toString := qParams.Get("To")
	if toString == "" {
		toString = qParams.Get("to")
	}

	// Parse From/from get param
	fromString := qParams.Get("From")
	if fromString == "" {
		fromString = qParams.Get("from")
	}

	// Fetch the last log row if any of the from/to are missing.
	var lastLogRow *pg.TSLogRow
	var err error

	if fromString == "" || toString == "" {
		lastLogRow, err = pg.NewTSLog(dbs.GetTSLog(accessTokenRow.ClusterID)).LastByClusterID(nil, accessTokenRow.ClusterID)
		if err != nil && err.Error() != "sql: no rows in result set" {
			libhttp.HandleErrorJson(w, err)
			return
		}
	}

	to, err := strconv.ParseInt(toString, 10, 64)
	if err != nil {
		to = lastLogRow.Created.Unix()
	}

	from, err := strconv.ParseInt(fromString, 10, 64)
	if err != nil {
		from = lastLogRow.Created.Add(-30 * time.Minute).Unix()
	}

	clusterRow, err := pg.NewCluster(dbs.Core).GetByID(nil, accessTokenRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	deletedFrom := clusterRow.GetDeletedFromUNIXTimestampForSelect("ts_logs")

	tsLogs, err := pg.NewTSLog(dbs.GetTSLog(accessTokenRow.ClusterID)).AllByClusterIDRangeAndQuery(
		nil,
		accessTokenRow.ClusterID,
		int64(math.Min(float64(from), float64(to))),
		int64(math.Max(float64(from), float64(to))),
		qParams.Get("q"),
		deletedFrom)

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
