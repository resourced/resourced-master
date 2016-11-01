package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/csrf"
	"github.com/pressly/chi"

	"github.com/resourced/resourced-master/contexthelper"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/libslice"
	"github.com/resourced/resourced-master/messagebus"
	"github.com/resourced/resourced-master/models/pg"
)

func GetChecks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	currentUser := r.Context().Value("currentUser").(*pg.UserRow)

	currentCluster := r.Context().Value("currentCluster").(*pg.ClusterRow)

	accessToken, err := getAccessToken(w, r, "read")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	// -----------------------------------
	// Create channels to receive SQL rows
	// -----------------------------------
	checksChan := make(chan *pg.CheckRowsWithError)
	defer close(checksChan)

	metricsChan := make(chan *pg.MetricRowsWithError)
	defer close(metricsChan)

	// --------------------------
	// Fetch SQL rows in parallel
	// --------------------------
	go func(currentCluster *pg.ClusterRow) {
		checksWithError := &pg.CheckRowsWithError{}
		checksWithError.Checks, checksWithError.Error = pg.NewCheck(r.Context()).AllByClusterID(nil, currentCluster.ID)
		checksChan <- checksWithError
	}(currentCluster)

	go func(currentCluster *pg.ClusterRow) {
		metricsWithError := &pg.MetricRowsWithError{}
		metricsWithError.Metrics, metricsWithError.Error = pg.NewMetric(r.Context()).AllByClusterID(nil, currentCluster.ID)
		metricsChan <- metricsWithError
	}(currentCluster)

	// -----------------------------------
	// Wait for channels to return results
	// -----------------------------------
	hasError := false

	checksWithError := <-checksChan
	if checksWithError.Error != nil && checksWithError.Error.Error() != "sql: no rows in result set" {
		libhttp.HandleErrorHTML(w, checksWithError.Error, 500)
		hasError = true
	}

	metricsWithError := <-metricsChan
	if metricsWithError.Error != nil && metricsWithError.Error.Error() != "sql: no rows in result set" {
		libhttp.HandleErrorHTML(w, metricsWithError.Error, 500)
		hasError = true
	}

	if hasError {
		return
	}

	data := struct {
		CSRFToken      string
		Addr           string
		CurrentUser    *pg.UserRow
		AccessToken    *pg.AccessTokenRow
		Clusters       []*pg.ClusterRow
		CurrentCluster *pg.ClusterRow
		Checks         []*pg.CheckRow
		Metrics        []*pg.MetricRow
	}{
		csrf.Token(r),
		r.Context().Value("Addr").(string),
		currentUser,
		accessToken,
		r.Context().Value("clusters").([]*pg.ClusterRow),
		currentCluster,
		checksWithError.Checks,
		metricsWithError.Metrics,
	}

	var tmpl *template.Template

	currentUserPermission := currentCluster.GetLevelByUserID(currentUser.ID)
	if currentUserPermission == "read" {
		tmpl, err = template.ParseFiles("templates/dashboard.html.tmpl", "templates/checks/list-readonly.html.tmpl")
	} else {
		tmpl, err = template.ParseFiles("templates/dashboard.html.tmpl", "templates/checks/list.html.tmpl")
	}
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	tmpl.Execute(w, data)
}

func PostChecks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	currentCluster := r.Context().Value("currentCluster").(*pg.ClusterRow)

	errLogger, err := contexthelper.GetLogger(r.Context(), "ErrLogger")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
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
		libhttp.HandleErrorHTML(w, err, 500)
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

	_, err = pg.NewCheck(r.Context()).Create(nil, currentCluster.ID, data)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	bus := r.Context().Value("bus").(*messagebus.MessageBus)
	go func() {
		err := bus.Publish("checks-refetch", "true")
		if err != nil {
			errLogger.WithFields(logrus.Fields{"Error": err}).Error("Failed to publish checks-refetch message to message bus")
		}
	}()

	http.Redirect(w, r, "/checks", 301)
}

func PostPutDeleteCheckID(w http.ResponseWriter, r *http.Request) {
	errLogger, err := contexthelper.GetLogger(r.Context(), "ErrLogger")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	method := r.FormValue("_method")
	if method == "" {
		method = "put"
	}

	bus := r.Context().Value("bus").(*messagebus.MessageBus)
	go func() {
		err := bus.Publish("checks-refetch", "true")
		if err != nil {
			errLogger.WithFields(logrus.Fields{"Error": err}).Error("Failed to publish checks-refetch message to message bus")
		}
	}()

	if method == "post" || method == "put" {
		PutCheckID(w, r)
	} else if method == "delete" {
		DeleteCheckID(w, r)
	}
}

func PutCheckID(w http.ResponseWriter, r *http.Request) {
	id, err := getInt64SlugFromPath(w, r, "checkID")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
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
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	data := make(map[string]interface{})
	data["name"] = r.FormValue("Name")
	data["interval"] = intervalInSeconds + "s"
	data["hosts_query"] = r.FormValue("HostsQuery")
	data["hosts_list"] = hostsListJSON
	data["expressions"] = r.FormValue("Expressions")

	_, err = pg.NewCheck(r.Context()).UpdateByID(nil, data, id)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

func DeleteCheckID(w http.ResponseWriter, r *http.Request) {
	id, err := getInt64SlugFromPath(w, r, "checkID")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	currentCluster := r.Context().Value("currentCluster").(*pg.ClusterRow)

	_, err = pg.NewCheck(r.Context()).DeleteByClusterIDAndID(nil, currentCluster.ID, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

func PostCheckIDSilence(w http.ResponseWriter, r *http.Request) {
	id, err := getInt64SlugFromPath(w, r, "checkID")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	check := pg.NewCheck(r.Context())

	checkRow, err := check.GetByID(nil, id)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	data := make(map[string]interface{})
	data["is_silenced"] = !checkRow.IsSilenced

	_, err = check.UpdateByID(nil, data, id)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

func newCheckTriggerFromForm(r *http.Request) (pg.CheckTrigger, error) {
	lowViolationsCountString := r.FormValue("LowViolationsCount")
	lowViolationsCount, err := strconv.ParseInt(lowViolationsCountString, 10, 64)
	if err != nil {
		return pg.CheckTrigger{}, err
	}

	// Set highViolationsCount arbitrarily high by default.
	// Because it means that the user does not want to set max value.
	highViolationsCount := int64(1000000)
	highViolationsCountString := r.FormValue("HighViolationsCount")
	if highViolationsCountString != "" {
		highViolationsCount, err = strconv.ParseInt(highViolationsCountString, 10, 64)
		if err != nil {
			return pg.CheckTrigger{}, err
		}
	}

	createdIntervalMinute := int64(1)
	createdIntervalMinuteString := r.FormValue("CreatedIntervalMinute")
	if createdIntervalMinuteString != "" {
		createdIntervalMinute, err = strconv.ParseInt(createdIntervalMinuteString, 10, 64)
		if err != nil {
			return pg.CheckTrigger{}, err
		}
	}

	action := pg.CheckTriggerAction{}
	action.Transport = r.FormValue("ActionTransport")
	action.Email = r.FormValue("ActionEmail")
	action.SMSCarrier = r.FormValue("ActionSMSCarrier")
	action.SMSPhone = r.FormValue("ActionSMSPhone")
	action.PagerDutyServiceKey = r.FormValue("ActionPagerDutyServiceKey")
	action.PagerDutyDescription = r.FormValue("ActionPagerDutyDescription")

	trigger := pg.CheckTrigger{}
	trigger.LowViolationsCount = lowViolationsCount
	trigger.HighViolationsCount = highViolationsCount
	trigger.CreatedIntervalMinute = createdIntervalMinute
	trigger.Action = action

	return trigger, nil
}

func PostChecksTriggers(w http.ResponseWriter, r *http.Request) {
	errLogger, err := contexthelper.GetLogger(r.Context(), "ErrLogger")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	checkIDString := chi.URLParam(r, "checkID")
	if checkIDString == "" {
		libhttp.HandleErrorHTML(w, fmt.Errorf("id cannot be empty."), 500)
		return
	}

	checkID, err := strconv.ParseInt(checkIDString, 10, 64)
	if err != nil {
		libhttp.HandleErrorHTML(w, fmt.Errorf("id cannot be non numeric."), 500)
		return
	}

	trigger, err := newCheckTriggerFromForm(r)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	check := pg.NewCheck(r.Context())

	trigger.ID = pg.NewExplicitID()

	checkRow, err := check.GetByID(nil, checkID)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	_, err = check.AddTrigger(nil, checkRow, trigger)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	bus := r.Context().Value("bus").(*messagebus.MessageBus)
	go func() {
		err := bus.Publish("checks-refetch", "true")
		if err != nil {
			errLogger.WithFields(logrus.Fields{"Error": err}).Error("Failed to publish checks-refetch message to message bus")
		}
	}()

	http.Redirect(w, r, r.Referer(), 301)
}

func PostPutDeleteCheckTriggerID(w http.ResponseWriter, r *http.Request) {
	errLogger, err := contexthelper.GetLogger(r.Context(), "ErrLogger")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	method := r.FormValue("_method")
	if method == "" {
		method = "put"
	}

	bus := r.Context().Value("bus").(*messagebus.MessageBus)
	go func() {
		err := bus.Publish("checks-refetch", "true")
		if err != nil {
			errLogger.WithFields(logrus.Fields{"Error": err}).Error("Failed to publish checks-refetch message to message bus")
		}
	}()

	if method == "post" || method == "put" {
		PutCheckTriggerID(w, r)
	} else if method == "delete" {
		DeleteCheckTriggerID(w, r)
	}
}

func PutCheckTriggerID(w http.ResponseWriter, r *http.Request) {
	checkIDString := chi.URLParam(r, "checkID")
	if checkIDString == "" {
		libhttp.HandleErrorJson(w, fmt.Errorf("checkID cannot be empty."))
		return
	}

	checkID, err := strconv.ParseInt(checkIDString, 10, 64)
	if err != nil {
		libhttp.HandleErrorJson(w, fmt.Errorf("checkID cannot be non numeric."))
		return
	}

	triggerID, err := getInt64SlugFromPath(w, r, "triggerID")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	trigger, err := newCheckTriggerFromForm(r)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	trigger.ID = triggerID

	check := pg.NewCheck(r.Context())

	checkRow, err := check.GetByID(nil, checkID)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	_, err = check.UpdateTrigger(nil, checkRow, trigger)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

func DeleteCheckTriggerID(w http.ResponseWriter, r *http.Request) {
	checkIDString := chi.URLParam(r, "checkID")
	if checkIDString == "" {
		libhttp.HandleErrorJson(w, fmt.Errorf("checkID cannot be empty."))
		return
	}

	checkID, err := strconv.ParseInt(checkIDString, 10, 64)
	if err != nil {
		libhttp.HandleErrorJson(w, fmt.Errorf("checkID cannot be non numeric."))
		return
	}

	triggerID, err := getInt64SlugFromPath(w, r, "triggerID")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	trigger := pg.CheckTrigger{}
	trigger.ID = triggerID

	check := pg.NewCheck(r.Context())

	checkRow, err := check.GetByID(nil, checkID)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	_, err = check.DeleteTrigger(nil, checkRow, trigger)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

func GetApiCheckIDResults(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accessTokenRow := r.Context().Value("accessToken").(*pg.AccessTokenRow)

	qParams := r.URL.Query()

	limitString := qParams.Get("Limit")
	if limitString == "" {
		limitString = qParams.Get("limit")
	}
	limit, err := strconv.ParseInt(limitString, 10, 64)
	if err != nil || limit <= 0 {
		limit = 10
	}

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	checkRow, err := pg.NewCheck(r.Context()).GetByID(nil, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	if accessTokenRow.ClusterID != checkRow.ClusterID {
		libhttp.HandleErrorJson(w, fmt.Errorf("No permission to access check with ID: %v", id))
		return
	}

	tsCheckRows, err := pg.NewTSCheck(r.Context(), checkRow.ClusterID).LastByClusterIDCheckIDAndLimit(nil, checkRow.ClusterID, checkRow.ID, limit)

	tsCheckRowsJSON, err := json.Marshal(tsCheckRows)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(tsCheckRowsJSON)
}
