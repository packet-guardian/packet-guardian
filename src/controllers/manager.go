package controllers

import (
	"bufio"
	"html/template"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/dragonrider23/verbose"
	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/dhcp"
	"github.com/onesimus-systems/packet-guardian/src/models"
)

const (
	nonAdminAutoReg       = "na-auto"
	nonAdminManReg        = "na-man"
	nonAdminManRegNologin = "na-man-nologin"
	adminReg              = "admin"
)

type Manager struct {
	e *common.Environment
}

func NewManagerController(e *common.Environment) *Manager {
	return &Manager{e: e}
}

func (m *Manager) RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	sessionUser := models.GetUserFromContext(r)
	session := common.GetSessionFromContext(r)
	flash := ""
	man := (r.FormValue("manual") == "1")
	loggedIn := session.GetBool("loggedin")
	ip := net.ParseIP(strings.Split(r.RemoteAddr, ":")[0])
	reg, _ := dhcp.IsRegisteredByIP(m.e, ip)
	if !man && reg {
		http.Redirect(w, r, "/manage", http.StatusTemporaryRedirect)
		return
	}

	formType := nonAdminAutoReg
	if man {
		if !m.e.Config.Registration.AllowManualRegistrations {
			flash = "Manual registrations are not allowed"
		} else {
			formType = nonAdminManReg
			if loggedIn {
				formType = nonAdminManRegNologin
			}
		}
	}

	username := sessionUser.Username
	if r.FormValue("username") != "" && sessionUser.IsAdmin() {
		username = r.FormValue("username")
		formType = adminReg
		flash = ""
	}

	// TODO: Don't show the registration page at all if blacklisted
	if formType == nonAdminAutoReg {
		lease, err := dhcp.GetLeaseByIP(m.e, ip)
		if err != nil || lease.ID == 0 {
			m.e.Log.WithFields(verbose.Fields{
				"IP":    ip.String(),
				"Lease": lease,
				"Error": err,
			}).Error("Failed to get MAC from IP")
		} else {
			device, err := models.GetDeviceByMAC(m.e, lease.MAC)
			if err != nil {
				m.e.Log.Errorf("Error getting device for reg check: %s", err.Error())
			} else {
				if device.IsBlacklisted {
					flash = "The device appears to be blacklisted"
				}
			}
		}
	}

	session.AddFlash(flash)
	session.Save(r, w)
	data := map[string]interface{}{
		"policy":   m.loadPolicyText(),
		"type":     formType,
		"username": username,
	}

	m.e.Views.NewView("register", r).Render(w, data)
}

func (m *Manager) ManageHandler(w http.ResponseWriter, r *http.Request) {
	sessionUser := models.GetUserFromContext(r)

	if sessionUser.IsHelpDesk() {
		http.Redirect(w, r, "/admin/manage/"+sessionUser.Username, http.StatusTemporaryRedirect)
		return
	}

	results, err := models.GetDevicesForUser(m.e, sessionUser)
	if err != nil {
		m.e.Log.Errorf("Error getting devices for user %s: %s", sessionUser.Username, err.Error())
		// TODO: Show error page to user
		return
	}

	showAddBtn := (m.e.Config.Registration.AllowManualRegistrations && !sessionUser.IsBlacklisted())

	data := make(map[string]interface{})
	data["sessionUser"] = sessionUser
	data["devices"] = results
	data["showAddBtn"] = showAddBtn

	m.e.Views.NewView("manage", r).Render(w, data)
}

func (m *Manager) loadPolicyText() []template.HTML {
	f, err := os.Open(m.e.Config.Registration.RegistrationPolicyFile)
	if err != nil {
		return nil
	}
	defer f.Close()

	var policy []template.HTML
	currentParagraph := ""
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		t := strings.TrimSpace(scanner.Text())
		if t == "" {
			policy = append(policy, template.HTML(currentParagraph))
			currentParagraph = ""
			continue
		}
		currentParagraph += " " + t
	}
	policy = append(policy, template.HTML(currentParagraph))
	return policy
}
