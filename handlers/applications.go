package handlers

import (
	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/context"
	"github.com/jmoiron/sqlx"
	resourcedmaster_dal "github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/libtemplate"
	"net/http"
)

func GetApplicationsCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	riceBoxes := context.Get(r, "riceBoxes").(map[string]*rice.Box)

	tmpl, err := libtemplate.GetFromRicebox(riceBoxes["templates"], false, "dashboard.html.tmpl", "applications/create.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tmpl.Execute(w, nil)
}

func PostApplications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	db := context.Get(r, "db").(*sqlx.DB)
	userRow := getCurrentUser(w, r)

	appName := r.FormValue("Name")

	_, err := resourcedmaster_dal.NewUser(db).CreateApplicationRow(nil, userRow.ID, appName)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/", 301)
}

func GetApplications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	db := context.Get(r, "db").(*sqlx.DB)
	userRow := getCurrentUser(w, r)

	_, err := resourcedmaster_dal.NewApplication(db).AllApplicationsByUserID(nil, userRow.ID)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	riceBoxes := context.Get(r, "riceBoxes").(map[string]*rice.Box)

	tmpl, err := libtemplate.GetFromRicebox(riceBoxes["templates"], false, "dashboard.html.tmpl", "applications/list.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tmpl.Execute(w, nil)
}
