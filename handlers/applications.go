package handlers

import (
	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/context"
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
