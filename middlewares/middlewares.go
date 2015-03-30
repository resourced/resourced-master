// Package middlewares provides common middleware handlers.
package middlewares

import (
	// "fmt"
	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	resourcedmaster_dal "github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
	"net/http"
)

func SetDB(db *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			context.Set(req, "db", db)

			next.ServeHTTP(res, req)
		})
	}
}

func SetRiceBoxes(riceBoxes map[string]*rice.Box) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			context.Set(req, "riceBoxes", riceBoxes)

			next.ServeHTTP(res, req)
		})
	}
}

func SetCookieStore(cookieStore *sessions.CookieStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			context.Set(req, "cookieStore", cookieStore)

			next.ServeHTTP(res, req)
		})
	}
}

// MustLogin is a middleware that checks existence of current user.
func MustLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		cookieStore := context.Get(req, "cookieStore").(*sessions.CookieStore)
		session, _ := cookieStore.Get(req, "resourcedmaster-session")
		userRowInterface := session.Values["user"]

		if userRowInterface == nil {
			http.Redirect(res, req, "/login", 301)
			return
		}

		next.ServeHTTP(res, req)
	})
}

// MustLoginApi is a middleware that checks /api login.
func MustLoginApi(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		auth := req.Header.Get("Authorization")

		if auth == "" {
			libhttp.BasicAuthUnauthorized(res, nil)
			return
		}

		accessTokenString, _, ok := libhttp.ParseBasicAuth(auth)
		if !ok {
			libhttp.BasicAuthUnauthorized(res, nil)
			return
		}

		db := context.Get(req, "db").(*sqlx.DB)

		accessTokenRow, err := resourcedmaster_dal.NewAccessToken(db).GetByAccessToken(nil, accessTokenString)
		if err != nil {
			libhttp.BasicAuthUnauthorized(res, nil)
			return
		}
		if accessTokenRow == nil {
			libhttp.BasicAuthUnauthorized(res, nil)
			return
		}

		context.Set(req, "accessTokenRow", accessTokenRow)

		next.ServeHTTP(res, req)
	})
}
