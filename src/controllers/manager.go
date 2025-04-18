// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/lfkeitel/verbose/v4"
	dhcp "github.com/packet-guardian/dhcp-lib"
	"github.com/packet-guardian/packet-guardian/src/auth"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
)

const (
	nonAdminAutoReg       = "na-auto"
	nonAdminManReg        = "na-man"
	nonAdminManRegNologin = "na-man-nologin"
	adminReg              = "admin"
	manualNotAllowed      = "man-not-allowed"
)

type Manager struct {
	e       *common.Environment
	users   stores.UserStore
	devices stores.DeviceStore
	leases  stores.LeaseStore
}

func NewManagerController(e *common.Environment, ds stores.DeviceStore, ls stores.LeaseStore, us stores.UserStore) *Manager {
	return &Manager{
		e:       e,
		devices: ds,
		leases:  ls,
		users:   us,
	}
}

func (m *Manager) RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	if m.e.Config.Guest.GuestOnly && !auth.IsLoggedIn(r) {
		http.Redirect(w, r, "/register/guest", http.StatusTemporaryRedirect)
		return
	}
	sessionUser := models.GetUserFromContext(r)
	man := (r.FormValue("manual") == "1")
	loggedIn := auth.IsLoggedIn(r)
	ip := common.GetIPFromContext(r)
	reg, _ := dhcp.IsRegisteredByIP(m.leases, ip)
	if !man && reg {
		http.Redirect(w, r, "/manage", http.StatusTemporaryRedirect)
		return
	}

	formType := nonAdminAutoReg
	if man {
		if !m.e.Config.Registration.AllowManualRegistrations && !sessionUser.Can(models.CreateDevice) {
			formType = manualNotAllowed
		} else {
			formType = nonAdminManReg
			if loggedIn {
				formType = nonAdminManRegNologin
			}
		}
	}

	username := sessionUser.Username
	if r.FormValue("username") != "" {
		if sessionUser.Can(models.CreateDevice) { // global admin
			username = r.FormValue("username")
			formType = adminReg
		} else {
			// Delegate user
			user, err := m.users.GetUserByUsername(r.FormValue("username"))
			if err == nil && user.DelegateCan(sessionUser.Username, models.CreateDevice) {
				username = r.FormValue("username")
				formType = nonAdminManRegNologin
			}
		}
	}

	data := map[string]interface{}{
		"policy":   common.LoadPolicyText(m.e.Config.Registration.RegistrationPolicyFile),
		"type":     formType,
		"username": strings.ToLower(username),
	}

	m.e.Views.NewView("user-register", r).Render(w, data)
}

func (m *Manager) DelegateManageHandler(w http.ResponseWriter, r *http.Request) {
	sessionUser := models.GetUserFromContext(r)

	// Redirect privileged users to the full-featured management page
	if sessionUser.Can(models.ViewDevices) {
		http.Redirect(w, r, "/admin/manage/"+sessionUser.Username, http.StatusTemporaryRedirect)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	username := pathParts[len(pathParts)-1]

	user, err := m.users.GetUserByUsername(username)
	if err != nil || !user.DelegateCan(sessionUser.Username, models.ViewDevices) {
		m.ManageHandler(w, r)
		return
	}

	pageNum := 1
	if page, _ := strconv.Atoi(r.URL.Query().Get("page")); page > 0 {
		pageNum = page
	}

	results, err := m.devices.GetDevicesForUserPage(user, pageNum)
	if err != nil {
		m.e.Log.WithFields(verbose.Fields{
			"error":    err,
			"package":  "controllers:manager",
			"username": user.Username,
		}).Error("Error getting devices")
		m.e.Views.RenderError(w, r, nil)
		return
	}

	deviceCnt, err := m.devices.GetDeviceCountForUser(user)
	if err != nil {
		m.e.Log.WithFields(verbose.Fields{
			"error":    err,
			"package":  "controllers:manager",
			"username": user.Username,
		}).Error("Error getting devices")
		m.e.Views.RenderError(w, r, nil)
		return
	}

	pageEnd := pageNum * common.PageSize
	if deviceCnt < pageEnd {
		pageEnd = deviceCnt
	}

	showAddBtn := (m.e.Config.Registration.AllowManualRegistrations && !user.IsBlacklisted())

	data := map[string]interface{}{
		"user":            sessionUser,
		"delegated":       sessionUser.Delegated(),
		"currentUser":     user.Username,
		"devices":         results,
		"deviceCnt":       deviceCnt,
		"usePages":        deviceCnt > common.PageSize,
		"page":            pageNum,
		"adminManage":     false,
		"pageStart":       ((pageNum - 1) * common.PageSize) + 1,
		"pageEnd":         pageEnd,
		"hasNextPage":     pageNum*common.PageSize < deviceCnt,
		"showAddBtn":      showAddBtn && user.DelegateCan(sessionUser.Username, models.CreateDevice) && !user.IsBlacklisted(),
		"canEditDevice":   user.DelegateCan(sessionUser.Username, models.EditDevice) && !sessionUser.IsBlacklisted(),
		"canDeleteDevice": user.DelegateCan(sessionUser.Username, models.DeleteDevice) && !sessionUser.IsBlacklisted(),
	}

	m.e.Views.NewView("user-manage", r).Render(w, data)
}

func (m *Manager) ManageHandler(w http.ResponseWriter, r *http.Request) {
	sessionUser := models.GetUserFromContext(r)

	// Redirect privileged users to the full-featured management page
	if sessionUser.Can(models.ViewDevices) {
		http.Redirect(w, r, "/admin/manage/"+sessionUser.Username, http.StatusTemporaryRedirect)
		return
	}

	pageNum := 1
	if page, _ := strconv.Atoi(r.URL.Query().Get("page")); page > 0 {
		pageNum = page
	}

	results, err := m.devices.GetDevicesForUserPage(sessionUser, pageNum)
	if err != nil {
		m.e.Log.WithFields(verbose.Fields{
			"error":    err,
			"package":  "controllers:manager",
			"username": sessionUser.Username,
		}).Error("Error getting devices")
		m.e.Views.RenderError(w, r, nil)
		return
	}

	deviceCnt, err := m.devices.GetDeviceCountForUser(sessionUser)
	if err != nil {
		m.e.Log.WithFields(verbose.Fields{
			"error":    err,
			"package":  "controllers:manager",
			"username": sessionUser.Username,
		}).Error("Error getting devices")
		m.e.Views.RenderError(w, r, nil)
		return
	}

	pageEnd := pageNum * common.PageSize
	if deviceCnt < pageEnd {
		pageEnd = deviceCnt
	}

	showAddBtn := (m.e.Config.Registration.AllowManualRegistrations && !sessionUser.IsBlacklisted())

	data := map[string]interface{}{
		"user":            sessionUser,
		"delegated":       sessionUser.Delegated(),
		"currentUser":     sessionUser.Username,
		"devices":         results,
		"deviceCnt":       deviceCnt,
		"usePages":        deviceCnt > common.PageSize,
		"page":            pageNum,
		"adminManage":     false,
		"pageStart":       ((pageNum - 1) * common.PageSize) + 1,
		"pageEnd":         pageEnd,
		"hasNextPage":     pageNum*common.PageSize < deviceCnt,
		"showAddBtn":      showAddBtn && sessionUser.Can(models.CreateOwn) && !sessionUser.IsBlacklisted(),
		"canEditDevice":   sessionUser.Can(models.EditOwn) && !sessionUser.IsBlacklisted(),
		"canDeleteDevice": sessionUser.Can(models.DeleteOwn) && !sessionUser.IsBlacklisted(),
	}

	m.e.Views.NewView("user-manage", r).Render(w, data)
}
