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
	params := gorilla_mux.Vars(r)

	w.Header().Set("Content-Type", "application/json")

	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	user, err := resourcedmaster_dao.NewUserGivenJson(store, r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}
	user.ApplicationId = params["id"]

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
	params := gorilla_mux.Vars(r)

	w.Header().Set("Content-Type", "application/json")

	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	users, err := resourcedmaster_dao.AllUsersByAppId(store, params["id"])
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

	user, err := resourcedmaster_dao.UpdateUserByNameGivenJson(store, params["name"], r.Body)
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
	params := gorilla_mux.Vars(r)

	w.Header().Set("Content-Type", "application/json")

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

// **GET** `/api/app/:id/hosts/hardware-addr/:address` Displays list of hosts by MAC-48/EUI-48/EUI-64 address.
func GetApiAppIdHostsHardwareAddr(w http.ResponseWriter, r *http.Request) {
	params := gorilla_mux.Vars(r)

	w.Header().Set("Content-Type", "application/json")

	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	host, err := resourcedmaster_dao.GetHostByAppIdAndHardwareAddr(store, params["id"], params["address"])
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	hostJson, err := json.Marshal(host)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(hostJson)
}

// **GET** `/api/app/:id/hosts/ip-addr/:address` Displays list of hosts by IP address.
func GetApiAppIdHostsIpAddr(w http.ResponseWriter, r *http.Request) {
	params := gorilla_mux.Vars(r)

	w.Header().Set("Content-Type", "application/json")

	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	host, err := resourcedmaster_dao.GetHostByAppIdAndIpAddr(store, params["id"], params["address"])
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	hostJson, err := json.Marshal(host)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(hostJson)
}

func PostApiAppIdReaderWriter(w http.ResponseWriter, r *http.Request) {
	params := gorilla_mux.Vars(r)

	w.Header().Set("Content-Type", "application/json")

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

	if _, ok := data["Hostname"]; !ok {
		err = errors.New("Data does not contain Hostname.")
		libhttp.HandleErrorJson(w, err)
		return
	}

	app, err := resourcedmaster_dao.GetApplicationById(store, params["id"])
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	hostname := data["Hostname"].(string)

	host := resourcedmaster_dao.NewHost(store, hostname, app.Id)
	host.Tags = data["Tags"].([]string)
	host.NetworkInterfaces = data["NetworkInterfaces"].(map[string]map[string]interface{})

	err = host.Save()
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	if params["reader-or-writer"] == "r" {
		err = app.SaveReaderWriterJson("reader", params["path"], dataJson)
		if err != nil {
			libhttp.HandleErrorJson(w, err)
			return
		}

		err = app.SaveReaderWriterByHostJson("reader", hostname, params["path"], dataJson)
		if err != nil {
			libhttp.HandleErrorJson(w, err)
			return
		}

	} else if params["reader-or-writer"] == "w" {
		err = app.SaveReaderWriterJson("writer", params["path"], dataJson)
		if err != nil {
			libhttp.HandleErrorJson(w, err)
			return
		}

		err = app.SaveReaderWriterByHostJson("writer", hostname, params["path"], dataJson)
		if err != nil {
			libhttp.HandleErrorJson(w, err)
			return
		}

	}

	messageJson, err := json.Marshal(
		map[string]string{
			"Message": fmt.Sprintf("%v{Path: %v} is saved.", params["reader-or-writer"], params["path"])})

	w.Write(messageJson)
}
