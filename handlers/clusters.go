package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/context"
	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/mailer"
)

// GetClusters displays the /clusters UI.
func GetClusters(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	dbs := context.Get(r, "dbs").(*config.DBConfig)

	currentUser := context.Get(r, "currentUser").(*dal.UserRow)

	clusters := context.Get(r, "clusters").([]*dal.ClusterRow)

	currentCluster := context.Get(r, "currentCluster").(*dal.ClusterRow)

	accessTokens := make(map[int64][]*dal.AccessTokenRow)

	for _, cluster := range clusters {
		accessTokensSlice, err := dal.NewAccessToken(dbs.Core).AllByClusterID(nil, cluster.ID)
		if err != nil {
			libhttp.HandleErrorHTML(w, err, 500)
			return
		}

		accessTokens[cluster.ID] = accessTokensSlice
	}

	data := struct {
		CSRFToken      string
		CurrentUser    *dal.UserRow
		Clusters       []*dal.ClusterRow
		CurrentCluster *dal.ClusterRow
		AccessTokens   map[int64][]*dal.AccessTokenRow
	}{
		csrf.Token(r),
		currentUser,
		clusters,
		currentCluster,
		accessTokens,
	}

	var tmpl *template.Template
	var err error

	currentUserPermission := currentCluster.GetLevelByUserID(currentUser.ID)
	if currentUserPermission == "read" {
		tmpl, err = template.ParseFiles("templates/dashboard.html.tmpl", "templates/clusters/list-readonly.html.tmpl")
	} else {
		tmpl, err = template.ParseFiles("templates/dashboard.html.tmpl", "templates/clusters/list.html.tmpl")
	}
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	tmpl.Execute(w, data)
}

// PostClusters creates a new cluster.
func PostClusters(w http.ResponseWriter, r *http.Request) {
	dbs := context.Get(r, "dbs").(*config.DBConfig)

	currentUser := context.Get(r, "currentUser").(*dal.UserRow)

	_, err := dal.NewCluster(dbs.Core).Create(nil, currentUser, r.FormValue("Name"))
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

// PostClustersCurrent sets a cluster to be the current one on the UI.
func PostClustersCurrent(w http.ResponseWriter, r *http.Request) {
	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)

	session, _ := cookieStore.Get(r, "resourcedmaster-session")

	clusterIDString := r.FormValue("ClusterID")
	clusterID, err := strconv.ParseInt(clusterIDString, 10, 64)
	if err != nil {
		http.Redirect(w, r, r.Referer(), 301)
		return
	}

	clusterRows := context.Get(r, "clusters").([]*dal.ClusterRow)
	for _, clusterRow := range clusterRows {
		if clusterRow.ID == clusterID {
			session.Values["currentCluster"] = clusterRow

			err := session.Save(r, w)
			if err != nil {
				libhttp.HandleErrorJson(w, err)
				return
			}
			break
		}
	}

	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		w.Write([]byte(""))
	} else {
		http.Redirect(w, r, r.Referer(), 301)
	}
}

// PostPutDeleteClusterIDUsers is a subrouter that modify cluster information.
func PostPutDeleteClusterID(w http.ResponseWriter, r *http.Request) {
	method := r.FormValue("_method")
	if method == "" {
		method = "put"
	}

	if method == "post" || method == "put" {
		PutClusterID(w, r)
	} else if method == "delete" {
		DeleteClusterID(w, r)
	}
}

// PutClusterID updates a cluster's information.
func PutClusterID(w http.ResponseWriter, r *http.Request) {
	dbs := context.Get(r, "dbs").(*config.DBConfig)

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	name := r.FormValue("Name")

	dataRetention := make(map[string]int)
	for _, table := range []string{"ts_checks", "ts_events", "ts_executor_logs", "ts_logs", "ts_metrics", "ts_metrics_aggr_15m"} {
		dataRetentionValue, err := strconv.ParseInt(r.FormValue("Table:"+table), 10, 64)
		if err != nil {
			dataRetentionValue = int64(1)
		}
		dataRetention[table] = int(dataRetentionValue)
	}

	dataRetentionJSON, err := json.Marshal(dataRetention)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	data := make(map[string]interface{})
	data["name"] = name
	data["data_retention"] = dataRetentionJSON

	_, err = dal.NewCluster(dbs.Core).UpdateByID(nil, data, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

// DeleteClusterID deletes a cluster.
func DeleteClusterID(w http.ResponseWriter, r *http.Request) {
	dbs := context.Get(r, "dbs").(*config.DBConfig)

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	cluster := dal.NewCluster(dbs.Core)

	clustersByUser, err := cluster.AllByUserID(nil, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	if len(clustersByUser) <= 1 {
		libhttp.HandleErrorJson(w, errors.New("You only have 1 cluster. Thus, Delete is disallowed."))
		return
	}

	_, err = cluster.DeleteByID(nil, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

// PostPutDeleteClusterIDUsers is a subrouter that handles user membership to a cluster.
func PostPutDeleteClusterIDUsers(w http.ResponseWriter, r *http.Request) {
	method := r.FormValue("_method")
	if method == "" {
		method = "put"
	}

	if method == "post" || method == "put" {
		PutClusterIDUsers(w, r)
	} else if method == "delete" {
		DeleteClusterIDUsers(w, r)
	}
}

// PutClusterIDUsers adds a user as a member to a particular cluster.
func PutClusterIDUsers(w http.ResponseWriter, r *http.Request) {
	dbs := context.Get(r, "dbs").(*config.DBConfig)

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	email := r.FormValue("Email")
	level := r.FormValue("Level")
	enabled := false

	if r.FormValue("Enabled") == "on" {
		enabled = true
	}

	userRow, err := dal.NewUser(dbs.Core).GetByEmail(nil, email)
	if err != nil && strings.Contains(err.Error(), "no rows in result set") {

		// 1. Create a user with temporary password
		userRow, err = dal.NewUser(dbs.Core).SignupRandomPassword(nil, email)
		if err != nil {
			libhttp.HandleErrorJson(w, err)
			return
		}

		// 2. Send email invite to user
		if userRow.EmailVerificationToken.String != "" {
			clusterRow, err := dal.NewCluster(dbs.Core).GetByID(nil, id)
			if err != nil {
				libhttp.HandleErrorJson(w, err)
				return
			}

			mailer := context.Get(r, "mailer.GeneralConfig").(*mailer.Mailer)

			vipAddr := context.Get(r, "vipAddr").(string)
			vipProtocol := context.Get(r, "vipProtocol").(string)

			url := fmt.Sprintf("%v://%v/signup?email=%v&token=%v", vipProtocol, vipAddr, email, userRow.EmailVerificationToken.String)

			body := fmt.Sprintf(`ResourceD is a monitoring and alerting service.

Your coleague has invited you to join cluster: %v. Click the following link to signup:

%v`, clusterRow.Name, url)

			mailer.Send(email, fmt.Sprintf("You have been invited to join ResourceD on %v", vipAddr), body)
		}
	}

	// Add user as a member to this cluster
	err = dal.NewCluster(dbs.Core).UpdateMember(nil, id, userRow, level, enabled)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

// DeleteClusterIDUsers removes user's membership from a particular cluster.
func DeleteClusterIDUsers(w http.ResponseWriter, r *http.Request) {
	dbs := context.Get(r, "dbs").(*config.DBConfig)

	id, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	email := r.FormValue("Email")

	existingUser, _ := dal.NewUser(dbs.Core).GetByEmail(nil, email)
	if existingUser != nil {
		err := dal.NewCluster(dbs.Core).RemoveMember(nil, id, existingUser)
		if err != nil {
			libhttp.HandleErrorJson(w, err)
			return
		}
	}

	http.Redirect(w, r, r.Referer(), 301)
}
