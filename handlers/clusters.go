package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func GetClusters(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	db := context.Get(r, "db.Core").(*sqlx.DB)

	currentUser := getCurrentUser(w, r)

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
		CurrentUser        *dal.UserRow
		Clusters           []*dal.ClusterRow
		CurrentClusterJson string
		AccessTokens       map[int64][]*dal.AccessTokenRow
	}{
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

	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)

	session, _ := cookieStore.Get(r, "resourcedmaster-session")

	currentUser := session.Values["user"].(*dal.UserRow)

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

	redirectPath := "/"

	recentRequestPathInterface := session.Values["recentRequestPath"]
	if recentRequestPathInterface != nil {
		redirectPath = recentRequestPathInterface.(string)
	}

	clusterIDString := r.FormValue("ClusterID")
	clusterID, err := strconv.ParseInt(clusterIDString, 10, 64)
	if err != nil {
		http.Redirect(w, r, redirectPath, 301)
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

	http.Redirect(w, r, redirectPath, 301)
}
