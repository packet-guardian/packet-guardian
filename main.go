package main

import (
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"

	"github.com/onesimus-systems/net-guardian/common"
	"github.com/onesimus-systems/net-guardian/devices"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if devices.IsRegistered(r.RemoteAddr) {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	} else {
		http.Redirect(w, r, "/register", http.StatusTemporaryRedirect)
	}
}

func main() {
	common.LoadConfig("")
	common.StartSessionStore()
	common.ConnectDatabase()

	r := mux.NewRouter()
	r.HandleFunc("/", rootHandler)
	r.HandleFunc("/register", devices.RegisterHTTPHandler).Methods("GET")
	r.HandleFunc("/register/auto", devices.AutoRegisterHandler).Methods("POST")
	//r.PathPrefix("/static").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	// r.HandleFunc("/overview", overviewHandler)
	//
	// r.HandleFunc("/login", auth.LoginPageHandler).Methods("GET")
	// r.HandleFunc("/login", common.NotImplementedHandler).Methods("POST")

	common.StartServer(r)
}
