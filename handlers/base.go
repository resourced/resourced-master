// Package handlers provides HTTP request handlers for the Application.
package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pressly/chi"

	"github.com/resourced/resourced-master/models/pg"
)

func getInt64SlugFromPath(w http.ResponseWriter, r *http.Request, slug string) (int64, error) {
	inString := chi.URLParam(r, slug)
	if inString == "" {
		return -1, errors.New(fmt.Sprintf("%v cannot be empty.", slug))
	}

	data, err := strconv.ParseInt(inString, 10, 64)
	if err != nil {
		return -1, err
	}

	return data, nil
}

func getAccessToken(w http.ResponseWriter, r *http.Request, level string) (*pg.AccessTokenRow, error) {
	accessTokensInterface := r.Context().Value("accessTokens")
	if accessTokensInterface == nil {
		return nil, errors.New("Failed to get access token because the full list of access tokens is nil.")
	}

	for _, accessToken := range accessTokensInterface.([]*pg.AccessTokenRow) {
		if level == "read" {
			return accessToken, nil

		} else if level == "write" {
			if (accessToken.Level == "write") || (accessToken.Level == "execute") {
				return accessToken, nil
			}
		} else if level == "execute" {
			if accessToken.Level == "execute" {
				return accessToken, nil
			}
		}
	}

	return nil, errors.New("Failed to pick the right access token.")
}
