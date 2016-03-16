package common

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// HTTPMux is a gorilla/router object for packages to register their handlers
var HTTPMux *mux.Router

func init() {
	HTTPMux = mux.NewRouter()
}

// StartServer starts and HTTP server using the given address and HTTPMux as the router
func StartServer(address string) {
	http.ListenAndServe(address, HTTPMux)
}

// NotImplementedHandler is a mock handler for paths that aren't implemented yet
func NotImplementedHandler(w http.ResponseWriter, r *http.Request) {
	resp := fmt.Sprintf("The path \"%s\" is not implemented yet\n", r.URL.Path)
	w.Write([]byte(resp))
}
