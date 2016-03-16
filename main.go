package main

import (
	"fmt"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"

	_ "github.com/onesimus-systems/net-guardian/auth"
	"github.com/onesimus-systems/net-guardian/common"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	session := common.GetSession(r, "net-guardian")
	if session.GetBool("loggedin", false) {
		http.Redirect(w, r, "/overview", http.StatusTemporaryRedirect)
	} else {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	}
}

func main() {
	common.StartSessionStore()
	common.HTTPMux.HandleFunc("/", rootHandler)
	common.HTTPMux.PathPrefix("/static").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	bindAddr := ""
	bindPort := "8000"
	if common.Config.Webserver.Address != "" {
		bindAddr = common.Config.Webserver.Address
	}
	if common.Config.Webserver.Port != 0 {
		bindPort = strconv.Itoa(common.Config.Webserver.Port)
	}

	fmt.Printf("Now listening on %s:%s\n", bindAddr, bindPort)
	common.StartServer(bindAddr + ":" + bindPort)
}
