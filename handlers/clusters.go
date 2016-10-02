package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/mailer"
	"github.com/resourced/resourced-master/models/pg"
)

// GetClusters displays the /clusters UI.
func GetClusters(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	dbs := r.Context().Value("pg-dbs").(*config.PGDBConfig)

	currentUser := r.Context().Value("currentUser").(*pg.UserRow)

	clusters := r.Context().Value("clusters").([]*pg.ClusterRow)

	currentCluster := r.Context().Value("currentCluster").(*pg.ClusterRow)

	accessTokens := make(map[int64][]*pg.AccessTokenRow)

	for _, cluster := range clusters {
		accessTokensSlice, err := pg.NewAccessToken(dbs.Core).AllByClusterID(nil, cluster.ID)
		if err != nil {
			libhttp.HandleErrorHTML(w, err, 500)
			return
		}

		accessTokens[cluster.ID] = accessTokensSlice
	}

	data := struct {
		CSRFToken      string
		CurrentUser    *pg.UserRow
		Clusters       []*pg.ClusterRow
		CurrentCluster *pg.ClusterRow
		AccessTokens   map[int64][]*pg.AccessTokenRow
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
	dbs := r.Context().Value("pg-dbs").(*config.PGDBConfig)

	currentUser := r.Context().Value("currentUser").(*pg.UserRow)

	_, err := pg.NewCluster(dbs.Core).Create(nil, currentUser, r.FormValue("Name"))
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

// PostClusterIDCurrent sets a cluster to be the current one on the UI.
func PostClusterIDCurrent(w http.ResponseWriter, r *http.Request) {
	cookieStore := r.Context().Value("cookieStore").(*sessions.CookieStore)

	session, _ := cookieStore.Get(r, "resourcedmaster-session")

	clusterID, err := getInt64SlugFromPath(w, r, "clusterID")
	if err != nil {
		http.Redirect(w, r, r.Referer(), 301)
		return
	}

	clusterRows := r.Context().Value("clusters").([]*pg.ClusterRow)
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
	dbs := r.Context().Value("pg-dbs").(*config.PGDBConfig)

	clusterID, err := getInt64SlugFromPath(w, r, "clusterID")
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
		if dataRetentionValue < int64(1) {
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

	_, err = pg.NewCluster(dbs.Core).UpdateByID(nil, data, clusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

// DeleteClusterID deletes a cluster.
func DeleteClusterID(w http.ResponseWriter, r *http.Request) {
	dbs := r.Context().Value("pg-dbs").(*config.PGDBConfig)

	clusterID, err := getInt64SlugFromPath(w, r, "clusterID")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	currentUser := r.Context().Value("currentUser").(*pg.UserRow)

	cluster := pg.NewCluster(dbs.Core)

	clustersByUser, err := cluster.AllByUserID(nil, currentUser.ID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	if len(clustersByUser) <= 1 {
		libhttp.HandleErrorJson(w, errors.New("You only have 1 cluster. Thus, Delete is disallowed."))
		return
	}

	_, err = cluster.DeleteByID(nil, clusterID)
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
	dbs := r.Context().Value("pg-dbs").(*config.PGDBConfig)

	clusterID, err := getInt64SlugFromPath(w, r, "clusterID")
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

	userRow, err := pg.NewUser(dbs.Core).GetByEmail(nil, email)
	if err != nil && strings.Contains(err.Error(), "no rows in result set") {

		// 1. Create a user with temporary password
		userRow, err = pg.NewUser(dbs.Core).SignupRandomPassword(nil, email)
		if err != nil {
			libhttp.HandleErrorJson(w, err)
			return
		}

		// 2. Send email invite to user
		if userRow.EmailVerificationToken.String != "" {
			clusterRow, err := pg.NewCluster(dbs.Core).GetByID(nil, clusterID)
			if err != nil {
				libhttp.HandleErrorJson(w, err)
				return
			}

			mailer := r.Context().Value("mailer.GeneralConfig").(*mailer.Mailer)

			vipAddr := r.Context().Value("vipAddr").(string)
			vipProtocol := r.Context().Value("vipProtocol").(string)

			url := fmt.Sprintf("%v://%v/signup?email=%v&token=%v", vipProtocol, vipAddr, email, userRow.EmailVerificationToken.String)

			body := fmt.Sprintf(`ResourceD is a monitoring and alerting service.

Your coleague has invited you to join cluster: %v. Click the following link to signup:

%v`, clusterRow.Name, url)

			mailer.Send(email, fmt.Sprintf("You have been invited to join ResourceD on %v", vipAddr), body)
		}
	}

	// Add user as a member to this cluster
	err = pg.NewCluster(dbs.Core).UpdateMember(nil, clusterID, userRow, level, enabled)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

// DeleteClusterIDUsers removes user's membership from a particular cluster.
func DeleteClusterIDUsers(w http.ResponseWriter, r *http.Request) {
	dbs := r.Context().Value("pg-dbs").(*config.PGDBConfig)

	clusterID, err := getInt64SlugFromPath(w, r, "clusterID")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	email := r.FormValue("Email")

	existingUser, _ := pg.NewUser(dbs.Core).GetByEmail(nil, email)
	if existingUser != nil {
		err := pg.NewCluster(dbs.Core).RemoveMember(nil, clusterID, existingUser)
		if err != nil {
			libhttp.HandleErrorJson(w, err)
			return
		}
	}

	http.Redirect(w, r, r.Referer(), 301)
}
