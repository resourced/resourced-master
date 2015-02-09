package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/context"
	gorilla_mux "github.com/gorilla/mux"
	resourcedmaster_dao "github.com/resourced/resourced-master/dao"
	"github.com/resourced/resourced-master/libhttp"
	resourcedmaster_storage "github.com/resourced/resourced-master/storage"
	"io/ioutil"
	"net/http"
)

//
// Admin level access
//

// PostApiUser
func PostApiUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	user, err := resourcedmaster_dao.NewUserGivenJson(store, r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	err = user.Save()
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	userJson, err := json.Marshal(user)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(userJson)
}

func GetApiUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	users, err := resourcedmaster_dao.AllUsers(store)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	usersJson, err := json.Marshal(users)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(usersJson)
}

func GetApiUserName(w http.ResponseWriter, r *http.Request) {
	params := gorilla_mux.Vars(r)

	w.Header().Set("Content-Type", "application/json")

	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	user, err := resourcedmaster_dao.GetUserByName(store, params["name"])
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	userJson, err := json.Marshal(user)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(userJson)
}

func PutApiUserName(w http.ResponseWriter, r *http.Request) {
	params := gorilla_mux.Vars(r)

	w.Header().Set("Content-Type", "application/json")

	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	currentUser := context.Get(r, "currentUser").(*resourcedmaster_dao.User)

	allowLevelUpdate := false

	if currentUser != nil && currentUser.Level == "staff" {
		allowLevelUpdate = true
	}

	user, err := resourcedmaster_dao.UpdateUserByNameGivenJson(store, params["name"], allowLevelUpdate, r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	userJson, err := json.Marshal(user)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(userJson)
}

func DeleteApiUserName(w http.ResponseWriter, r *http.Request) {
	params := gorilla_mux.Vars(r)

	w.Header().Set("Content-Type", "application/json")

	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	err := resourcedmaster_dao.DeleteUserByName(store, params["name"])
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	messageJson, err := json.Marshal(
		map[string]string{
			"Message": fmt.Sprintf("User{Name: %v} is deleted.", params["name"])})

	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(messageJson)
}

func PutApiUserNameAccessToken(w http.ResponseWriter, r *http.Request) {
	params := gorilla_mux.Vars(r)

	w.Header().Set("Content-Type", "application/json")

	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	user, err := resourcedmaster_dao.UpdateUserTokenByName(store, params["name"])
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	userJson, err := json.Marshal(user)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(userJson)
}

func PostApiApplicationIdAccessToken(w http.ResponseWriter, r *http.Request) {
	params := gorilla_mux.Vars(r)

	w.Header().Set("Content-Type", "application/json")

	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	app, err := resourcedmaster_dao.GetApplicationById(store, params["id"])
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	user, err := resourcedmaster_dao.NewAccessTokenUser(store, app)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	err = user.Save()
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	userJson, err := json.Marshal(user)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(userJson)
}

func DeleteApiApplicationIdAccessToken(w http.ResponseWriter, r *http.Request) {
	params := gorilla_mux.Vars(r)

	w.Header().Set("Content-Type", "application/json")

	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	err := resourcedmaster_dao.DeleteUserByName(store, params["token"])
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	err = resourcedmaster_dao.DeleteApplicationByAccessToken(store, params["token"])
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	messageJson, err := json.Marshal(
		map[string]string{
			"Message": fmt.Sprintf("AccessToken{Token: %v} is deleted.", params["token"])})

	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(messageJson)
}

//
// Basic level access
//

func GetRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/api", 301)
}

func GetApi(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	currentUser := context.Get(r, "currentUser").(*resourcedmaster_dao.User)

	if currentUser.Level == "staff" {
		http.Redirect(w, r, "/api/app", 301)

	} else {
		if currentUser.ApplicationId == "" {
			libhttp.HandleErrorJson(w, errors.New("User does not belong to application."))
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/api/app/%v/hosts", currentUser.ApplicationId), 301)
	}
}

func GetApiApp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	currentUser := context.Get(r, "currentUser").(*resourcedmaster_dao.User)

	if currentUser.Level != "staff" {
		err := errors.New("Access level is too low.")
		libhttp.HandleErrorJson(w, err)
		return
	}

	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	applications, err := resourcedmaster_dao.AllApplications(store)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	applicationsJson, err := json.Marshal(applications)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(applicationsJson)
}

// **GET** `/api/app/:id/hosts` Displays list of all hosts.
func GetApiAppIdHosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := gorilla_mux.Vars(r)
	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	hosts, err := resourcedmaster_dao.AllHosts(store, params["id"])
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	hostsJson, err := json.Marshal(hosts)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(hostsJson)
}

func hostAndDataPayloadJson(store resourcedmaster_storage.Storer, appId string, host *resourcedmaster_dao.Host) ([]byte, error) {
	payload := make(map[string]interface{})
	var payloadJson []byte

	payload["Host"] = host

	hostData, err := resourcedmaster_dao.AllApplicationDataByHost(store, appId, host.Name)
	if err != nil {
		return payloadJson, err
	}

	payload["Data"] = hostData

	payloadJson, err = json.Marshal(payload)
	if err != nil {
		return payloadJson, err
	}

	return payloadJson, nil
}

// **GET** `/api/app/:id/hosts/:name` Displays host data.
func GetApiAppIdHostsName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := gorilla_mux.Vars(r)
	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	host, err := resourcedmaster_dao.GetHostByAppId(store, params["id"], params["name"])
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	payloadJson, err := hostAndDataPayloadJson(store, params["id"], host)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(payloadJson)
}

func PostApiAppIdHostsName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := gorilla_mux.Vars(r)
	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	dataJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	var data map[string]interface{}

	if err := json.Unmarshal(dataJson, &data); err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	if _, ok := data["Host"]; !ok {
		err = errors.New("Data does not contain Host information.")
		libhttp.HandleErrorJson(w, err)
		return
	}

	app, err := resourcedmaster_dao.GetApplicationById(store, params["id"])
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	hostData := data["Host"].(map[string]interface{})
	hostname := hostData["Name"].(string)

	host := resourcedmaster_dao.NewHost(store, hostname, app.Id)

	// TODO(didip): Enabling these 2 cause panic.
	// Panic: interface conversion: interface is []interface {}, not []string
	// hostTags := hostData["Tags"].([]string)
	// hostNetInterfaces := hostData["NetworkInterfaces"].(map[string]map[string]interface{})

	err = host.Save()
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	var messageJson []byte

	if _, ok := data["Path"]; ok {
		path := data["Path"].(string)

		err = app.SaveDataJson(hostname, path, dataJson)
		if err != nil {
			libhttp.HandleErrorJson(w, err)
			return
		}

		messageJson, err = json.Marshal(
			map[string]string{
				"Message": fmt.Sprintf("Data{Path: %v} is saved.", params["path"])})
	} else {
		messageJson, err = json.Marshal(
			map[string]string{
				"Message": "Data is saved."})
	}

	w.Write(messageJson)
}
