package main

import (
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/resourced/resourced-master/libenv"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Welcome to the home page!")
	})

	serverAddress := libenv.EnvWithDefault("RESOURCED_MASTER_ADDR", ":55655")

	n := negroni.Classic()
	n.UseHandler(mux)

	http.ListenAndServe(serverAddress, n)
}
