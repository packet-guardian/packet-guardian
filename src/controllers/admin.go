// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package controllers

import (
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"

	"github.com/julienschmidt/httprouter"
	"github.com/lfkeitel/verbose"
	"github.com/usi-lfkeitel/packet-guardian/src/common"
	"github.com/usi-lfkeitel/packet-guardian/src/models"
	"github.com/usi-lfkeitel/packet-guardian/src/models/stores"
	"github.com/usi-lfkeitel/packet-guardian/src/reports"
	"github.com/usi-lfkeitel/packet-guardian/src/stats"
	"github.com/usi-lfkeitel/pg-dhcp"
)

var (
	ipStartRegex  = regexp.MustCompile(`^[0-9]{1,3}\.`)
	macStartRegex = regexp.MustCompile(`^[0-f]{2}\:`)
)

type Admin struct {
	e *common.Environment
}

func NewAdminController(e *common.Environment) *Admin {
	return &Admin{e: e}
}

func (a *Admin) redirectToRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (a *Admin) DashboardHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sessionUser := models.GetUserFromContext(r)
	if !sessionUser.Can(models.ViewAdminPage) {
		a.redirectToRoot(w, r)
		return
	}

	deviceTotal, deviceAvg := stats.GetDeviceStats(a.e)
	data := map[string]interface{}{
		"canViewUsers": sessionUser.Can(models.ViewUsers),
		"leaseStats":   stats.GetLeaseStats(a.e),
		"deviceTotal":  deviceTotal,
		"deviceAvg":    deviceAvg,
	}
	a.e.Views.NewView("admin-dash", r).Render(w, data)
}

func (a *Admin) ManageHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sessionUser := models.GetUserFromContext(r)
	if !sessionUser.Can(models.ViewDevices) {
		a.redirectToRoot(w, r)
		return
	}

	user, err := stores.GetUserStore(a.e).GetUserByUsername(p.ByName("username"))

	results, err := stores.GetDeviceStore(a.e).GetDevicesForUser(user)
	if err != nil {
		a.e.Log.WithFields(verbose.Fields{
			"error":    err,
			"package":  "controllers:admin",
			"username": user.Username,
		}).Error("Error getting devices")
		a.e.Views.RenderError(w, r, nil)
		return
	}

	data := map[string]interface{}{
		"user":               user,
		"devices":            results,
		"canCreateDevice":    sessionUser.Can(models.CreateDevice),
		"canEditDevice":      sessionUser.Can(models.EditDevice),
		"canDeleteDevice":    sessionUser.Can(models.DeleteDevice),
		"canReassignDevice":  sessionUser.Can(models.ReassignDevice),
		"canManageBlacklist": sessionUser.Can(models.ManageBlacklist),
		"canEditUser":        sessionUser.Can(models.EditUser),
	}

	a.e.Views.NewView("admin-manage", r).Render(w, data)
}

func (a *Admin) ShowDeviceHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sessionUser := models.GetUserFromContext(r)
	if !sessionUser.Can(models.ViewDevices) {
		a.redirectToRoot(w, r)
		return
	}

	mac, err := net.ParseMAC(p.ByName("mac"))
	if err != nil {
		a.e.Views.RenderError(w, r, map[string]interface{}{
			"title": "No device found",
			"body":  "Incorrectly formed MAC address: " + p.ByName("mac"),
		})
		return
	}
	device, err := stores.GetDeviceStore(a.e).GetDeviceByMAC(mac)
	if err != nil {
		a.e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "controllers:admin",
			"mac":     mac.String(),
		}).Error("Error getting device")
		a.e.Views.RenderError(w, r, nil)
		return
	}
	device.LoadLeaseHistory()
	user, err := stores.GetUserStore(a.e).GetUserByUsername(device.Username)
	if err != nil {
		a.e.Log.WithFields(verbose.Fields{
			"error":    err,
			"package":  "controllers:admin",
			"username": device.Username,
		}).Error("Error getting user")
		a.e.Views.RenderError(w, r, nil)
		return
	}

	data := map[string]interface{}{
		"user":               user,
		"device":             device,
		"canEditDevice":      sessionUser.Can(models.EditDevice),
		"canDeleteDevice":    sessionUser.Can(models.DeleteDevice),
		"canReassignDevice":  sessionUser.Can(models.ReassignDevice),
		"canManageBlacklist": sessionUser.Can(models.ManageBlacklist),
		"canEditUser":        sessionUser.Can(models.EditUser),
	}

	a.e.Views.NewView("admin-manage-device", r).Render(w, data)
}

type searchResults struct {
	D *models.Device
	L *dhcp.Lease
}

func (a *Admin) SearchHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sessionUser := models.GetUserFromContext(r)
	if !sessionUser.Can(models.ViewAdminPage | models.ViewDevices) {
		a.redirectToRoot(w, r)
		return
	}

	query := r.FormValue("q")
	leaseStore := stores.GetLeaseStore(a.e)
	var results []*searchResults
	var devices []*models.Device
	var searchType string
	var err error

	if query != "" {
		if macStartRegex.MatchString(query) {
			searchType = "mac"
			devices, err = stores.GetDeviceStore(a.e).SearchDevicesByField("mac", "%"+query+"%")
			if len(devices) == 1 {
				http.Redirect(w, r,
					"/admin/manage/device/"+url.QueryEscape(devices[0].GetMAC().String()),
					http.StatusTemporaryRedirect,
				)
				return
			}
			for _, d := range devices {
				results = append(results, &searchResults{
					D: d,
				})
			}
		} else if ipStartRegex.MatchString(query) {
			searchType = "ip"
			// Get leases matching IP
			var leases []*dhcp.Lease
			leases, err = leaseStore.SearchLeases(`"ip" LIKE ?`, "%"+query+"%")
			// Get devices corresponding to each lease
			var d *models.Device
			for _, l := range leases {
				d, err = stores.GetDeviceStore(a.e).GetDeviceByMAC(l.MAC)
				if err != nil {
					continue
				}
				results = append(results, &searchResults{
					D: d,
					L: l,
				})
			}
		} else {
			searchType = "user"
			// Search for a local user account
			var users []*models.User
			users, err = stores.GetUserStore(a.e).SearchUsersByField("username", "%"+query+"%")
			if len(users) == 1 {
				http.Redirect(w, r,
					"/admin/manage/user/"+url.QueryEscape(users[0].Username),
					http.StatusTemporaryRedirect,
				)
				return
			}

			// Search for devices with the username
			exact := true
			devices, err = stores.GetDeviceStore(a.e).SearchDevicesByField("username", "%"+query+"%")
			if len(devices) == 0 {
				exact = false
			}
			for _, d := range devices { // Check if all the devices have the same username
				if d.GetUsername() != query {
					exact = false
					break
				}
			}

			if exact { // If they're all the same user, go directly to the user's page
				http.Redirect(w, r,
					"/admin/manage/user/"+url.QueryEscape(query),
					http.StatusTemporaryRedirect,
				)
				return
			}

			// All else fails, search the user agent for the query
			if len(devices) == 0 {
				devices, err = stores.GetDeviceStore(a.e).SearchDevicesByField("user_agent", "%"+query+"%")
			}

			for _, d := range devices {
				results = append(results, &searchResults{
					D: d,
				})
			}
		}
	}

	for _, r := range results {
		if r.L != nil {
			continue
		}
		lease, err := leaseStore.GetRecentLeaseByMAC(r.D.MAC)
		if err != nil || lease.ID == 0 {
			continue
		}
		r.L = lease
	}

	if err != nil {
		a.e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "controllers:admin",
		}).Error("Error getting search results")
	}

	data := map[string]interface{}{
		"query":      query,
		"results":    results,
		"searchType": searchType,
	}

	a.e.Views.NewView("admin-search", r).Render(w, data)
}

func (a *Admin) AdminUserListHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sessionUser := models.GetUserFromContext(r)
	if !sessionUser.Can(models.ViewUsers) {
		a.redirectToRoot(w, r)
		return
	}

	users, err := stores.GetUserStore(a.e).GetAllUsers()
	if err != nil {
		a.e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "controllers:admin",
		}).Error("Error getting users")
	}

	data := map[string]interface{}{
		"users":         users,
		"canEditUser":   sessionUser.Can(models.EditUser),
		"canCreateUser": sessionUser.Can(models.CreateUser),
	}

	a.e.Views.NewView("admin-user-list", r).Render(w, data)
}

func (a *Admin) AdminUserHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sessionUser := models.GetUserFromContext(r)
	if !sessionUser.Can(models.EditUser) {
		a.redirectToRoot(w, r)
		return
	}

	username := p.ByName("username")
	user, err := stores.GetUserStore(a.e).GetUserByUsername(username)
	if err != nil {
		a.e.Log.WithFields(verbose.Fields{
			"error":    err,
			"package":  "controllers:admin",
			"username": username,
		}).Error("Error getting user")
	}

	data := map[string]interface{}{
		"user": user,
	}

	a.e.Views.NewView("admin-user", r).Render(w, data)
}

func (a *Admin) ReportHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sessionUser := models.GetUserFromContext(r)
	if !sessionUser.Can(models.ViewReports) {
		a.redirectToRoot(w, r)
		return
	}

	report := p.ByName("report")
	if report != "" {
		reports.RenderReport(report, w, r)
		return
	}

	allReports := reports.GetReports()
	i := 0
	reportSlice := make([]*reports.Report, len(allReports))
	for _, v := range allReports {
		reportSlice[i] = v
		i++
	}

	sort.Sort(reports.ReportSorter(reportSlice))

	data := map[string]interface{}{
		"reports": reportSlice,
	}

	a.e.Views.NewView("admin-reports", r).Render(w, data)
}
