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

	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/mailer"
	"github.com/resourced/resourced-master/models/pg"
	"github.com/resourced/resourced-master/models/shared"
)

// GetClusters displays the /clusters UI.
func GetClusters(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	currentUser := r.Context().Value("currentUser").(*shared.UserRow)

	clusters := r.Context().Value("clusters").([]*pg.ClusterRow)

	currentCluster := r.Context().Value("currentCluster").(*pg.ClusterRow)

	accessTokens := make(map[int64][]*pg.AccessTokenRow)

	for _, cluster := range clusters {
		accessTokensSlice, err := pg.NewAccessToken(r.Context()).AllByClusterID(nil, cluster.ID)
		if err != nil {
			libhttp.HandleErrorHTML(w, err, 500)
			return
		}

		accessTokens[cluster.ID] = accessTokensSlice
	}

	data := struct {
		CSRFToken      string
		CurrentUser    *shared.UserRow
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
	currentUser := r.Context().Value("currentUser").(*shared.UserRow)

	_, err := pg.NewCluster(r.Context()).Create(nil, currentUser, r.FormValue("Name"))
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

// PostClusterIDCurrent sets a cluster to be the current one on the UI.
func PostClusterIDCurrent(w http.ResponseWriter, r *http.Request) {
	cookieStore := r.Context().Value("CookieStore").(*sessions.CookieStore)

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
	clusterID, err := getInt64SlugFromPath(w, r, "clusterID")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	name := r.FormValue("Name")

	dataRetention := make(map[string]int)
	for _, table := range []string{"ts_checks", "ts_events", "ts_executor_logs", "ts_logs", "ts_metrics"} {
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

	_, err = pg.NewCluster(r.Context()).UpdateByID(nil, data, clusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

// DeleteClusterID deletes a cluster.
func DeleteClusterID(w http.ResponseWriter, r *http.Request) {
	clusterID, err := getInt64SlugFromPath(w, r, "clusterID")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	currentUser := r.Context().Value("currentUser").(*shared.UserRow)

	cluster := pg.NewCluster(r.Context())

	clustersByUser, err := cluster.AllByUserID(nil, currentUser.ID)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	if len(clustersByUser) <= 1 {
		libhttp.HandleErrorJson(w, errors.New("You only have 1 cluster. Thus, Delete is disallowed."))
		return
	}

	_, err = cluster.DeleteByID(nil, clusterID)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
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

	userRow, err := pg.NewUser(r.Context()).GetByEmail(nil, email)
	if err != nil && strings.Contains(err.Error(), "no rows in result set") {

		// 1. Create a user with temporary password
		userRow, err = pg.NewUser(r.Context()).SignupRandomPassword(nil, email)
		if err != nil {
			libhttp.HandleErrorHTML(w, err, 500)
			return
		}

		// 2. Send email invite to user
		if userRow.EmailVerificationToken != "" {
			clusterRow, err := pg.NewCluster(r.Context()).GetByID(nil, clusterID)
			if err != nil {
				libhttp.HandleErrorHTML(w, err, 500)
				return
			}

			mailer := r.Context().Value("mailer.GeneralConfig").(*mailer.Mailer)

			vipAddr := r.Context().Value("vipAddr").(string)
			vipProtocol := r.Context().Value("vipProtocol").(string)

			url := fmt.Sprintf("%v://%v/signup?email=%v&token=%v", vipProtocol, vipAddr, email, userRow.EmailVerificationToken)

			body := fmt.Sprintf(`ResourceD is a monitoring and alerting service.

Your coleague has invited you to join cluster: %v. Click the following link to signup:

%v`, clusterRow.Name, url)

			mailer.Send(email, fmt.Sprintf("You have been invited to join ResourceD on %v", vipAddr), body)
		}
	}

	// Add user as a member to this cluster
	err = pg.NewCluster(r.Context()).UpdateMember(nil, clusterID, userRow, level, enabled)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, r.Referer(), 301)
}

// DeleteClusterIDUsers removes user's membership from a particular cluster.
func DeleteClusterIDUsers(w http.ResponseWriter, r *http.Request) {
	clusterID, err := getInt64SlugFromPath(w, r, "clusterID")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	email := r.FormValue("Email")

	existingUser, _ := pg.NewUser(r.Context()).GetByEmail(nil, email)
	if existingUser != nil {
		err := pg.NewCluster(r.Context()).RemoveMember(nil, clusterID, existingUser)
		if err != nil {
			libhttp.HandleErrorHTML(w, err, 500)
			return
		}
	}

	http.Redirect(w, r, r.Referer(), 301)
}
