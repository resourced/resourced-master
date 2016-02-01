package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func readTriggersFormData(r *http.Request) (map[string]interface{}, error) {
	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)

	session, _ := cookieStore.Get(r, "resourcedmaster-session")

	currentClusterInterface := session.Values["currentCluster"]
	if currentClusterInterface == nil {
		return nil, errors.New("Current cluster is nil")
	}
	currentCluster := currentClusterInterface.(*dal.ClusterRow)

	watcherIDString := mux.Vars(r)["watcherid"]
	if watcherIDString == "" {
		return nil, errors.New("id cannot be empty.")
	}

	watcherID, err := strconv.ParseInt(watcherIDString, 10, 64)
	if err != nil {
		return nil, errors.New("id cannot be non numeric.")
	}

	lowViolationsCountString := r.FormValue("LowViolationsCount")
	lowViolationsCount, err := strconv.ParseInt(lowViolationsCountString, 10, 64)
	if err != nil {
		return nil, err
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
			return nil, err
		}
	}

	createdInterval := r.FormValue("CreatedInterval")

	actionTransport := r.FormValue("ActionTransport")

	actionEmail := r.FormValue("ActionEmail")
	actionSMSCarrier := r.FormValue("ActionSMSCarrier")
	actionSMSPhone := r.FormValue("ActionSMSPhone")
	actionPagerDutyServiceKey := r.FormValue("ActionPagerDutyServiceKey")
	actionPagerDutyDescription := r.FormValue("ActionPagerDutyDescription")

	actions := make(map[string]interface{})
	actions["Transport"] = actionTransport
	actions["Email"] = actionEmail
	actions["SMSCarrier"] = actionSMSCarrier
	actions["SMSPhone"] = actionSMSPhone
	actions["PagerDutyServiceKey"] = actionPagerDutyServiceKey
	actions["PagerDutyDescription"] = actionPagerDutyDescription

	actionsJson, err := json.Marshal(actions)
	if err != nil {
		return nil, err
	}

	db := context.Get(r, "db.Core").(*sqlx.DB)

	return dal.NewWatcherTrigger(db).CreateOrUpdateParameters(currentCluster.ID, watcherID, lowViolationsCount, highViolationsCount, createdInterval, actionsJson), nil
}

func PostWatchersTriggers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	createParams, err := readTriggersFormData(r)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	db := context.Get(r, "db.Core").(*sqlx.DB)

	_, err = dal.NewWatcherTrigger(db).Create(nil, createParams)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

func PostPutDeleteWatcherTriggerID(w http.ResponseWriter, r *http.Request) {
	method := r.FormValue("_method")
	if method == "" {
		method = "put"
	}

	if method == "post" || method == "put" {
		PutWatcherTriggerID(w, r)
	} else if method == "delete" {
		DeleteWatcherTriggerID(w, r)
	}
}

func PutWatcherTriggerID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	updateParams, err := readTriggersFormData(r)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	db := context.Get(r, "db.Core").(*sqlx.DB)

	_, err = dal.NewWatcherTrigger(db).UpdateByID(nil, updateParams, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

func DeleteWatcherTriggerID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	db := context.Get(r, "db.Core").(*sqlx.DB)

	_, err = dal.NewWatcherTrigger(db).DeleteByID(nil, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}
