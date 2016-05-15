// Package handlers provides HTTP request handlers for the Application.
package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func getInt64SlugFromPath(w http.ResponseWriter, r *http.Request, slug string) (int64, error) {
	inString := mux.Vars(r)[slug]
	if inString == "" {
		return -1, errors.New(fmt.Sprintf("%v cannot be empty.", slug))
	}

	data, err := strconv.ParseInt(inString, 10, 64)
	if err != nil {
		return -1, err
	}

	return data, nil
}
