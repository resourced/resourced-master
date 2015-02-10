// Package middlewares provides common middleware handlers.
package middlewares

import (
	"fmt"
	"github.com/gorilla/context"
	resourcedmaster_dao "github.com/resourced/resourced-master/dao"
	"github.com/resourced/resourced-master/libhttp"
	resourcedmaster_storage "github.com/resourced/resourced-master/storage"
	"net/http"
	"strings"
)

func SetStore(store resourcedmaster_storage.Storer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			context.Set(req, "store", store)

			next.ServeHTTP(res, req)
		})
	}
}

func SetCurrentApplication(store resourcedmaster_storage.Storer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			urlChunks := strings.Split(req.URL.Path, "/")

			if len(urlChunks) >= 4 && urlChunks[1] == "api" && urlChunks[2] == "app" {
				appId := urlChunks[3]

				app, err := resourcedmaster_dao.GetApplicationById(store, appId)
				if err != nil {
					libhttp.BasicAuthUnauthorized(res, err)
					return
				}

				context.Set(req, "currentApp", app)
			}

			next.ServeHTTP(res, req)
		})
	}
}

func AccessTokenAuth(users []*resourcedmaster_dao.User) func(http.Handler) http.Handler {
	getCurrentApp := func(req *http.Request) *resourcedmaster_dao.Application {
		currentAppInterface := context.Get(req, "currentApp")
		if currentAppInterface == nil {
			return nil
		}
		return currentAppInterface.(*resourcedmaster_dao.Application)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			auth := req.Header.Get("Authorization")

			if auth == "" && len(users) > 0 {
				libhttp.BasicAuthUnauthorized(res, nil)
				return
			}

			username, _, ok := libhttp.ParseBasicAuth(auth)
			if !ok {
				libhttp.BasicAuthUnauthorized(res, nil)
				return
			}

			currentApp := getCurrentApp(req)
			accessAllowed := false
			adminLevelOrAbove := false

			if currentApp == nil {
				adminLevelOrAbove = strings.HasPrefix(req.URL.Path, "/api/users")
			} else {
				adminLevelOrAbove = strings.HasPrefix(req.URL.Path, fmt.Sprintf("/api/app/%v/users", currentApp.Id))
			}

			for _, user := range users {
				if username == user.Token {
					if user.Level == "staff" {
						context.Set(req, "currentUser", user)
						accessAllowed = true
						break
					}

					// Check application id if user.Level != staff
					if currentApp != nil && currentApp.Id == user.ApplicationId {
						if user.Level == "admin" {
							context.Set(req, "currentUser", user)
							accessAllowed = true
							break

						} else if user.Level == "basic" && !adminLevelOrAbove {
							context.Set(req, "currentUser", user)
							accessAllowed = true
							break
						}
					}
				}
			}

			if !accessAllowed {
				libhttp.BasicAuthUnauthorized(res, nil)
				return
			}

			next.ServeHTTP(res, req)
		})
	}
}
