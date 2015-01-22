package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
	resourcedmaster_dao "github.com/resourced/resourced-master/dao"
	"github.com/resourced/resourced-master/libhttp"
	resourcedmaster_storage "github.com/resourced/resourced-master/storage"
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

	accessToken, err := resourcedmaster_dao.NewAccessToken(store, app)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	err = accessToken.Save()
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	accessTokenJson, err := json.Marshal(accessToken)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	w.Write(accessTokenJson)
}

func DeleteApiApplicationIdAccessToken(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	store := context.Get(r, "store").(resourcedmaster_storage.Storer)

	err := resourcedmaster_dao.DeleteAccessTokenByToken(store, ps.ByName("token"))
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	err = resourcedmaster_dao.DeleteApplicationAccessToken(store, ps.ByName("token"))
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
	fmt.Fprint(w, "/api\n")
}
