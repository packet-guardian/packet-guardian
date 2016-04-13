package main

import (
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/onesimus-systems/packet-guardian/auth"
	"github.com/onesimus-systems/packet-guardian/common"
	"github.com/onesimus-systems/packet-guardian/dhcp"
)

func rootHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := strings.Split(r.RemoteAddr, ":")[0]
		reg, err := dhcp.IsRegisteredByIP(e.DB, net.ParseIP(ip))
		if err != nil {
			e.Log.Errorf("Error checking auto registration IP: %s", err.Error())
		}

		if auth.IsLoggedIn(e, r) {
			http.Redirect(w, r, "/manage", http.StatusTemporaryRedirect)
			return
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

		admin := auth.IsAdminUser(e, r)
		bl, _ := dhcp.IsBlacklisted(e.DB, username)
		showAddBtn := (admin || (e.Config.Core.AllowManualRegistrations && !bl))

		data := struct {
			SiteTitle     string
			CompanyName   string
			Username      string
			IsAdmin       bool
			Devices       []dhcp.Device
			FlashMessage  string
			ShowAddBtn    bool
			IsBlacklisted bool
		}{
			SiteTitle:     e.Config.Core.SiteTitle,
			CompanyName:   e.Config.Core.SiteCompanyName,
			Username:      username,
			IsAdmin:       auth.IsAdminUser(e, r),
			Devices:       results,
			ShowAddBtn:    showAddBtn,
			IsBlacklisted: bl,
		}
		if err := e.Templates.ExecuteTemplate(w, "manage", data); err != nil {
			e.Log.Error(err.Error())
		}
	}
}

func adminHomeHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			SiteTitle    string
			CompanyName  string
			FlashMessage string
		}{
			SiteTitle:   e.Config.Core.SiteTitle,
			CompanyName: e.Config.Core.SiteCompanyName,
		}
		if err := e.Templates.ExecuteTemplate(w, "admin-dash", data); err != nil {
			e.Log.Error(err.Error())
		}
	}
}

func adminSearchHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Was a search query performed
		query := r.FormValue("q")
		var results []dhcp.Device
		q := dhcp.Query{}
		if query == "*" {
			q.User = ""
		} else if query != "" {
			if m, err := dhcp.FormatMacAddress(query); err == nil {
				q.MAC = m
			} else if ip := net.ParseIP(query); ip != nil {
				q.IP = ip
			} else {
				q.User = query
			}
		}

		if query != "" {
			results = q.Search(e)
		}

		noResultType := ""
		if query != "" && len(results) == 0 {
			if q.User != "" {
				noResultType = "username"
			} else if q.MAC != nil {
				noResultType = "mac"
			} else if q.IP != nil {
				noResultType = "ip"
			}
		}

		data := struct {
			SiteTitle     string
			CompanyName   string
			Query         string
			SearchResults []dhcp.Device
			FlashMessage  string
			NoResultType  string
		}{
			SiteTitle:     e.Config.Core.SiteTitle,
			CompanyName:   e.Config.Core.SiteCompanyName,
			Query:         query,
			SearchResults: results,
			NoResultType:  noResultType,
		}
		if err := e.Templates.ExecuteTemplate(w, "admin-search", data); err != nil {
			e.Log.Error(err.Error())
		}
	}
}

func adminBlacklistHandler(e *common.Environment, all bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var black []interface{}

		if r.FormValue("devices") != "" {
			devices := strings.Split(r.FormValue("devices"), ",")
			for _, d := range devices {
				black = append(black, d)
			}
		} else {
			username, ok := mux.Vars(r)["username"]
			if !ok {
				common.NewAPIResponse(common.APIStatusGenericError, "No username given", nil).WriteTo(w)
				return
			}
			black = append(black, username)

			splitPath := strings.Split(r.URL.Path, "/")
			if splitPath[len(splitPath)-1] == "all" {
				results := dhcp.Query{User: username}.Search(e)
				for _, r := range results {
					black = append(black, r.MAC)
				}
			}
		}

		if r.Method == "DELETE" {
			err := dhcp.RemoveFromBlacklist(e.DB, black...)
			if err != nil {
				e.Log.Errorf("Error removing from blacklist: %s", err.Error())
				common.NewAPIResponse(common.APIStatusGenericError, "Error removing from blacklist", nil).WriteTo(w)
				return
			}
			for _, d := range black {
				e.Log.Infof("Removed user/MAC %s from blacklist", d)
			}
			common.NewAPIOK("Unblacklisting successful", nil).WriteTo(w)
		} else {
			err := dhcp.AddToBlacklist(e.DB, black...)
			if err != nil {
				e.Log.Errorf("Error blacklisting: %s", err.Error())
				common.NewAPIResponse(common.APIStatusGenericError, "Error blacklisting", nil).WriteTo(w)
				return
			}
			for _, d := range black {
				e.Log.Infof("Blacklisted user/MAC %s", d)
			}
			common.NewAPIOK("Blacklisting successful", nil).WriteTo(w)
		}
	}
}
