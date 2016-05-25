// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package controllers

import (
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/dhcp"
	"github.com/onesimus-systems/packet-guardian/src/models"
	"github.com/onesimus-systems/packet-guardian/src/stats"
)

type Admin struct {
	e *common.Environment
}

func NewAdminController(e *common.Environment) *Admin {
	return &Admin{e: e}
}

func (a *Admin) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	deviceTotal, deviceAvg := stats.GetDeviceStats(a.e)
	data := map[string]interface{}{
		"sessionUser": models.GetUserFromContext(r),
		"leaseStats":  stats.GetLeaseStats(a.e),
		"deviceTotal": deviceTotal,
		"deviceAvg":   deviceAvg,
	}
	a.e.Views.NewView("admin-dash", r).Render(w, data)
}

func (a *Admin) ManageHandler(w http.ResponseWriter, r *http.Request) {
	user, err := models.GetUserByUsername(a.e, mux.Vars(r)["username"])

	results, err := models.GetDevicesForUser(a.e, user)
	if err != nil {
		a.e.Log.Errorf("Error getting devices for user %s: %s", user.Username, err.Error())
		a.e.Views.RenderError(w, r, nil)
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
			lease, err := dhcp.GetLeaseByIP(a.e, ip)
			if err == nil {
				results, err = models.SearchDevicesByField(a.e, "mac", lease.MAC.String())
			}
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

func (a *Admin) AdminUserListHandler(w http.ResponseWriter, r *http.Request) {
	users, err := models.GetAllUsers(a.e)
	if err != nil {
		a.e.Log.Errorf("Error getting all users: %s", err.Error())
	}

	data := map[string]interface{}{
		"users": users,
	}

	a.e.Views.NewView("admin-user-list", r).Render(w, data)
}

func (a *Admin) AdminUserHandler(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	user, err := models.GetUserByUsername(a.e, username)
	if err != nil {
		a.e.Log.Errorf("Error getting user %s: %s", username, err.Error())
	}

	data := map[string]interface{}{
		"user": user,
	}

	a.e.Views.NewView("admin-user", r).Render(w, data)
}
