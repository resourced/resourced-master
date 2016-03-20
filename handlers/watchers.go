package handlers

import (
	"encoding/base64"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/context"
	"github.com/gorilla/csrf"
	"github.com/jmoiron/sqlx"

	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func GetWatchers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	db := context.Get(r, "db.Core").(*sqlx.DB)

	currentUser := context.Get(r, "currentUser").(*dal.UserRow)

	currentCluster := context.Get(r, "currentCluster").(*dal.ClusterRow)

	// -----------------------------------
	// Create channels to receive SQL rows
	// -----------------------------------
	watchersChan := make(chan *dal.WatcherRowsWithError)
	defer close(watchersChan)

	savedQueriesChan := make(chan *dal.SavedQueryRowsWithError)
	defer close(savedQueriesChan)

	triggersChan := make(chan *dal.WatcherTriggerRowsWithError)
	defer close(triggersChan)

	// --------------------------
	// Fetch SQL rows in parallel
	// --------------------------
	go func(currentCluster *dal.ClusterRow) {
		watchersWithError := &dal.WatcherRowsWithError{}
		watchersWithError.Watchers, watchersWithError.Error = dal.NewWatcher(db).AllPassiveByClusterID(nil, currentCluster.ID)
		watchersChan <- watchersWithError
	}(currentCluster)

	go func(currentCluster *dal.ClusterRow) {
		savedQueriesWithError := &dal.SavedQueryRowsWithError{}
		savedQueriesWithError.SavedQueries, savedQueriesWithError.Error = dal.NewSavedQuery(db).AllByClusterID(nil, currentCluster.ID)
		savedQueriesChan <- savedQueriesWithError
	}(currentCluster)

	go func(currentCluster *dal.ClusterRow) {
		triggersWithError := &dal.WatcherTriggerRowsWithError{}
		triggersWithError.Triggers, triggersWithError.Error = dal.NewWatcherTrigger(db).AllByClusterID(nil, currentCluster.ID)
		triggersChan <- triggersWithError
	}(currentCluster)

	// -----------------------------------
	// Wait for channels to return results
	// -----------------------------------
	watchersWithError := <-watchersChan
	if watchersWithError.Error != nil && watchersWithError.Error.Error() != "sql: no rows in result set" {
		libhttp.HandleErrorJson(w, watchersWithError.Error)
		return
	}

	savedQueriesWithError := <-savedQueriesChan
	if savedQueriesWithError.Error != nil && savedQueriesWithError.Error.Error() != "sql: no rows in result set" {
		libhttp.HandleErrorJson(w, savedQueriesWithError.Error)
		return
	}

	triggersWithError := <-triggersChan
	if triggersWithError.Error != nil && triggersWithError.Error.Error() != "sql: no rows in result set" {
		libhttp.HandleErrorJson(w, triggersWithError.Error)
		return
	}

	triggersByWatcher := make(map[int64][]*dal.WatcherTriggerRow)
	for _, trigger := range triggersWithError.Triggers {
		if _, ok := triggersByWatcher[trigger.WatcherID]; !ok {
			triggersByWatcher[trigger.WatcherID] = make([]*dal.WatcherTriggerRow, 0)
		}

		triggersByWatcher[trigger.WatcherID] = append(triggersByWatcher[trigger.WatcherID], trigger)
	}

	data := struct {
		CSRFToken          string
		Addr               string
		CurrentUser        *dal.UserRow
		Clusters           []*dal.ClusterRow
		CurrentClusterJson string
		Watchers           []*dal.WatcherRow
		SavedQueries       []*dal.SavedQueryRow
		TriggersByWatcher  map[int64][]*dal.WatcherTriggerRow
	}{
		csrf.Token(r),
		context.Get(r, "addr").(string),
		currentUser,
		context.Get(r, "clusters").([]*dal.ClusterRow),
		string(context.Get(r, "currentClusterJson").([]byte)),
		watchersWithError.Watchers,
		savedQueriesWithError.SavedQueries,
		triggersByWatcher,
	}

	tmpl, err := template.ParseFiles("templates/dashboard.html.tmpl", "templates/watchers/list-passive.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tmpl.Execute(w, data)
}

func GetWatchersActive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	db := context.Get(r, "db.Core").(*sqlx.DB)

	currentUser := context.Get(r, "currentUser").(*dal.UserRow)

	currentCluster := context.Get(r, "currentCluster").(*dal.ClusterRow)

	// -----------------------------------
	// Create channels to receive SQL rows
	// -----------------------------------
	watchersChan := make(chan *dal.WatcherRowsWithError)
	defer close(watchersChan)

	triggersChan := make(chan *dal.WatcherTriggerRowsWithError)
	defer close(triggersChan)

	// --------------------------
	// Fetch SQL rows in parallel
	// --------------------------
	go func(currentCluster *dal.ClusterRow) {
		watchersWithError := &dal.WatcherRowsWithError{}
		watchersWithError.Watchers, watchersWithError.Error = dal.NewWatcher(db).AllActiveByClusterID(nil, currentCluster.ID)
		watchersChan <- watchersWithError
	}(currentCluster)

	go func(currentCluster *dal.ClusterRow) {
		triggersWithError := &dal.WatcherTriggerRowsWithError{}
		triggersWithError.Triggers, triggersWithError.Error = dal.NewWatcherTrigger(db).AllByClusterID(nil, currentCluster.ID)
		triggersChan <- triggersWithError
	}(currentCluster)

	// -----------------------------------
	// Wait for channels to return results
	// -----------------------------------
	watchersWithError := <-watchersChan
	if watchersWithError.Error != nil && watchersWithError.Error.Error() != "sql: no rows in result set" {
		libhttp.HandleErrorJson(w, watchersWithError.Error)
		return
	}

	triggersWithError := <-triggersChan
	if triggersWithError.Error != nil && triggersWithError.Error.Error() != "sql: no rows in result set" {
		libhttp.HandleErrorJson(w, triggersWithError.Error)
		return
	}

	triggersByWatcher := make(map[int64][]*dal.WatcherTriggerRow)
	for _, trigger := range triggersWithError.Triggers {
		if _, ok := triggersByWatcher[trigger.WatcherID]; !ok {
			triggersByWatcher[trigger.WatcherID] = make([]*dal.WatcherTriggerRow, 0)
		}

		triggersByWatcher[trigger.WatcherID] = append(triggersByWatcher[trigger.WatcherID], trigger)
	}

	data := struct {
		CSRFToken          string
		Addr               string
		CurrentUser        *dal.UserRow
		Clusters           []*dal.ClusterRow
		CurrentClusterJson string
		Watchers           []*dal.WatcherRow
		TriggersByWatcher  map[int64][]*dal.WatcherTriggerRow
	}{
		csrf.Token(r),
		context.Get(r, "addr").(string),
		currentUser,
		context.Get(r, "clusters").([]*dal.ClusterRow),
		string(context.Get(r, "currentClusterJson").([]byte)),
		watchersWithError.Watchers,
		triggersByWatcher,
	}

	tmpl, err := template.ParseFiles("templates/dashboard.html.tmpl", "templates/watchers/list-active.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tmpl.Execute(w, data)
}

func watcherPassiveFormData(r *http.Request) (map[string]interface{}, error) {
	currentCluster := context.Get(r, "currentCluster").(*dal.ClusterRow)

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

	db := context.Get(r, "db.Core").(*sqlx.DB)

	return dal.NewWatcher(db).CreateOrUpdateParameters(
		currentCluster.ID, savedQuery, name,
		lowAffectedHosts, hostsLastUpdated, checkInterval, nil), nil
}

func watcherActiveFormData(r *http.Request) (map[string]interface{}, error) {
	currentCluster := context.Get(r, "currentCluster").(*dal.ClusterRow)

	data := make(map[string]interface{})
	data["Command"] = r.FormValue("Command")
	data["SSHUser"] = r.FormValue("SSHUser")
	data["SSHPort"] = r.FormValue("SSHPort")
	data["HTTPHeaders"] = r.FormValue("HTTPHeaders")
	data["HTTPScheme"] = r.FormValue("HTTPScheme")
	data["HTTPPort"] = r.FormValue("HTTPPort")
	data["HTTPPath"] = r.FormValue("HTTPPath")
	data["HTTPMethod"] = r.FormValue("HTTPMethod")
	data["HTTPPostBody"] = r.FormValue("HTTPPostBody")
	data["HTTPUser"] = r.FormValue("HTTPUser")
	data["HostsList"] = r.FormValue("HostsList")

	if r.FormValue("HTTPPass") != "" {
		data["HTTPPass"] = base64.StdEncoding.EncodeToString([]byte(r.FormValue("HTTPPass")))
	}
	if r.FormValue("HTTPCode") != "" {
		httpCode, err := strconv.ParseInt(r.FormValue("HTTPCode"), 10, 64)
		if err != nil {
			return nil, err
		}
		data["HTTPCode"] = httpCode
	}

	name := r.FormValue("Name")
	if name == "" {
		name = data["Command"].(string)
	}

	lowAffectedHostsString := r.FormValue("LowAffectedHosts")
	lowAffectedHosts, err := strconv.ParseInt(lowAffectedHostsString, 10, 64)
	if err != nil {
		return nil, err
	}

	hostsLastUpdated := r.FormValue("HostsLastUpdated")
	checkInterval := r.FormValue("CheckInterval")

	db := context.Get(r, "db.Core").(*sqlx.DB)

	return dal.NewWatcher(db).CreateOrUpdateParameters(
		currentCluster.ID, "", name,
		lowAffectedHosts, hostsLastUpdated, checkInterval, data), nil
}

func PostWatchersActive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	createParams, err := watcherActiveFormData(r)
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

	http.Redirect(w, r, "/watchers/active", 301)
}

func PostWatchers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	createParams, err := watcherPassiveFormData(r)
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
	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	updateParams, err := watcherPassiveFormData(r)
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

	http.Redirect(w, r, r.Referer(), 301)
}

func DeleteWatcherID(w http.ResponseWriter, r *http.Request) {
	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	db := context.Get(r, "db.Core").(*sqlx.DB)

	currentCluster := context.Get(r, "currentCluster").(*dal.ClusterRow)

	_, err = dal.NewWatcher(db).DeleteByClusterIDAndID(nil, currentCluster.ID, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

func PostPutDeleteWatcherActiveID(w http.ResponseWriter, r *http.Request) {
	method := r.FormValue("_method")
	if method == "" {
		method = "put"
	}

	if method == "post" || method == "put" {
		PutWatcherActiveID(w, r)
	} else if method == "delete" {
		DeleteWatcherID(w, r)
	}
}

func PutWatcherActiveID(w http.ResponseWriter, r *http.Request) {
	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	updateParams, err := watcherActiveFormData(r)
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

	http.Redirect(w, r, r.Referer(), 301)
}

func PostWatcherIDSilence(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "db.Core").(*sqlx.DB)

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	watcher := dal.NewWatcher(db)

	watcherRow, err := watcher.GetByID(nil, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	data := make(map[string]interface{})
	data["is_silenced"] = !watcherRow.IsSilenced

	_, err = watcher.UpdateByID(nil, data, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}
