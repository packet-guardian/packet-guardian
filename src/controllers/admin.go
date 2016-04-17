package controllers

import (
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/models"
)

type Admin struct {
	e *common.Environment
}

func NewAdminController(e *common.Environment) *Admin {
	return &Admin{e: e}
}

func (a *Admin) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"sessionUser": models.GetUserFromContext(r),
	}
	a.e.Views.NewView("admin-dash", r).Render(w, data)
}

func (a *Admin) ManageHandler(w http.ResponseWriter, r *http.Request) {
	user, err := models.GetUserByUsername(a.e, mux.Vars(r)["username"])

	results, err := models.GetDevicesForUser(a.e, user)
	if err != nil {
		a.e.Log.Errorf("Error getting devices for user %s: %s", user.Username, err.Error())
		// TODO: Show error page to user
		return
	}

	data := make(map[string]interface{})
	data["user"] = user
	data["sessionUser"] = models.GetUserFromContext(r)
	data["devices"] = results

	a.e.Views.NewView("admin-manage", r).Render(w, data)
}

func (a *Admin) SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("q")
	var results []*models.Device
	var err error

	if query == "*" {
		results, err = models.SearchDevicesByField(a.e, "username", "%")
	} else if query != "" {
		if m, err := common.FormatMacAddress(query); err == nil {
			results, err = models.SearchDevicesByField(a.e, "mac", m.String())
		} else if ip := net.ParseIP(query); ip != nil {
			//results, err = models.SearchDevicesByField(a.e, "registred_from", ip.String())
			// TODO: Finish IP search when the leases system is implemented
		} else {
			results, err = models.SearchDevicesByField(a.e, "username", query+"%")
			if len(results) == 0 {
				results, err = models.SearchDevicesByField(a.e, "user_agent", "%"+query+"%")
			}
		}
	}

	if err != nil {
		a.e.Log.Errorf("Error getting search results: %s", err.Error())
	}

	data := map[string]interface{}{
		"query":         query,
		"searchResults": results,
	}

	a.e.Views.NewView("admin-search", r).Render(w, data)
}

//
// func (a *Admin) adminBlacklistHandler(w http.ResponseWriter, r *http.Request) {
// 	// Slice of MAC addresses
// 	var black []interface{}
//
// 	if r.FormValue("devices") != "" {
// 		deviceIDs := strings.Split(r.FormValue("devices"), ",")
// 		// Need to convert strings to int for database search
// 		ids := make([]int, len(deviceIDs))
// 		for i := range deviceIDs {
// 			in, _ := strconv.Atoi(deviceIDs[i])
// 			ids[i] = in
// 		}
//
// 		devices := dhcp.Query{ID: ids}.Search(a.e)
// 		for i := range devices {
// 			black = append(black, devices[i].MAC)
// 		}
// 	} else {
// 		username, ok := mux.Vars(r)["username"]
// 		if !ok {
// 			common.NewAPIResponse(common.APIStatusGenericError, "No username given", nil).WriteTo(w)
// 			return
// 		}
// 		black = append(black, username)
//
// 		splitPath := strings.Split(r.URL.Path, "/")
// 		if splitPath[len(splitPath)-1] == "all" {
// 			results := dhcp.Query{User: username}.Search(a.e)
// 			for _, r := range results {
// 				black = append(black, r.MAC)
// 			}
// 		}
// 	}
//
// 	if r.Method == "DELETE" {
// 		err := dhcp.RemoveFromBlacklist(a.e.DB, black...)
// 		if err != nil {
// 			a.e.Log.Errorf("Error removing from blacklist: %s", err.Error())
// 			common.NewAPIResponse(common.APIStatusGenericError, "Error removing from blacklist", nil).WriteTo(w)
// 			return
// 		}
// 		for _, d := range black {
// 			a.e.Log.Infof("Removed user/MAC %s from blacklist", d)
// 		}
// 		common.NewAPIOK("Unblacklisting successful", nil).WriteTo(w)
// 	} else {
// 		err := dhcp.AddToBlacklist(a.e.DB, black...)
// 		if err != nil {
// 			a.e.Log.Errorf("Error blacklisting: %s", err.Error())
// 			common.NewAPIResponse(common.APIStatusGenericError, "Error blacklisting", nil).WriteTo(w)
// 			return
// 		}
// 		for _, d := range black {
// 			a.e.Log.Infof("Blacklisted user/MAC %s", d)
// 		}
// 		common.NewAPIOK("Blacklisting successful", nil).WriteTo(w)
// 	}
// }
//
// func (a *Admin) adminUserHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method == "POST" {
// 		a.saveUserHandler(w, r)
// 		return
// 	} else if r.Method == "DELETE" {
// 		a.deleteUserHandler(w, r)
// 		return
// 	}
//
// 	data := struct {
// 		Query        string
// 		Users        []*models.User
// 		FlashMessage string
// 	}{}
//
// 	username := mux.Vars(r)["username"]
// 	var template string
// 	if username == "" {
// 		users, err := models.GetAllUsers(a.e)
// 		if err != nil {
// 			a.e.Log.Errorf("Error getting users: %s", err.Error())
// 			data.FlashMessage = "Error getting users"
// 		}
// 		data.Users = users
// 		template = "admin-users"
// 	} else {
// 		user, _ := models.GetUserByUsername(a.e, username)
// 		if user.ID == 0 {
// 			user.Username = username
// 		}
// 		data.Users = []*models.User{user}
// 		template = "admin-user"
// 	}
//
// 	if err := a.e.Views.NewView(template).Render(w, data); err != nil {
// 		a.e.Log.Error(err.Error())
// 	}
// }
//
// func (a *Admin) saveUserHandler(w http.ResponseWriter, r *http.Request) {
// 	username := r.FormValue("username")
// 	// Get or create user
// 	user, _ := models.GetUserByUsername(a.e, username)
// 	if user == nil {
// 		a.e.Log.Info("Creating user")
// 		user = &models.User{
// 			ID:       common.ConvertToInt(r.FormValue("user-id")),
// 			Username: username,
// 		}
// 	}
//
// 	// Password
// 	user.ClearPassword = (r.FormValue("clear-pass") == "true")
// 	if r.FormValue("password") != "" {
// 		user.NewPassword(r.FormValue("password"))
// 	}
//
// 	// Registered device limit
// 	limitType := r.FormValue("special-limit")
// 	if limitType == "global" {
// 		user.DeviceLimit = -1
// 	} else if limitType == "unlimited" {
// 		user.DeviceLimit = 0
// 	} else {
// 		user.DeviceLimit = common.ConvertToInt(r.FormValue("device-limit"))
// 	}
//
// 	// Expiration times
// 	loc, _ := time.LoadLocation("Local")
// 	if r.FormValue("device-expiration") == "0" || r.FormValue("device-expiration") == "" {
// 		user.DefaultExpiration = time.Unix(0, 0)
// 	} else if r.FormValue("device-expiration") == "1" {
// 		user.DefaultExpiration = time.Unix(1, 0)
// 	} else {
// 		user.DefaultExpiration, _ = time.ParseInLocation("2006-01-02 15:04:05", r.FormValue("device-expiration"), loc)
// 	}
//
// 	if r.FormValue("valid-after") == "0" || r.FormValue("valid-after") == "" {
// 		user.ValidAfter = time.Unix(0, 0)
// 	} else {
// 		user.ValidAfter, _ = time.ParseInLocation("2006-01-02 15:04:05", r.FormValue("valid-after"), loc)
// 	}
//
// 	if r.FormValue("valid-before") == "0" || r.FormValue("valid-before") == "" {
// 		user.ValidBefore = time.Unix(0, 0)
// 	} else {
// 		user.ValidBefore, _ = time.ParseInLocation("2006-01-02 15:04:05", r.FormValue("valid-before"), loc)
// 	}
//
// 	if err := user.Save(a.e.DB); err != nil {
// 		a.e.Log.Errorf("Error saving user: %s", err.Error())
// 		common.NewAPIResponse(common.APIStatusGenericError, "Error saving user", nil).WriteTo(w)
// 		return
// 	}
//
// 	a.e.Log.Infof("Created user augmentation: %s", user.Username)
// 	common.NewAPIOK("User created", nil).WriteTo(w)
// }
//
// func (a *Admin) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
// 	username := mux.Vars(r)["username"]
//
// 	user, err := models.GetUser(a.e.DB, username)
// 	if user == nil {
// 		a.e.Log.Errorf("Error deleting user: %s", err.Error())
// 		common.NewAPIResponse(common.APIStatusGenericError, "Error deleting user", nil).WriteTo(w)
// 		return
// 	}
//
// 	sql := "DELETE FROM \"user\" WHERE \"username\" = ?"
// 	_, err = a.e.DB.Exec(sql, username)
// 	if err != nil {
// 		a.e.Log.Errorf("Error deleting user: %s", err.Error())
// 		common.NewAPIResponse(common.APIStatusGenericError, "Error deleting user", nil).WriteTo(w)
// 		return
// 	}
//
// 	a.e.Log.Infof("Deleted user: %s", username)
// 	common.NewAPIOK("User deleted", nil).WriteTo(w)
// }
