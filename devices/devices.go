package devices

import (
	"net/http"

	"github.com/onesimus-systems/net-guardian/auth"
	"github.com/onesimus-systems/net-guardian/common"
)

// IsRegistered checks if a MAC address is registed in the database
func IsRegistered(mac string) bool {
	return false
}

// GetMacFromIP finds the mac address that has the lease ip
func GetMacFromIP(ip string) string {
	return ""
}

// Register a new device to a user
func Register(mac, user, platform, ip, ua, subnet string) error {
	return nil
}

// IsBlacklisted checks if a username or MAC is in the blacklist
func IsBlacklisted(value string) bool {
	return false
}

// RegisterHTTPHandler serves and handles the registration page for an end user
func RegisterHTTPHandler(w http.ResponseWriter, r *http.Request) {
	common.Templates.ExecuteTemplate(w, "register.tmpl", nil)
}

// AutoRegisterHandler handles the path /register/auto
func AutoRegisterHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.CheckLogin(r.FormValue("username"), r.FormValue("password")) {
		common.Templates.ExecuteTemplate(w, "error.tmpl", struct{ Message string }{"Incorrect username or password"})
	}
}
