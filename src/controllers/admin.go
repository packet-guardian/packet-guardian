// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package controllers

import (
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/lfkeitel/verbose/v4"
	dhcp "github.com/packet-guardian/dhcp-lib"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
	"github.com/packet-guardian/packet-guardian/src/reports"
	"github.com/packet-guardian/packet-guardian/src/stats"
)

var (
	ipStartRegex  = regexp.MustCompile(`^[0-9]{1,3}\.`)
	macStartRegex = regexp.MustCompile(`^([0-9a-fA-F]{2}[\:\-]|[0-9a-fA-F]{4}\.)`)
)

type Admin struct {
	e      *common.Environment
	stores stores.StoreCollection
}

func NewAdminController(e *common.Environment, stores stores.StoreCollection) *Admin {
	return &Admin{
		e:      e,
		stores: stores,
	}
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

	user, err := a.stores.Users.GetUserByUsername(p.ByName("username"))
	if err != nil {
		a.e.Log.WithFields(verbose.Fields{
			"error":    err,
			"package":  "controllers:admin",
			"username": p.ByName("username"),
		}).Error("Error getting user from database")
		a.e.Views.RenderError(w, r, nil)
		return
	}

	pageNum := 1
	if page, _ := strconv.Atoi(r.URL.Query().Get("page")); page > 0 {
		pageNum = page
	}

	results, err := a.stores.Devices.GetDevicesForUserPage(user, pageNum)
	if err != nil {
		a.e.Log.WithFields(verbose.Fields{
			"error":    err,
			"package":  "controllers:admin",
			"username": user.Username,
		}).Error("Error getting devices")
		a.e.Views.RenderError(w, r, nil)
		return
	}

	deviceCnt, err := a.stores.Devices.GetDeviceCountForUser(user)
	if err != nil {
		a.e.Log.WithFields(verbose.Fields{
			"error":    err,
			"package":  "controllers:admin",
			"username": user.Username,
		}).Error("Error getting devices")
		a.e.Views.RenderError(w, r, nil)
		return
	}

	pageEnd := pageNum * common.PageSize
	if deviceCnt < pageEnd {
		pageEnd = deviceCnt
	}

	data := map[string]interface{}{
		"user":               user,
		"devices":            results,
		"deviceCnt":          deviceCnt,
		"usePages":           deviceCnt > common.PageSize,
		"page":               pageNum,
		"hasNextPage":        pageNum*common.PageSize < deviceCnt,
		"adminManage":        true,
		"pageStart":          ((pageNum - 1) * common.PageSize) + 1,
		"pageEnd":            pageEnd,
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
	device, err := a.stores.Devices.GetDeviceByMAC(mac)
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
	user, err := a.stores.Users.GetUserByUsername(device.Username)
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
	U string
}

func (a *Admin) SearchHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sessionUser := models.GetUserFromContext(r)
	if !sessionUser.Can(models.ViewAdminPage | models.ViewDevices) {
		a.redirectToRoot(w, r)
		return
	}

	query := r.FormValue("q")
	var results []*searchResults
	var searchType string
	var err error

	if query != "" {
		results, searchType, err = a.search(query)
		if searchType == "user" && len(results) == 1 {
			http.Redirect(w, r, "/admin/manage/user/"+url.QueryEscape(results[0].U), http.StatusTemporaryRedirect)
			return
		}
	}

	for _, r := range results {
		if r.L != nil {
			continue
		}
		lease, err := a.stores.Leases.GetRecentLeaseByMAC(r.D.MAC)
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

func (a *Admin) search(query string) ([]*searchResults, string, error) {
	if macStartRegex.MatchString(query) {
		return a.macSearch(query)
	} else if ipStartRegex.MatchString(query) {
		return a.ipSearch(query)
	}
	return a.userSearch(query)
}

func (a *Admin) macSearch(query string) ([]*searchResults, string, error) {
	mac, _ := common.FormatMacAddress(query)

	var results []*searchResults
	devices, err := a.stores.Devices.SearchDevicesByField("mac", "%"+mac.String()+"%")
	for _, d := range devices {
		results = append(results, &searchResults{
			D: d,
		})
	}
	return results, "mac", err
}

func (a *Admin) ipSearch(query string) ([]*searchResults, string, error) {
	var results []*searchResults
	// Get leases matching IP
	var leases []*dhcp.Lease
	leases, err := a.stores.Leases.SearchLeases(`"ip" LIKE ?`, "%"+query+"%")
	// Get devices corresponding to each lease
	var d *models.Device
	for _, l := range leases {
		d, err = a.stores.Devices.GetDeviceByMAC(l.MAC)
		if err != nil {
			continue
		}
		results = append(results, &searchResults{
			D: d,
			L: l,
		})
	}
	return results, "ip", err
}

func (a *Admin) userSearch(query string) ([]*searchResults, string, error) {
	// Search for devices with the username
	exact := true
	devices, err := a.stores.Devices.SearchDevicesByField("username", "%"+query+"%")
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
		return []*searchResults{&searchResults{U: query}}, "user", nil
	}

	// All else fails, search the user agent for the query
	if len(devices) == 0 {
		devices, err = a.stores.Devices.SearchDevicesByField("user_agent", "%"+query+"%")
	}

	results := make([]*searchResults, len(devices))
	for i, d := range devices {
		results[i] = &searchResults{
			D: d,
		}
	}
	return results, "user", err
}

func (a *Admin) AdminUserListHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sessionUser := models.GetUserFromContext(r)
	if !sessionUser.Can(models.ViewUsers) {
		a.redirectToRoot(w, r)
		return
	}

	users, err := a.stores.Users.GetAllUsers()
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
	user, err := a.stores.Users.GetUserByUsername(username)
	if err != nil {
		a.e.Log.WithFields(verbose.Fields{
			"error":    err,
			"package":  "controllers:admin",
			"username": username,
		}).Error("Error getting user")
	}

	data := map[string]interface{}{
		"user":        user,
		"delegateFor": user.Delegated(),
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
		reports.RenderReport(report, w, r, a.stores)
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

func (a *Admin) RenderImportExportPage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	a.e.Views.NewView("admin-import-export", r).Render(w, nil)
}

func (a *Admin) Import(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	resource := p.ByName("resource")

	switch resource {
	case "devices":
		a.importDevices(w, r)
	default:
		common.NewAPIResponse("", nil).WriteResponse(w, http.StatusNotFound)
	}
}

func (a *Admin) importDevices(w http.ResponseWriter, r *http.Request) {
	session := common.GetSessionFromContext(r)
	sessionUser := models.GetUserFromContext(r)
	ip := common.GetIPFromContext(r)

	csvr := csv.NewReader(strings.NewReader(r.PostFormValue("import-data")))
	csvr.Comment = '#'
	csvr.ReuseRecord = true
	csvr.TrimLeadingSpace = true
	csvr.FieldsPerRecord = 4

	var record []string
	var err error
	userCache := make(map[string]*models.User)

	errors := make([]string, 0, 5)

	record, err = csvr.Read() // header
	if !common.StringSliceEqual(record, []string{"username", "mac", "description", "platform"}) {
		session.AddFlash(common.FlashMessage{
			Message: "Import data missing CSV headers",
			Type:    common.FlashMessageError,
		})
		a.e.Views.NewView("admin-import-export", r).Render(w, nil)
		return
	}

	record, err = csvr.Read()
	for err != io.EOF {
		var username string
		var description string
		var platform string
		var mac net.HardwareAddr
		var device *models.Device
		var deviceUser *models.User

		if err != nil {
			errors = append(errors, err.Error())
			a.e.Log.Error(err.Error())
			goto next
		}

		username = record[0]
		description = record[2]
		platform = record[3]
		mac, err = common.FormatMacAddress(record[1])
		if err != nil {
			errors = append(errors, fmt.Sprintf("Invalid MAC address '%s'", record[1]))
			a.e.Log.Errorf("Invalid MAC address '%s'", record[1])
			goto next
		}

		device, err = a.stores.Devices.GetDeviceByMAC(mac)
		if err != nil {
			a.e.Log.WithFields(verbose.Fields{
				"error":   err,
				"package": "controllers:admin:importDevices",
				"mac":     mac.String(),
			}).Error("Error getting device")
			goto next
		}

		if device.ID != 0 {
			errors = append(errors, fmt.Sprintf("Device already registered '%s'", record[1]))
			goto next
		}

		deviceUser = userCache[username]
		if deviceUser == nil {
			deviceUser, err = a.stores.Users.GetUserByUsername(username)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Error getting user '%s'", username))
				a.e.Log.WithFields(verbose.Fields{
					"error":    err,
					"package":  "controllers:admin:importDevices",
					"username": username,
				}).Error("Error getting user")
				goto next
			}
			userCache[username] = deviceUser
		}

		device.Username = username
		device.Description = description
		device.RegisteredFrom = ip
		device.Platform = platform
		device.Expires = deviceUser.DeviceExpiration.NextExpiration(a.e, time.Now())
		device.DateRegistered = time.Now()
		device.LastSeen = time.Now()
		device.UserAgent = "Manual"

		if err = device.Save(); err != nil {
			errors = append(errors, fmt.Sprintf("Error saving device '%s'", record[1]))
			a.e.Log.WithFields(verbose.Fields{
				"error":   err,
				"package": "controllers:admin:importDevices",
			}).Error("Error saving device")
			goto next
		}
		a.e.Log.WithFields(verbose.Fields{
			"package":    "controllers:admin:importDevices",
			"mac":        mac.String(),
			"changed-by": sessionUser.Username,
			"username":   username,
			"action":     "register_device",
			"manual":     true,
		}).Info("Device registered")

	next:
		record, err = csvr.Read()
	}

	if len(errors) != 0 {
		session.AddFlash(common.FlashMessage{
			Message: "Error: " + strings.Join(errors, "<br>"),
			Type:    common.FlashMessageError,
		})
	} else {
		session.AddFlash(common.FlashMessage{Message: "Import Successful"})
	}
	a.e.Views.NewView("admin-import-export", r).Render(w, nil)
}
