package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/onesimus-systems/packet-guardian/auth"
	"github.com/onesimus-systems/packet-guardian/common"
	"github.com/onesimus-systems/packet-guardian/dhcp"
)

func makeRoutes(e *common.Environment) http.Handler {
	r := mux.NewRouter()
	// Root routes to either the auto registration page or the management page
	r.HandleFunc("/", rootHandler(e))
	// Static assets, images, CSS, JS, etc.
	r.PathPrefix("/public").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public/"))))

	// Automatic registration page
	r.HandleFunc("/register", dhcp.RegistrationPageHandler(e)).Methods("GET")
	r.HandleFunc("/register", dhcp.AutoRegisterHandler(e)).Methods("POST")

	// Login page
	r.HandleFunc("/login", auth.LoginPageHandler(e)).Methods("GET")
	r.HandleFunc("/login", auth.LoginHandler(e)).Methods("POST")
	r.HandleFunc("/logout", auth.LogoutHandler(e)).Methods("GET")

	// User management page
	r.HandleFunc("/manage", auth.CheckAuthMid(e, manageHandler(e))).Methods("GET")

	// Device actions
	r.HandleFunc("/devices/delete", auth.CheckAuthAPIMid(e, dhcp.DeleteHandler(e))).Methods("POST")

	// Development only routes
	if dev {
		r.HandleFunc("/dev/reloadtemp", reloadTemplates(e)).Methods("GET")
		r.HandleFunc("/dev/reloadconf", reloadConfiguration(e)).Methods("GET")
	}
	return r
}

// Dev mode route handlers
func reloadTemplates(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templates, err := parseTemplates("templates/*.tmpl")
		if err != nil {
			w.Write([]byte("Error loading HTML templates: " + err.Error()))
			return
		}
		e.Templates = templates
		w.Write([]byte("Templates reloaded"))
	}
}

func reloadConfiguration(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		config, err := loadConfig("")
		if err != nil {
			w.Write([]byte("Error loading config: " + err.Error()))
			return
		}
		e.Config = config
		w.Write([]byte("Configuration reloaded"))
	}
}
