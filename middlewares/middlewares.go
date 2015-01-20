package middlewares

import (
	resourcedmaster_dao "github.com/resourced/resourced-master/dao"
	"github.com/resourced/resourced-master/libhttp"
	"net/http"
)

func AccessTokenAuth(accessTokens []*resourcedmaster_dao.AccessToken) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			auth := req.Header.Get("Authorization")

			if auth == "" && len(accessTokens) > 0 {
				libhttp.BasicAuthUnauthorized(res)
				return
			}

			username, _, ok := libhttp.ParseBasicAuth(auth)
			if !ok {
				libhttp.BasicAuthUnauthorized(res)
				return
			}

			accessTokenMatched := false

			for _, accessToken := range accessTokens {
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
