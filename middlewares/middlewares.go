package middlewares

import (
	"github.com/gorilla/context"
	resourcedmaster_dao "github.com/resourced/resourced-master/dao"
	"github.com/resourced/resourced-master/libhttp"
	resourcedmaster_storage "github.com/resourced/resourced-master/storage"
	"net/http"
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

			accessTokenMatched := false

			for _, accessToken := range users {
				if username == accessToken.Token {
					accessTokenMatched = true
					break
				}
			}

			if !accessTokenMatched {
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
