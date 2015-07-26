package handlers

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func GetMetadata(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)

	session, _ := cookieStore.Get(r, "resourcedmaster-session")
	currentUserRow, ok := session.Values["user"].(*dal.UserRow)
	if !ok {
		http.Redirect(w, r, "/logout", 301)
		return
	}

	db := context.Get(r, "db").(*sqlx.DB)

	currentClusterInterface := session.Values["currentCluster"]
	if currentClusterInterface == nil {
		http.Redirect(w, r, "/", 301)
		return
	}

	currentCluster := currentClusterInterface.(*dal.ClusterRow)

	metadataRows, err := dal.NewMetadata(db).AllByClusterID(nil, currentCluster.ID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	data := struct {
		CurrentUser        *dal.UserRow
		Clusters           []*dal.ClusterRow
		CurrentClusterJson string
		MetadataRows       []*dal.MetadataRow
	}{
		currentUserRow,
		context.Get(r, "clusters").([]*dal.ClusterRow),
		string(context.Get(r, "currentClusterJson").([]byte)),
		metadataRows,
	}

	tmpl, err := template.ParseFiles("templates/dashboard.html.tmpl", "templates/metadata/list.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tmpl.Execute(w, data)
}

func PostMetadata(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)

	session, _ := cookieStore.Get(r, "resourcedmaster-session")

	currentClusterInterface := session.Values["currentCluster"]
	if currentClusterInterface == nil {
		http.Redirect(w, r, "/", 301)
		return
	}
	currentCluster := currentClusterInterface.(*dal.ClusterRow)

	key := r.FormValue("Key")
	data := r.FormValue("Data")

	db := context.Get(r, "db").(*sqlx.DB)

	_, err := dal.NewMetadata(db).CreateOrUpdate(nil, currentCluster.ID, key, []byte(data))
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/metadata", 301)
}

func GetApiMetadata(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := context.Get(r, "db").(*sqlx.DB)

	accessTokenRow := context.Get(r, "accessTokenRow").(*dal.AccessTokenRow)

	metadataRows, err := dal.NewMetadata(db).AllByClusterID(nil, accessTokenRow.ClusterID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	metadataRowsJson, err := json.Marshal(metadataRows)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(metadataRowsJson)
}

func PostApiMetadataKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := context.Get(r, "db").(*sqlx.DB)

	accessTokenRow := context.Get(r, "accessTokenRow").(*dal.AccessTokenRow)

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	vars := mux.Vars(r)
	key := vars["key"]

	metadataRow, err := dal.NewMetadata(db).CreateOrUpdate(nil, accessTokenRow.ClusterID, key, dataJson)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	metadataRowJson, err := json.Marshal(metadataRow)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(metadataRowJson)
}

func GetApiMetadataKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := context.Get(r, "db").(*sqlx.DB)

	accessTokenRow := context.Get(r, "accessTokenRow").(*dal.AccessTokenRow)

	vars := mux.Vars(r)
	key := vars["key"]

	metadataRow, err := dal.NewMetadata(db).GetByClusterIDAndKey(nil, accessTokenRow.ClusterID, key)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	metadataRowJson, err := json.Marshal(metadataRow)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(metadataRowJson)
}
