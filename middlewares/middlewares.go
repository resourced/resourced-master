package middlewares

import (
	"github.com/gorilla/context"
	resourcedmaster_dao "github.com/resourced/resourced-master/dao"
	"github.com/resourced/resourced-master/libhttp"
	resourcedmaster_storage "github.com/resourced/resourced-master/storage"
	"net/http"
	"strings"
)

func AccessTokenAuth(users []*resourcedmaster_dao.User) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			auth := req.Header.Get("Authorization")

			if auth == "" && len(users) > 0 {
				libhttp.BasicAuthUnauthorized(res)
				return
			}

			username, _, ok := libhttp.ParseBasicAuth(auth)
			if !ok {
				libhttp.BasicAuthUnauthorized(res)
				return
			}

			requireAdminLevelOrAbove := strings.HasPrefix(req.URL.Path, "/api/users/") || strings.HasPrefix(req.URL.Path, "/api/app/")
			accessAllowed := false

			for _, user := range users {
				if username == user.Token {
					if !requireAdminLevelOrAbove {
						context.Set(req, "currentUser", user)
						accessAllowed = true
						break
					} else if user.Level == "staff" {
						context.Set(req, "currentUser", user)
						accessAllowed = true
						break
					} else if user.Level == "admin" {
						context.Set(req, "currentUser", user)

						// This is not quite right. We need to compare to ApplicationId
						accessAllowed = true
						break
					}
				}
			}

			if !accessAllowed {
				libhttp.BasicAuthUnauthorized(res)
				return
			}

			next.ServeHTTP(res, req)
		})
	}
}

func SetStore(store resourcedmaster_storage.Storer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			context.Set(req, "store", store)

			next.ServeHTTP(res, req)
		})
	}
}
