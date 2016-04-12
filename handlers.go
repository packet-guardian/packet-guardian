package main

import (
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/onesimus-systems/packet-guardian/common"
	"github.com/onesimus-systems/packet-guardian/dhcp"
)

type device struct {
	MAC            string
	UA             string
	Platform       string
	IP             string
	DateRegistered string
}

func rootHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := strings.Split(r.RemoteAddr, ":")[0]
		reg, err := dhcp.IsRegisteredByIP(e.DB, net.ParseIP(ip))
		if err != nil {
			e.Log.Errorf("Error checking auto registration IP: %s", err.Error())
		}
		if reg {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		} else {
			http.Redirect(w, r, "/register", http.StatusTemporaryRedirect)
		}
	}
}

func userDeviceListHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, ok := mux.Vars(r)["username"]
		if !ok {
			username = e.Sessions.GetSession(r, e.Config.Webserver.SessionName).GetString("username")
		}
		results := dhcp.Query{User: username}.Search(e)

		data := struct {
			SiteTitle   string
			CompanyName string
			Username    string
			Devices     []dhcp.Result
		}{
			SiteTitle:   e.Config.Core.SiteTitle,
			CompanyName: e.Config.Core.SiteCompanyName,
			Username:    username,
			Devices:     results,
		}
		e.Templates.ExecuteTemplate(w, "manage", data)
	}
}

func adminHomeHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin/search", http.StatusTemporaryRedirect)
	}
}

func adminSearchHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Was a search query performed
		query := r.FormValue("q")
		var results []dhcp.Result
		if query == "*" {
			q := dhcp.Query{}
			q.User = ""
			results = q.Search(e)
		} else if query != "" {
			e.Log.Infof("Searching for %s", query)
			q := dhcp.Query{}

			if m, err := dhcp.FormatMacAddress(query); err == nil {
				q.MAC = m
			} else if ip := net.ParseIP(query); ip != nil {
				q.IP = ip
			}
			q.User = query
			results = q.Search(e)
		}

		data := struct {
			SiteTitle     string
			CompanyName   string
			Query         string
			SearchResults []dhcp.Result
		}{
			SiteTitle:     e.Config.Core.SiteTitle,
			CompanyName:   e.Config.Core.SiteCompanyName,
			Query:         query,
			SearchResults: results,
		}
		e.Templates.ExecuteTemplate(w, "admin-search", data)
	}
}
