package main

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

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
			if auth.IsAdminUser(e, r) {
				http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
			} else {
				http.Redirect(w, r, "/manage", http.StatusTemporaryRedirect)
			}
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

func adminBlacklistHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Slice of MAC addresses
		var black []interface{}

		if r.FormValue("devices") != "" {
			deviceIDs := strings.Split(r.FormValue("devices"), ",")
			// Need to convert strings to int for database search
			ids := make([]int, len(deviceIDs))
			for i := range deviceIDs {
				in, _ := strconv.Atoi(deviceIDs[i])
				ids[i] = in
			}

			devices := dhcp.Query{ID: ids}.Search(e)
			for i := range devices {
				black = append(black, devices[i].MAC)
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

func adminUserHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			saveUserHandler(e, w, r)
			return
		} else if r.Method == "DELETE" {
			deleteUserHandler(e, w, r)
			return
		}

		data := struct {
			SiteTitle    string
			CompanyName  string
			Query        string
			Users        []*auth.User
			FlashMessage string
		}{
			SiteTitle:   e.Config.Core.SiteTitle,
			CompanyName: e.Config.Core.SiteCompanyName,
		}

		username := mux.Vars(r)["username"]
		var template string
		if username == "" {
			users, err := auth.GetAllUsers(e.DB)
			if err != nil {
				e.Log.Errorf("Error getting users: %s", err.Error())
				data.FlashMessage = "Error getting users"
			}
			data.Users = users
			template = "admin-users"
		} else {
			user, _ := auth.GetUser(e.DB, username)
			if user == nil {
				user = auth.NewUser()
				user.Username = username
			}
			data.Users = []*auth.User{user}
			template = "admin-user"
		}

		if err := e.Templates.ExecuteTemplate(w, template, data); err != nil {
			e.Log.Error(err.Error())
		}
	}
}

func saveUserHandler(e *common.Environment, w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	// Get or create user
	user, _ := auth.GetUser(e.DB, username)
	if user == nil {
		e.Log.Info("Creating user")
		user = &auth.User{
			ID:       common.ConvertToInt(r.FormValue("user-id")),
			Username: username,
		}
	}

	// Password
	user.ClearPassword = (r.FormValue("clear-pass") == "true")
	if r.FormValue("password") != "" {
		user.NewPassword(r.FormValue("password"))
	}

	// Registered device limit
	limitType := r.FormValue("special-limit")
	if limitType == "global" {
		user.DeviceLimit = -1
	} else if limitType == "unlimited" {
		user.DeviceLimit = 0
	} else {
		user.DeviceLimit = common.ConvertToInt(r.FormValue("device-limit"))
	}

	// Expiration times
	loc, _ := time.LoadLocation("Local")
	if r.FormValue("device-expiration") == "0" || r.FormValue("device-expiration") == "" {
		user.DefaultExpiration = time.Unix(0, 0)
	} else if r.FormValue("device-expiration") == "1" {
		user.DefaultExpiration = time.Unix(1, 0)
	} else {
		user.DefaultExpiration, _ = time.ParseInLocation("2006-01-02 15:04:05", r.FormValue("device-expiration"), loc)
	}

	if r.FormValue("valid-after") == "0" || r.FormValue("valid-after") == "" {
		user.ValidAfter = time.Unix(0, 0)
	} else {
		user.ValidAfter, _ = time.ParseInLocation("2006-01-02 15:04:05", r.FormValue("valid-after"), loc)
	}

	if r.FormValue("valid-before") == "0" || r.FormValue("valid-before") == "" {
		user.ValidBefore = time.Unix(0, 0)
	} else {
		user.ValidBefore, _ = time.ParseInLocation("2006-01-02 15:04:05", r.FormValue("valid-before"), loc)
	}

	if err := user.Save(e.DB); err != nil {
		e.Log.Errorf("Error saving user: %s", err.Error())
		common.NewAPIResponse(common.APIStatusGenericError, "Error saving user", nil).WriteTo(w)
		return
	}

	e.Log.Infof("Created user augmentation: %s", user.Username)
	common.NewAPIOK("User created", nil).WriteTo(w)
}

func deleteUserHandler(e *common.Environment, w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]

	user, err := auth.GetUser(e.DB, username)
	if user == nil {
		e.Log.Errorf("Error deleting user: %s", err.Error())
		common.NewAPIResponse(common.APIStatusGenericError, "Error deleting user", nil).WriteTo(w)
		return
	}

	sql := "DELETE FROM \"user\" WHERE \"username\" = ?"
	_, err = e.DB.Exec(sql, username)
	if err != nil {
		e.Log.Errorf("Error deleting user: %s", err.Error())
		common.NewAPIResponse(common.APIStatusGenericError, "Error deleting user", nil).WriteTo(w)
		return
	}

	e.Log.Infof("Deleted user: %s", username)
	common.NewAPIOK("User deleted", nil).WriteTo(w)
}
