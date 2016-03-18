package common

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// StartServer starts and HTTP server using the given address and HTTPMux as the router
func StartServer(router *mux.Router) {
	bindAddr := ""
	bindPort := "8000"
	if Config.Webserver.Address != "" {
		bindAddr = Config.Webserver.Address
	}
	if Config.Webserver.Port != 0 {
		bindPort = strconv.Itoa(Config.Webserver.Port)
	}
	if bindAddr == "" {
		fmt.Printf("Now listening on *:%s\n", bindPort)
	} else {
		fmt.Printf("Now listening on %s:%s\n", bindAddr, bindPort)
	}
	http.ListenAndServe(bindAddr+":"+bindPort, router)
}

// NotImplementedHandler is a mock handler for paths that aren't implemented yet
func NotImplementedHandler(w http.ResponseWriter, r *http.Request) {
	resp := fmt.Sprintf("The path \"%s\" is not implemented yet\n", r.URL.Path)
	w.Write([]byte(resp))
}
