// Package middlewares provides common middleware handlers.
package middlewares

import (
	"context"
	"errors"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"github.com/pressly/chi"

	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/models/cassandra"
)

// CSRFMiddleware is a constructor that creates github.com/gorilla/csrf.CSRF struct
func CSRFMiddleware(useHTTPS bool, salt string) func(http.Handler) http.Handler {
	CSRFOptions := csrf.Secure(false)
	if useHTTPS {
		CSRFOptions = csrf.Secure(true)
	}
	return csrf.Protect([]byte(salt), CSRFOptions)
}

// SetContext assigns context struct to every request handler
func SetContext(ctx context.Context) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			// Make sure we don't lose Chi's context.
			chiContext := r.Context().Value(chi.RouteCtxKey)
			r = r.WithContext(ctx)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiContext))

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// SetClusters sets clusters data in context based on logged in user ID.
func SetClusters(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookieStore := r.Context().Value("CookieStore").(*sessions.CookieStore)
		session, _ := cookieStore.Get(r, "resourcedmaster-session")
		userRowInterface := session.Values["user"]

		if userRowInterface == nil {
			http.Redirect(w, r, "/login", 301)
			return
		}

		userRow := userRowInterface.(*cassandra.UserRow)

		clusterRows, err := cassandra.NewCluster(r.Context()).AllByUserID(userRow.ID)
		if err != nil {
			libhttp.HandleErrorJson(w, err)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), "clusters", clusterRows))

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

		currentClusterInterface := session.Values["currentCluster"]
		if currentClusterInterface != nil {
			currentClusterRow := currentClusterInterface.(*cassandra.ClusterRow)

			r = r.WithContext(context.WithValue(r.Context(), "currentCluster", currentClusterRow))
		}

		next.ServeHTTP(w, r)
	})
}

// SetAccessTokens sets clusters data in context based on logged in user ID.
func SetAccessTokens(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentClusterInterface := r.Context().Value("currentCluster")
		if currentClusterInterface == nil {
			libhttp.HandleErrorJson(w, errors.New("Unable to get access tokens because current cluster is nil."))
			return
		}

		accessTokenRows, err := cassandra.NewAccessToken(r.Context()).AllByClusterID(currentClusterInterface.(*cassandra.ClusterRow).ID)
		if err != nil {
			libhttp.HandleErrorJson(w, err)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), "accessTokens", accessTokenRows))

		next.ServeHTTP(w, r)
	})
}

// MustLogin is a middleware that checks existence of current user.
func MustLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookieStore := r.Context().Value("CookieStore").(*sessions.CookieStore)
		session, _ := cookieStore.Get(r, "resourcedmaster-session")
		userRowInterface := session.Values["user"]

		if userRowInterface == nil {
			cookieStore := r.Context().Value("CookieStore").(*sessions.CookieStore)
			session, err := cookieStore.Get(r, "resourcedmaster-session")
			if err == nil {
				delete(session.Values, "user")
				session.Options.MaxAge = -1
			}

			http.Redirect(w, r, "/login", 301)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), "currentUser", userRowInterface.(*cassandra.UserRow)))

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

		accessTokenRow, err := cassandra.NewAccessToken(r.Context()).GetByAccessToken(accessTokenString)
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

		r = r.WithContext(context.WithValue(r.Context(), "accessToken", accessTokenRow))

		next.ServeHTTP(w, r)
	})
}

// MustLoginApiStream is a middleware that checks /api/.../stream login.
func MustLoginApiStream(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessTokenString := r.URL.Query().Get("accessToken")

		// Check the upper-case version
		if accessTokenString == "" {
			accessTokenString = r.URL.Query().Get("AccessToken")
		}

		// If still empty, then deny
		if accessTokenString == "" {
			libhttp.BasicAuthUnauthorized(w, nil)
			return
		}

		accessTokenRow, err := cassandra.NewAccessToken(r.Context()).GetByAccessToken(accessTokenString)
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

		r = r.WithContext(context.WithValue(r.Context(), "accessToken", accessTokenRow))

		next.ServeHTTP(w, r)
	})
}

func MustBeMember(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentClusterInterface := r.Context().Value("currentCluster")
		if currentClusterInterface == nil {
			http.Redirect(w, r, "/login", 301)
			return
		}

		currentUserInterface := r.Context().Value("currentUser")
		if currentUserInterface == nil {
			http.Redirect(w, r, "/login", 301)
			return
		}

		currentCluster := currentClusterInterface.(*cassandra.ClusterRow)
		currentUser := currentUserInterface.(*cassandra.UserRow)
		foundCurrentUser := false

		for _, member := range currentCluster.GetMembers() {
			if member.ID == currentUser.ID && member.Enabled {
				foundCurrentUser = true
				break
			}
		}

		if !foundCurrentUser {
			http.Redirect(w, r, "/login", 301)
			return
		}

		next.ServeHTTP(w, r)
	})
}
