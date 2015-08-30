package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/context"
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

	db := context.Get(r, "db").(*sqlx.DB)

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

	data := struct {
		Addr               string
		CurrentUser        *dal.UserRow
		Clusters           []*dal.ClusterRow
		CurrentClusterJson string
		Watchers           []*dal.WatcherRow
		SavedQueries       []*dal.SavedQueryRow
	}{
		context.Get(r, "addr").(string),
		currentUserRow,
		context.Get(r, "clusters").([]*dal.ClusterRow),
		string(context.Get(r, "currentClusterJson").([]byte)),
		watchers,
		savedQueries,
	}

	tmpl, err := template.ParseFiles("templates/dashboard.html.tmpl", "templates/watchers/list.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tmpl.Execute(w, data)
}

func PostWatchers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)

	session, _ := cookieStore.Get(r, "resourcedmaster-session")

	currentClusterInterface := session.Values["currentCluster"]
	if currentClusterInterface == nil {
		http.Redirect(w, r, "/", 301)
		return
	}
	currentCluster := currentClusterInterface.(*dal.ClusterRow)

	savedQueryIDString := r.FormValue("SavedQueryID")
	savedQueryID, err := strconv.ParseInt(savedQueryIDString, 10, 64)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	savedQuery := r.FormValue("SavedQuery")

	lowThresholdString := r.FormValue("LowThreshold")
	lowThreshold, err := strconv.ParseInt(lowThresholdString, 10, 64)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	highThresholdString := r.FormValue("HighThreshold")
	highThreshold, err := strconv.ParseInt(highThresholdString, 10, 64)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	lowAffectedHostsString := r.FormValue("LowAffectedHosts")
	lowAffectedHosts, err := strconv.ParseInt(lowAffectedHostsString, 10, 64)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	hostsLastUpdated := r.FormValue("HostsLastUpdated")
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
		libhttp.HandleErrorJson(w, err)
		return
	}

	db := context.Get(r, "db").(*sqlx.DB)

	_, err = dal.NewWatcher(db).Create(nil, currentCluster.ID, savedQueryID, savedQuery, lowThreshold, highThreshold, lowAffectedHosts, hostsLastUpdated, "", actionsJson)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/watchers", 301)
}
