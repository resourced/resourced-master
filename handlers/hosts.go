package handlers

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func GetHosts(w http.ResponseWriter, r *http.Request) {
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

	query := r.URL.Query().Get("q")

	hosts, err := dal.NewHost(db).AllByClusterIDAndQuery(nil, currentCluster.ID, query)
	if err != nil && err.Error() != "sql: no rows in result set" {
		libhttp.HandleErrorJson(w, err)
		return
	}

	savedQueries, err := dal.NewSavedQuery(db).AllByClusterID(nil, currentCluster.ID)
	if err != nil && err.Error() != "sql: no rows in result set" {
		libhttp.HandleErrorJson(w, err)
		return
	}

	accessTokenRow, err := dal.NewAccessToken(db).GetByUserID(nil, currentUserRow.ID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	metricsMap, err := dal.NewMetric(db).AllByClusterIDAsMap(nil, currentCluster.ID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	data := struct {
		Addr               string
		CurrentUser        *dal.UserRow
		AccessToken        *dal.AccessTokenRow
		Clusters           []*dal.ClusterRow
		CurrentClusterJson string
		Hosts              []*dal.HostRow
		SavedQueries       []*dal.SavedQueryRow
		MetricsMap         map[string]int64
	}{
		context.Get(r, "addr").(string),
		currentUserRow,
		accessTokenRow,
		context.Get(r, "clusters").([]*dal.ClusterRow),
		string(context.Get(r, "currentClusterJson").([]byte)),
		hosts,
		savedQueries,
		metricsMap,
	}

	tmpl, err := template.ParseFiles("templates/dashboard.html.tmpl", "templates/hosts/list.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tmpl.Execute(w, data)
}

func PostApiHosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := context.Get(r, "db").(*sqlx.DB)

	accessTokenRow := context.Get(r, "accessTokenRow").(*dal.AccessTokenRow)

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	hostRow, err := dal.NewHost(db).CreateOrUpdate(nil, accessTokenRow, dataJson)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	metricsMap, err := dal.NewMetric(db).AllByClusterIDAsMap(nil, hostRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tsMetric := dal.NewTSMetric(db)

	// Loop through every host's data and see if they are part of graph metrics.
	// If they are, insert a record in ts_metrics.
	for path, data := range hostRow.DataAsFlatKeyValue() {
		for dataKey, value := range data {
			metricKey := path + "." + dataKey

			if metricID, ok := metricsMap[metricKey]; ok {
				// Argh! Can't I use generic here?
				if trueValueInt64, ok := value.(int64); ok {
					err = tsMetric.Create(nil, hostRow.ClusterID, metricID, metricKey, hostRow.Name, float64(trueValueInt64))
					if err != nil {
						println(err.Error())
					}
				} else if trueValueInt, ok := value.(int); ok {
					err = tsMetric.Create(nil, hostRow.ClusterID, metricID, metricKey, hostRow.Name, float64(trueValueInt))
					if err != nil {
						println(err.Error())
					}
				} else if trueValueFloat64, ok := value.(float64); ok {
					err = tsMetric.Create(nil, hostRow.ClusterID, metricID, metricKey, hostRow.Name, trueValueFloat64)
					if err != nil {
						println(err.Error())
					}
				} else if trueValueFloat32, ok := value.(float32); ok {
					err = tsMetric.Create(nil, hostRow.ClusterID, metricID, metricKey, hostRow.Name, float64(trueValueFloat32))
					if err != nil {
						println(err.Error())
					}
				}
			}
		}
	}

	hostRowJson, err := json.Marshal(hostRow)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(hostRowJson)
}

func GetApiHosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := context.Get(r, "db").(*sqlx.DB)

	accessTokenRow := context.Get(r, "accessTokenRow").(*dal.AccessTokenRow)

	query := r.URL.Query().Get("q")

	hosts, err := dal.NewHost(db).AllByClusterIDAndQuery(nil, accessTokenRow.ID, query)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	hostRowsJson, err := json.Marshal(hosts)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(hostRowsJson)
}
