package handlers

import (
	"errors"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func GetWatchers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)

	session, _ := cookieStore.Get(r, "resourcedmaster-session")
	currentUserRow, ok := session.Values["user"].(*dal.UserRow)
	if !ok {
		http.Redirect(w, r, "/logout", 301)
		return
	}

	currentClusterInterface := session.Values["currentCluster"]
	if currentClusterInterface == nil {
		http.Redirect(w, r, "/", 301)
		return
	}

	currentCluster := currentClusterInterface.(*dal.ClusterRow)

	db := context.Get(r, "db.Core").(*sqlx.DB)

	watchers, err := dal.NewWatcher(db).AllByClusterID(nil, currentCluster.ID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	savedQueries, err := dal.NewSavedQuery(db).AllByClusterID(nil, currentCluster.ID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	triggers, err := dal.NewWatcherTrigger(db).AllByClusterID(nil, currentCluster.ID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	triggersByWatcher := make(map[int64][]*dal.WatcherTriggerRow)
	for _, trigger := range triggers {
		if _, ok := triggersByWatcher[trigger.WatcherID]; !ok {
			triggersByWatcher[trigger.WatcherID] = make([]*dal.WatcherTriggerRow, 0)
		}

		triggersByWatcher[trigger.WatcherID] = append(triggersByWatcher[trigger.WatcherID], trigger)
	}

	data := struct {
		Addr               string
		CurrentUser        *dal.UserRow
		Clusters           []*dal.ClusterRow
		CurrentClusterJson string
		Watchers           []*dal.WatcherRow
		SavedQueries       []*dal.SavedQueryRow
		TriggersByWatcher  map[int64][]*dal.WatcherTriggerRow
	}{
		context.Get(r, "addr").(string),
		currentUserRow,
		context.Get(r, "clusters").([]*dal.ClusterRow),
		string(context.Get(r, "currentClusterJson").([]byte)),
		watchers,
		savedQueries,
		triggersByWatcher,
	}

	tmpl, err := template.ParseFiles("templates/dashboard.html.tmpl", "templates/watchers/list.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tmpl.Execute(w, data)
}

func readFormData(r *http.Request) (map[string]interface{}, error) {
	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)

	session, _ := cookieStore.Get(r, "resourcedmaster-session")

	currentClusterInterface := session.Values["currentCluster"]
	if currentClusterInterface == nil {
		return nil, errors.New("Current cluster is nil")
	}
	currentCluster := currentClusterInterface.(*dal.ClusterRow)

	savedQueryIDString := r.FormValue("SavedQueryID")
	savedQueryID, err := strconv.ParseInt(savedQueryIDString, 10, 64)
	if err != nil {
		return nil, err
	}

	savedQuery := r.FormValue("SavedQuery")

	name := r.FormValue("Name")
	if name == "" {
		name = savedQuery
	}

	lowAffectedHostsString := r.FormValue("LowAffectedHosts")
	lowAffectedHosts, err := strconv.ParseInt(lowAffectedHostsString, 10, 64)
	if err != nil {
		return nil, err
	}

	hostsLastUpdated := r.FormValue("HostsLastUpdated")
	checkInterval := r.FormValue("CheckInterval")

	// actionTransport := r.FormValue("ActionTransport")

	// actionEmail := r.FormValue("ActionEmail")
	// actionSMSCarrier := r.FormValue("ActionSMSCarrier")
	// actionSMSPhone := r.FormValue("ActionSMSPhone")
	// actionPagerDutyServiceKey := r.FormValue("ActionPagerDutyServiceKey")
	// actionPagerDutyDescription := r.FormValue("ActionPagerDutyDescription")

	// actions := make(map[string]interface{})
	// actions["Transport"] = actionTransport
	// actions["Email"] = actionEmail
	// actions["SMSCarrier"] = actionSMSCarrier
	// actions["SMSPhone"] = actionSMSPhone
	// actions["PagerDutyServiceKey"] = actionPagerDutyServiceKey
	// actions["PagerDutyDescription"] = actionPagerDutyDescription

	// actionsJson, err := json.Marshal(actions)
	// if err != nil {
	// 	return nil, err
	// }

	db := context.Get(r, "db.Core").(*sqlx.DB)

	return dal.NewWatcher(db).CreateOrUpdateParameters(
		currentCluster.ID, savedQueryID, savedQuery, name,
		lowAffectedHosts, hostsLastUpdated, checkInterval), nil
}

func PostWatchers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	createParams, err := readFormData(r)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	db := context.Get(r, "db.Core").(*sqlx.DB)

	_, err = dal.NewWatcher(db).Create(nil, createParams)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/watchers", 301)
}

func PostPutDeleteWatcherID(w http.ResponseWriter, r *http.Request) {
	method := r.FormValue("_method")
	if method == "" {
		method = "put"
	}

	if method == "post" || method == "put" {
		PutWatcherID(w, r)
	} else if method == "delete" {
		DeleteWatcherID(w, r)
	}
}

func PutWatcherID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	idString := vars["id"]
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	updateParams, err := readFormData(r)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	db := context.Get(r, "db.Core").(*sqlx.DB)

	_, err = dal.NewWatcher(db).UpdateByID(nil, updateParams, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/watchers", 301)
}

func DeleteWatcherID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	idString := vars["id"]
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	db := context.Get(r, "db.Core").(*sqlx.DB)

	_, err = dal.NewWatcher(db).DeleteByID(nil, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/watchers", 301)
}
