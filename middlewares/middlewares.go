// Package middlewares provides common middleware handlers.
package middlewares

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/didip/stopwatch"
	"github.com/gorilla/context"
	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"github.com/justinas/alice"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/mailer"
	"github.com/resourced/resourced-master/pubsub"
)

// SetStringKeyValue set arbitrary key value on context and passes it around to every request handler
func SetStringKeyValue(key, value string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			context.Set(r, key, value)

			next.ServeHTTP(w, r)
		})
	}
}

// SetAddr passes daemon host and port to every request handler
func SetAddr(addr string) func(http.Handler) http.Handler {
	if strings.HasPrefix(addr, ":") {
		addr = "localhost" + addr
	}
	return SetStringKeyValue("addr", addr)
}

// SetVIPAddr passes VIP host and port to every request handler
func SetVIPAddr(vipAddr string) func(http.Handler) http.Handler {
	if strings.HasPrefix(vipAddr, ":") {
		vipAddr = "localhost" + vipAddr
	}
	return SetStringKeyValue("vipAddr", vipAddr)
}

// SetVIPProtocol passes VIP protocol to every request handler
func SetVIPProtocol(vipProtocol string) func(http.Handler) http.Handler {
	return SetStringKeyValue("vipProtocol", vipProtocol)
}

// SetDBs passes all database connections to every request handler
func SetDBs(dbConfig *config.DBConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			context.Set(r, "dbs", dbConfig)

			next.ServeHTTP(w, r)
		})
	}
}

// SetCookieStore passes cookie storage to every request handler
func SetCookieStore(cookieStore *sessions.CookieStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			context.Set(r, "cookieStore", cookieStore)

			next.ServeHTTP(w, r)
		})
	}
}

// SetMailers passes all mailers to every request handler
func SetMailers(mailers map[string]*mailer.Mailer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for key, mailr := range mailers {
				context.Set(r, "mailer."+key, mailr)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SetPubSubPublisher passes pubsub publisher socket to every request handler
func SetPubSubPublisher(publisher *pubsub.PubSub) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			context.Set(r, "pubsubPublisher", publisher)

			next.ServeHTTP(w, r)
		})
	}
}

// SetPubSubSubscribers passes pubsub subscribers socket to every request handler
func SetPubSubSubscribers(subscribers map[string]*pubsub.PubSub) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			context.Set(r, "pubsubSubscribers", subscribers)

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

		dbs := context.Get(r, "dbs").(*config.DBConfig)

		var clusterRows []*dal.ClusterRow
		var err error

		f := func() {
			clusterRows, err = dal.NewCluster(dbs.Core).AllByUserID(nil, userRow.ID)
		}

		// Measure the latency of AllByUserID because it is called on every request.
		latency := stopwatch.Measure(f)
		logrus.WithFields(logrus.Fields{
			"Method":              "Cluster.AllByUserID",
			"UserID":              userRow.ID,
			"LatencyNanoSeconds":  latency,
			"LatencyMicroSeconds": latency / 1000,
			"LatencyMilliSeconds": latency / 1000 / 1000,
		}).Info("Latency measurement")

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

		currentClusterInterface := session.Values["currentCluster"]
		if currentClusterInterface != nil {
			currentClusterRow := currentClusterInterface.(*dal.ClusterRow)
			context.Set(r, "currentCluster", currentClusterRow)
		}

		next.ServeHTTP(w, r)
	})
}

// SetAccessTokens sets clusters data in context based on logged in user ID.
func SetAccessTokens(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dbs := context.Get(r, "dbs").(*config.DBConfig)

		currentClusterInterface := context.Get(r, "currentCluster")
		if currentClusterInterface == nil {
			libhttp.HandleErrorJson(w, errors.New("Unable to get access tokens because current cluster is nil."))
			return
		}

		accessTokenRows, err := dal.NewAccessToken(dbs.Core).AllByClusterID(nil, currentClusterInterface.(*dal.ClusterRow).ID)
		if err != nil {
			libhttp.HandleErrorJson(w, err)
			return
		}

		context.Set(r, "accessTokens", accessTokenRows)

		next.ServeHTTP(w, r)
	})
}

// CSRFMiddleware is a constructor that creates github.com/gorilla/csrf.CSRF struct
func CSRFMiddleware(useHTTPS bool, salt string) func(http.Handler) http.Handler {
	CSRFOptions := csrf.Secure(false)
	if useHTTPS {
		CSRFOptions = csrf.Secure(true)
	}
	return csrf.Protect([]byte(salt), CSRFOptions)
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

		context.Set(r, "currentUser", userRowInterface.(*dal.UserRow))

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

		dbs := context.Get(r, "dbs").(*config.DBConfig)

		accessTokenRow, err := dal.NewAccessToken(dbs.Core).GetByAccessToken(nil, accessTokenString)
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

		context.Set(r, "accessToken", accessTokenRow)

		next.ServeHTTP(w, r)
	})
}

func MustBeMember(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentClusterInterface := context.Get(r, "currentCluster")
		if currentClusterInterface == nil {
			http.Redirect(w, r, "/login", 301)
			return
		}

		currentUserInterface := context.Get(r, "currentUser")
		if currentUserInterface == nil {
			http.Redirect(w, r, "/login", 301)
			return
		}

		currentCluster := currentClusterInterface.(*dal.ClusterRow)
		currentUser := currentUserInterface.(*dal.UserRow)
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

func MinimumMiddlewareChain(csrf func(http.Handler) http.Handler) alice.Chain {
	return alice.New(csrf, MustLogin, SetClusters, MustBeMember)
}

func MinimumAPIMiddlewareChain() alice.Chain {
	return alice.New(MustLoginApi)
}
