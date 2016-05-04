package handlers

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/context"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"

	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/libslice"
)

func GetChecks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	db := context.Get(r, "db.Core").(*sqlx.DB)

	currentUser := context.Get(r, "currentUser").(*dal.UserRow)

	currentCluster := context.Get(r, "currentCluster").(*dal.ClusterRow)

	// -----------------------------------
	// Create channels to receive SQL rows
	// -----------------------------------
	checksChan := make(chan *dal.CheckRowsWithError)
	defer close(checksChan)

	metricsChan := make(chan *dal.MetricRowsWithError)
	defer close(metricsChan)

	accessTokenChan := make(chan *dal.AccessTokenRowWithError)
	defer close(accessTokenChan)

	// --------------------------
	// Fetch SQL rows in parallel
	// --------------------------
	go func(currentCluster *dal.ClusterRow) {
		checksWithError := &dal.CheckRowsWithError{}
		checksWithError.Checks, checksWithError.Error = dal.NewCheck(db).AllByClusterID(nil, currentCluster.ID)
		checksChan <- checksWithError
	}(currentCluster)

	go func(currentCluster *dal.ClusterRow) {
		metricsWithError := &dal.MetricRowsWithError{}
		metricsWithError.Metrics, metricsWithError.Error = dal.NewMetric(db).AllByClusterID(nil, currentCluster.ID)
		metricsChan <- metricsWithError
	}(currentCluster)

	go func(currentUser *dal.UserRow) {
		accessTokenWithError := &dal.AccessTokenRowWithError{}
		accessTokenWithError.AccessToken, accessTokenWithError.Error = dal.NewAccessToken(db).GetByUserID(nil, currentUser.ID)
		accessTokenChan <- accessTokenWithError
	}(currentUser)

	// -----------------------------------
	// Wait for channels to return results
	// -----------------------------------
	checksWithError := <-checksChan
	if checksWithError.Error != nil && checksWithError.Error.Error() != "sql: no rows in result set" {
		libhttp.HandleErrorJson(w, checksWithError.Error)
		return
	}

	metricsWithError := <-metricsChan
	if metricsWithError.Error != nil && metricsWithError.Error.Error() != "sql: no rows in result set" {
		libhttp.HandleErrorJson(w, metricsWithError.Error)
		return
	}

	accessTokenWithError := <-accessTokenChan
	if accessTokenWithError.Error != nil {
		libhttp.HandleErrorJson(w, accessTokenWithError.Error)
		return
	}

	data := struct {
		CSRFToken          string
		Addr               string
		CurrentUser        *dal.UserRow
		AccessToken        *dal.AccessTokenRow
		Clusters           []*dal.ClusterRow
		CurrentClusterJson string
		Checks             []*dal.CheckRow
		Metrics            []*dal.MetricRow
	}{
		csrf.Token(r),
		context.Get(r, "addr").(string),
		currentUser,
		accessTokenWithError.AccessToken,
		context.Get(r, "clusters").([]*dal.ClusterRow),
		string(context.Get(r, "currentClusterJson").([]byte)),
		checksWithError.Checks,
		metricsWithError.Metrics,
	}

	tmpl, err := template.ParseFiles("templates/dashboard.html.tmpl", "templates/checks/list.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tmpl.Execute(w, data)
}

func PostChecks(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "db.Core").(*sqlx.DB)

	w.Header().Set("Content-Type", "text/html")

	currentCluster := context.Get(r, "currentCluster").(*dal.ClusterRow)

	intervalInSeconds := r.FormValue("IntervalInSeconds")
	if intervalInSeconds == "" {
		intervalInSeconds = "60"
	}

	hostsListWithNewlines := r.FormValue("HostsList")
	hostsList := strings.Split(hostsListWithNewlines, "\n")
	hostsList = libslice.RemoveEmpty(hostsList)

	hostsListJSON, err := json.Marshal(hostsList)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	data := make(map[string]interface{})
	data["name"] = r.FormValue("Name")
	data["interval"] = intervalInSeconds + "s"
	data["hosts_query"] = r.FormValue("HostsQuery")
	data["hosts_list"] = hostsListJSON
	data["expressions"] = r.FormValue("Expressions")
	data["triggers"] = []byte("[]")
	data["last_result_hosts"] = []byte("[]")
	data["last_result_expressions"] = []byte("[]")

	_, err = dal.NewCheck(db).Create(nil, currentCluster.ID, data)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/checks", 301)
}

func PostPutDeleteCheckID(w http.ResponseWriter, r *http.Request) {
	method := r.FormValue("_method")
	if method == "" {
		method = "put"
	}

	if method == "post" || method == "put" {
		PutCheckID(w, r)
	} else if method == "delete" {
		DeleteCheckID(w, r)
	}
}

func PutCheckID(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "db.Core").(*sqlx.DB)

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	intervalInSeconds := r.FormValue("IntervalInSeconds")
	if intervalInSeconds == "" {
		intervalInSeconds = "60"
	}

	hostsListWithNewlines := r.FormValue("HostsList")
	hostsList := strings.Split(hostsListWithNewlines, "\n")
	hostsList = libslice.RemoveEmpty(hostsList)

	hostsListJSON, err := json.Marshal(hostsList)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	data := make(map[string]interface{})
	data["name"] = r.FormValue("Name")
	data["interval"] = intervalInSeconds + "s"
	data["hosts_query"] = r.FormValue("HostsQuery")
	data["hosts_list"] = hostsListJSON
	data["expressions"] = r.FormValue("Expressions")

	_, err = dal.NewCheck(db).UpdateByID(nil, data, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

func DeleteCheckID(w http.ResponseWriter, r *http.Request) {
	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	db := context.Get(r, "db.Core").(*sqlx.DB)

	currentCluster := context.Get(r, "currentCluster").(*dal.ClusterRow)

	_, err = dal.NewCheck(db).DeleteByClusterIDAndID(nil, currentCluster.ID, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

func PostCheckIDSilence(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "db.Core").(*sqlx.DB)

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	check := dal.NewCheck(db)

	checkRow, err := check.GetByID(nil, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	data := make(map[string]interface{})
	data["is_silenced"] = !checkRow.IsSilenced

	_, err = check.UpdateByID(nil, data, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

func newCheckTriggerFromForm(r *http.Request) (dal.CheckTrigger, error) {
	lowViolationsCountString := r.FormValue("LowViolationsCount")
	lowViolationsCount, err := strconv.ParseInt(lowViolationsCountString, 10, 64)
	if err != nil {
		return dal.CheckTrigger{}, err
	}

	var highViolationsCount int64

	highViolationsCountString := r.FormValue("HighViolationsCount")
	if highViolationsCountString == "" {
		// Set highViolationsCount arbitrarily high when highViolationsCount value is missing.
		// Because it means that the user does not want to set max value.
		highViolationsCount = 1000000

	} else {
		highViolationsCount, err = strconv.ParseInt(highViolationsCountString, 10, 64)
		if err != nil {
			return dal.CheckTrigger{}, err
		}
	}

	action := dal.CheckTriggerAction{}
	action.Transport = r.FormValue("ActionTransport")
	action.Email = r.FormValue("ActionEmail")
	action.SMSCarrier = r.FormValue("ActionSMSCarrier")
	action.SMSPhone = r.FormValue("ActionSMSPhone")
	action.PagerDutyServiceKey = r.FormValue("ActionPagerDutyServiceKey")
	action.PagerDutyDescription = r.FormValue("ActionPagerDutyDescription")

	trigger := dal.CheckTrigger{}
	trigger.LowViolationsCount = lowViolationsCount
	trigger.HighViolationsCount = highViolationsCount
	trigger.CreatedInterval = r.FormValue("CreatedInterval")
	trigger.Action = action

	return trigger, nil
}

func PostChecksTriggers(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "db.Core").(*sqlx.DB)

	w.Header().Set("Content-Type", "text/html")

	checkIDString := mux.Vars(r)["checkid"]
	if checkIDString == "" {
		libhttp.HandleErrorJson(w, errors.New("id cannot be empty."))
		return
	}

	checkID, err := strconv.ParseInt(checkIDString, 10, 64)
	if err != nil {
		libhttp.HandleErrorJson(w, errors.New("id cannot be non numeric."))
		return
	}

	trigger, err := newCheckTriggerFromForm(r)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	check := dal.NewCheck(db)

	trigger.ID = check.NewExplicitID()

	checkRow, err := check.GetByID(nil, checkID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	_, err = check.AddTrigger(nil, checkRow, trigger)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

func PostPutDeleteCheckTriggerID(w http.ResponseWriter, r *http.Request) {
	method := r.FormValue("_method")
	if method == "" {
		method = "put"
	}

	if method == "post" || method == "put" {
		PutCheckTriggerID(w, r)
	} else if method == "delete" {
		DeleteCheckTriggerID(w, r)
	}
}

func PutCheckTriggerID(w http.ResponseWriter, r *http.Request) {
	checkIDString := mux.Vars(r)["checkid"]
	if checkIDString == "" {
		libhttp.HandleErrorJson(w, errors.New("id cannot be empty."))
		return
	}

	checkID, err := strconv.ParseInt(checkIDString, 10, 64)
	if err != nil {
		libhttp.HandleErrorJson(w, errors.New("id cannot be non numeric."))
		return
	}

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	trigger, err := newCheckTriggerFromForm(r)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	trigger.ID = id

	db := context.Get(r, "db.Core").(*sqlx.DB)

	check := dal.NewCheck(db)

	checkRow, err := check.GetByID(nil, checkID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	_, err = check.UpdateTrigger(nil, checkRow, trigger)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

func DeleteCheckTriggerID(w http.ResponseWriter, r *http.Request) {
	checkIDString := mux.Vars(r)["checkid"]
	if checkIDString == "" {
		libhttp.HandleErrorJson(w, errors.New("id cannot be empty."))
		return
	}

	checkID, err := strconv.ParseInt(checkIDString, 10, 64)
	if err != nil {
		libhttp.HandleErrorJson(w, errors.New("id cannot be non numeric."))
		return
	}

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	trigger := dal.CheckTrigger{}
	trigger.ID = id

	db := context.Get(r, "db.Core").(*sqlx.DB)

	check := dal.NewCheck(db)

	checkRow, err := check.GetByID(nil, checkID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	_, err = check.DeleteTrigger(nil, checkRow, trigger)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}
