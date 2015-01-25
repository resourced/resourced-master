package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
	resourcedmaster_dao "github.com/resourced/resourced-master/dao"
	"github.com/resourced/resourced-master/libhttp"
	resourcedmaster_storage "github.com/resourced/resourced-master/storage"
	"io/ioutil"
	"net/http"
	"strconv"
)

//
// Admin level access
//

// PostApiUser
func PostApiUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

func GetApiUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

func GetApiUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	user, err := resourcedmaster_dao.GetUserByName(store, ps.ByName("name"))
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

func PutApiUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	user, err := resourcedmaster_dao.UpdateUserByNameGivenJson(store, ps.ByName("name"), r.Body)
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

func DeleteApiUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	err := resourcedmaster_dao.DeleteUserByName(store, ps.ByName("name"))
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	messageJson, err := json.Marshal(
		map[string]string{
			"Message": fmt.Sprintf("User{Name: %v} is deleted.", ps.ByName("name"))})

	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(messageJson)
}

func PutApiUserNameAccessToken(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	user, err := resourcedmaster_dao.UpdateUserTokenByName(store, ps.ByName("name"))
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

func PostApiApplicationIdAccessToken(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	app, err := resourcedmaster_dao.GetApplicationById(store, id)
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

func DeleteApiApplicationIdAccessToken(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	err := resourcedmaster_dao.DeleteUserByName(store, ps.ByName("token"))
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	err = resourcedmaster_dao.DeleteApplicationByAccessToken(store, ps.ByName("token"))
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	messageJson, err := json.Marshal(
		map[string]string{
			"Message": fmt.Sprintf("AccessToken{Token: %v} is deleted.", ps.ByName("token"))})

	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(messageJson)
}

//
// Basic level access
//

func GetRoot(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.Redirect(w, r, "/api", 301)
}

func GetApi(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	currentUser := context.Get(r, "currentUser").(resourcedmaster_dao.User)

	if currentUser.Level == "staff" {
		http.Redirect(w, r, "/api/app", 301)

	} else {
		if currentUser.ApplicationId <= 0 {
			libhttp.BasicAuthUnauthorized(w)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/api/app/%v/hosts", currentUser.ApplicationId), 301)
	}
}

func GetApiApp(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	currentUser := context.Get(r, "currentUser").(resourcedmaster_dao.User)

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

// **POST** `/api/app/:id/r/:path` Submit reader JSON data from 1 host.

// **POST** `/api/app/:id/w/:path` Submit writer JSON data from 1 host.

func PostApiAppIdReader(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	app, err := resourcedmaster_dao.GetApplicationById(id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	err = app.SaveReaderWriter("reader", ps.ByName("path"), data)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	messageJson, err := json.Marshal(
		map[string]string{
			"Message": fmt.Sprintf("Reader{Path: %v} is saved.", ps.ByName("path"))})

	w.Write(messageJson)
}

func PostApiAppIdWriter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	app, err := resourcedmaster_dao.GetApplicationById(id)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	err = app.SaveReaderWriter("writer", ps.ByName("path"), data)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	messageJson, err := json.Marshal(
		map[string]string{
			"Message": fmt.Sprintf("Writer{Path: %v} is saved.", ps.ByName("path"))})

	w.Write(messageJson)
}
