// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package controllers

import (
	"net/http"
	"strings"

	"github.com/lfkeitel/verbose/v4"
	"github.com/packet-guardian/dhcp-lib"
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
	e *common.Environment
}

func NewManagerController(e *common.Environment) *Manager {
	return &Manager{e: e}
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
	reg, _ := dhcp.IsRegisteredByIP(stores.GetLeaseStore(m.e), ip)
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
	if r.FormValue("username") != "" && sessionUser.Can(models.CreateDevice) {
		username = r.FormValue("username")
		formType = adminReg
	}

	data := map[string]interface{}{
		"policy":   common.LoadPolicyText(m.e.Config.Registration.RegistrationPolicyFile),
		"type":     formType,
		"username": strings.ToLower(username),
	}

	m.e.Views.NewView("register", r).Render(w, data)
}

func (m *Manager) ManageHandler(w http.ResponseWriter, r *http.Request) {
	sessionUser := models.GetUserFromContext(r)

	// Redirect privileged users to the full-featured management page
	if sessionUser.Can(models.ViewDevices) {
		http.Redirect(w, r, "/admin/manage/"+sessionUser.Username, http.StatusTemporaryRedirect)
		return
	}

	results, err := stores.GetDeviceStore(m.e).GetDevicesForUser(sessionUser)
	if err != nil {
		m.e.Log.WithFields(verbose.Fields{
			"error":    err,
			"package":  "controllers:manager",
			"username": sessionUser.Username,
		}).Error("Error getting devices")
		m.e.Views.RenderError(w, r, nil)
		return
	}

	showAddBtn := (m.e.Config.Registration.AllowManualRegistrations && !sessionUser.IsBlacklisted())

	data := map[string]interface{}{
		"sessionUser":     sessionUser,
		"devices":         results,
		"showAddBtn":      showAddBtn && sessionUser.Can(models.CreateOwn) && !sessionUser.IsBlacklisted(),
		"canEditDevice":   sessionUser.Can(models.EditOwn) && !sessionUser.IsBlacklisted(),
		"canDeleteDevice": sessionUser.Can(models.DeleteOwn) && !sessionUser.IsBlacklisted(),
	}

	m.e.Views.NewView("manage", r).Render(w, data)
}
