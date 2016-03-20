package handlers

import (
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/csrf"

	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/multidb"
)

func GetLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	currentUser := context.Get(r, "currentUser").(*dal.UserRow)

	data := struct {
		CSRFToken          string
		Addr               string
		CurrentUser        *dal.UserRow
		Clusters           []*dal.ClusterRow
		CurrentClusterJson string
	}{
		csrf.Token(r),
		context.Get(r, "addr").(string),
		currentUser,
		context.Get(r, "clusters").([]*dal.ClusterRow),
		string(context.Get(r, "currentClusterJson").([]byte)),
	}

	tmpl, err := template.ParseFiles("templates/dashboard.html.tmpl", "templates/logs/list.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tmpl.Execute(w, data)
}

func PostApiLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accessTokenRow := context.Get(r, "accessTokenRow").(*dal.AccessTokenRow)

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	dbs := context.Get(r, "multidb.TSLogs").(*multidb.MultiDB).PickMultipleForWrites()
	for _, db := range dbs {
		err = dal.NewTSLog(db).CreateFromJSON(nil, accessTokenRow.ClusterID, dataJson)
		if err != nil {
			libhttp.HandleErrorJson(w, err)
			return
		}
	}

	w.Write([]byte(`{"Message": "Success"}`))
}
