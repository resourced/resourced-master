package handlers

import (
	"errors"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/context"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func GetClusters(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	db := context.Get(r, "db.Core").(*sqlx.DB)

	currentUser := context.Get(r, "currentUser").(*dal.UserRow)

	clusters := context.Get(r, "clusters").([]*dal.ClusterRow)

	accessTokens := make(map[int64][]*dal.AccessTokenRow)

	for _, cluster := range clusters {
		accessTokensSlice, err := dal.NewAccessToken(db).AllAccessTokensByClusterID(nil, cluster.ID)
		if err != nil {
			libhttp.HandleErrorHTML(w, err, 500)
			return
		}

		accessTokens[cluster.ID] = accessTokensSlice
	}

	data := struct {
		CSRFToken          string
		CurrentUser        *dal.UserRow
		Clusters           []*dal.ClusterRow
		CurrentClusterJson string
		AccessTokens       map[int64][]*dal.AccessTokenRow
	}{
		csrf.Token(r),
		currentUser,
		clusters,
		string(context.Get(r, "currentClusterJson").([]byte)),
		accessTokens,
	}

	tmpl, err := template.ParseFiles("templates/dashboard.html.tmpl", "templates/clusters/list.html.tmpl")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	tmpl.Execute(w, data)
}

func PostClusters(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "db.Core").(*sqlx.DB)

	currentUser := context.Get(r, "currentUser").(*dal.UserRow)

	_, err := dal.NewCluster(db).Create(nil, currentUser.ID, r.FormValue("Name"))
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	http.Redirect(w, r, "/clusters", 301)
}

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

	http.Redirect(w, r, r.Referer(), 301)
}

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

func PutClusterID(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "db.Core").(*sqlx.DB)

	vars := mux.Vars(r)

	idString := vars["id"]
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	name := r.FormValue("Name")

	data := make(map[string]interface{})
	data["name"] = name

	_, err = dal.NewCluster(db).UpdateByID(nil, data, id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/clusters", 301)
}

func DeleteClusterID(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "db.Core").(*sqlx.DB)

	vars := mux.Vars(r)

	idString := vars["id"]
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	cluster := dal.NewCluster(db)

	clustersByUser, err := cluster.AllClustersByUserID(nil, id)
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

	http.Redirect(w, r, "/clusters", 301)
}
