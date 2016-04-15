package controllers

import (
	"bufio"
	"html/template"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/onesimus-systems/packet-guardian/src/auth"
	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/dhcp"
	"github.com/onesimus-systems/packet-guardian/src/server/middleware"
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

func (m *Manager) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/register", m.registrationHandler).Methods("GET")
	r.HandleFunc("/manage", middleware.CheckAuth(m.e, m.manageHandler)).Methods("GET")
	r.HandleFunc("/admin/user/{username}", middleware.CheckAdmin(m.e, m.manageHandler)).Methods("GET")
}

func (m *Manager) registrationHandler(w http.ResponseWriter, r *http.Request) {
	flash := ""
	man := (r.FormValue("manual") == "1")
	loggedIn := auth.IsLoggedIn(m.e, r)
	ip := net.ParseIP(strings.Split(r.RemoteAddr, ":")[0])
	reg, _ := dhcp.IsRegisteredByIP(m.e.DB, ip)
	if !man && reg {
		http.Redirect(w, r, "/manage", http.StatusTemporaryRedirect)
		return
	}

	formType := nonAdminAutoReg
	if man {
		if !m.e.Config.Core.AllowManualRegistrations {
			flash = "Manual registrations are not allowed"
		} else {
			formType = nonAdminManReg
			if loggedIn {
				formType = nonAdminManRegNologin
			}
		}
	}

	username := m.e.Sessions.GetSession(r).GetString("username")
	if r.FormValue("username") != "" && auth.IsAdminUser(m.e, r) {
		username = r.FormValue("username")
		formType = "admin"
		flash = ""
	}

	// TODO: Don't show the registration page at all if blacklisted
	if formType == nonAdminAutoReg {
		ip := net.ParseIP(strings.Split(r.RemoteAddr, ":")[0])
		mac, err := dhcp.GetMacFromIP(ip)
		if err != nil {
			m.e.Log.Errorf("Failed to get MAC from IP for %s", ip.String())
		} else {
			bl, err := dhcp.IsBlacklisted(m.e.DB, mac.String())
			if err != nil {
				m.e.Log.Errorf("There was an error checking the blacklist for MAC %s", mac.String())
			}
			if bl {
				flash = "The device appears to be blacklisted"
			}
		}
	}

	data := struct {
		Policy       []template.HTML
		Type         string
		Username     string
		FlashMessage string
	}{
		Policy:       m.loadPolicyText(),
		Type:         formType,
		Username:     username,
		FlashMessage: flash,
	}

	if err := m.e.Views.NewView("register").Render(w, data); err != nil {
		m.e.Log.Error(err.Error())
	}
}

func (m *Manager) manageHandler(w http.ResponseWriter, r *http.Request) {
	username, ok := mux.Vars(r)["username"]
	if !ok {
		username = m.e.Sessions.GetSession(r).GetString("username")
	}
	results := dhcp.Query{User: username}.Search(m.e)

	admin := auth.IsAdminUser(m.e, r)
	bl, _ := dhcp.IsBlacklisted(m.e.DB, username)
	showAddBtn := (admin || (m.e.Config.Core.AllowManualRegistrations && !bl))

	data := struct {
		Username      string
		IsAdmin       bool
		Devices       []dhcp.Device
		FlashMessage  string
		ShowAddBtn    bool
		IsBlacklisted bool
	}{
		Username:      username,
		IsAdmin:       auth.IsAdminUser(m.e, r),
		Devices:       results,
		ShowAddBtn:    showAddBtn,
		IsBlacklisted: bl,
	}

	if err := m.e.Views.NewView("manage").Render(w, data); err != nil {
		m.e.Log.Error(err.Error())
	}
}

func (m *Manager) loadPolicyText() []template.HTML {
	f, err := os.Open(m.e.Config.Core.RegistrationPolicyFile)
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
