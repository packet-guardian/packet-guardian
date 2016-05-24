package controllers

import (
	"bufio"
	"html/template"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/onesimus-systems/packet-guardian/src/auth"
	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/dhcp"
	"github.com/onesimus-systems/packet-guardian/src/models"
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
	sessionUser := models.GetUserFromContext(r)
	man := (r.FormValue("manual") == "1")
	loggedIn := auth.IsLoggedIn(r)
	ip := net.ParseIP(strings.Split(r.RemoteAddr, ":")[0])
	reg, _ := dhcp.IsRegisteredByIP(m.e, ip)
	if !man && reg {
		http.Redirect(w, r, "/manage", http.StatusTemporaryRedirect)
		return
	}

	formType := nonAdminAutoReg
	if man {
		if !m.e.Config.Registration.AllowManualRegistrations && !sessionUser.IsAdmin() {
			formType = manualNotAllowed
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
	}

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
		m.e.Views.RenderError(w, r, nil)
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
