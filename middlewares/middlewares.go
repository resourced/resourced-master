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
	"strconv"
	"strings"
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

func SetCurrentApplication(db *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			urlChunks := strings.Split(req.URL.Path, "/")

			if len(urlChunks) >= 4 && urlChunks[1] == "api" && urlChunks[2] == "app" {
				appIdString := urlChunks[3]
				appId, err := strconv.ParseInt(appIdString, 10, 64)
				if err != nil {
					libhttp.BasicAuthUnauthorized(res, err)
					return
				}

				appRow, err := resourcedmaster_dal.NewApplication(db).GetById(nil, appId)
				if err != nil {
					libhttp.BasicAuthUnauthorized(res, err)
					return
				}

				context.Set(req, "currentApp", appRow)
			}

			next.ServeHTTP(res, req)
		})
	}
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
