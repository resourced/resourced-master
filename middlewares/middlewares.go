// Package middlewares provides common middleware handlers.
package middlewares

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
)

func SetAddr(addr string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(addr, ":") {
				addr = "localhost" + addr
			}

			context.Set(r, "addr", addr)

			next.ServeHTTP(w, r)
		})
	}
}

func SetDBs(dbConfig *config.DBConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			context.Set(r, "db.Core", dbConfig.Core)
			context.Set(r, "multidb.TSWatchers", dbConfig.TSWatchers)
			context.Set(r, "multidb.TSMetrics", dbConfig.TSMetrics)

			next.ServeHTTP(w, r)
		})
	}
}

func SetCookieStore(cookieStore *sessions.CookieStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			context.Set(r, "cookieStore", cookieStore)

			next.ServeHTTP(w, r)
		})
	}
}

// SetClusters sets clusters data in context based on logged in user ID.
func SetClusters(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)
		session, _ := cookieStore.Get(r, "resourcedmaster-session")
		userRowInterface := session.Values["user"]

		if userRowInterface == nil {
			http.Redirect(w, r, "/login", 301)
			return
		}

		userRow := userRowInterface.(*dal.UserRow)

		db := context.Get(r, "db.Core").(*sqlx.DB)

		clusterRows, err := dal.NewCluster(db).AllClustersByUserID(nil, userRow.ID)
		if err != nil {
			libhttp.HandleErrorJson(w, err)
			return
		}

		context.Set(r, "clusters", clusterRows)

		// Set currentCluster if not previously set.
		if len(clusterRows) > 0 {
			currentClusterInterface := session.Values["currentCluster"]
			if currentClusterInterface == nil {
				session.Values["currentCluster"] = clusterRows[0]

				err := session.Save(r, w)
				if err != nil {
					libhttp.HandleErrorJson(w, err)
					return
				}
			}
		}

		// Set currentClusterJson
		currentClusterInterface := session.Values["currentCluster"]
		if currentClusterInterface != nil {
			currentClusterRow := currentClusterInterface.(*dal.ClusterRow)

			currentClusterJson, err := json.Marshal(currentClusterRow)
			if err != nil {
				libhttp.HandleErrorJson(w, err)
				return
			}
			context.Set(r, "currentClusterJson", currentClusterJson)

		} else {
			context.Set(r, "currentClusterJson", []byte("{}"))
		}

		next.ServeHTTP(w, r)
	})
}

// MustLogin is a middleware that checks existence of current user.
func MustLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)
		session, _ := cookieStore.Get(r, "resourcedmaster-session")
		userRowInterface := session.Values["user"]

		if userRowInterface == nil {
			http.Redirect(w, r, "/login", 301)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// MustLoginApi is a middleware that checks /api login.
func MustLoginApi(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")

		if auth == "" {
			libhttp.BasicAuthUnauthorized(w, nil)
			return
		}

		accessTokenString, _, ok := libhttp.ParseBasicAuth(auth)
		if !ok {
			libhttp.BasicAuthUnauthorized(w, nil)
			return
		}

		db := context.Get(r, "db.Core").(*sqlx.DB)

		accessTokenRow, err := dal.NewAccessToken(db).GetByAccessToken(nil, accessTokenString)
		if err != nil {
			libhttp.BasicAuthUnauthorized(w, nil)
			return
		}
		if accessTokenRow == nil {
			libhttp.BasicAuthUnauthorized(w, nil)
			return
		}

		if !accessTokenRow.Enabled {
			libhttp.BasicAuthUnauthorized(w, nil)
			return
		}

		isAllowed := false

		if r.Method == "GET" {
			isAllowed = true
		} else if accessTokenRow.Level == "write" || accessTokenRow.Level == "execute" {
			isAllowed = true
		}

		if !isAllowed {
			libhttp.BasicAuthUnauthorized(w, nil)
			return
		}

		context.Set(r, "accessTokenRow", accessTokenRow)

		next.ServeHTTP(w, r)
	})
}
