// Package middlewares provides common middleware handlers.
package middlewares

import (
	// "fmt"
	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
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

// func AccessTokenAuth(users []*resourcedmaster_dal.UserRow) func(http.Handler) http.Handler {
// 	getCurrentApp := func(req *http.Request) *resourcedmaster_dal.ApplicationRow {
// 		currentAppInterface := context.Get(req, "currentApp")
// 		if currentAppInterface == nil {
// 			return nil
// 		}
// 		return currentAppInterface.(*resourcedmaster_dal.ApplicationRow)
// 	}

// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
// 			if strings.HasPrefix(req.URL.Path, "/api") {
// 				auth := req.Header.Get("Authorization")

// 				if auth == "" && len(users) > 0 {
// 					libhttp.BasicAuthUnauthorized(res, nil)
// 					return
// 				}

// 				username, _, ok := libhttp.ParseBasicAuth(auth)
// 				if !ok {
// 					libhttp.BasicAuthUnauthorized(res, nil)
// 					return
// 				}

// 				currentApp := getCurrentApp(req)
// 				accessAllowed := false
// 				adminLevelOrAbove := false

// 				if currentApp == nil {
// 					adminLevelOrAbove = strings.HasPrefix(req.URL.Path, "/api/users")
// 				} else {
// 					adminLevelOrAbove = strings.HasPrefix(req.URL.Path, fmt.Sprintf("/api/app/%v/users", currentApp.ID))
// 				}

// 				for _, user := range users {
// 					if username == user.Token.String {
// 						if user.Level == "staff" {
// 							context.Set(req, "currentUser", user)
// 							accessAllowed = true
// 							break
// 						}

// 						// Check application id if user.Level != staff
// 						if currentApp != nil && currentApp.ID == user.ApplicationID.Int64 {
// 							if user.Level == "admin" {
// 								context.Set(req, "currentUser", user)
// 								accessAllowed = true
// 								break

// 							} else if user.Level == "basic" && !adminLevelOrAbove {
// 								context.Set(req, "currentUser", user)
// 								accessAllowed = true
// 								break
// 							}
// 						}
// 					}
// 				}

// 				if !accessAllowed {
// 					libhttp.BasicAuthUnauthorized(res, nil)
// 					return
// 				}

// 				next.ServeHTTP(res, req)
// 			} else {
// 				next.ServeHTTP(res, req)
// 			}
// 		})
// 	}
// }
