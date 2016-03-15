package handlers

import (
	"io/ioutil"
	"net/http"

	"github.com/gorilla/context"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/multidb"
)

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
