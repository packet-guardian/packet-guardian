package dhcp

import (
	"net"
	"net/http"
	"strings"

	"github.com/onesimus-systems/packet-guardian/auth"
	"github.com/onesimus-systems/packet-guardian/common"
)

// RegisterHTTPHandler serves and handles the registration page for an end user
func RegisterHTTPHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		e.Templates.ExecuteTemplate(w, "register.tmpl", nil)
	}
}

// AutoRegisterHandler handles the path /register/auto
func AutoRegisterHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("username")
		// Authenticate
		if !auth.IsValidLogin(e.DB, username, r.FormValue("password")) {
			e.Log.Errorf("Failed authentication for %s", username)
			e.Templates.ExecuteTemplate(w, "error.tmpl", struct{ Message string }{"Incorrect username or password"})
			return
		}

		// Get MAC address
		ip := strings.Split(r.RemoteAddr, ":")[0]
		mac, err := GetMacFromIP(net.ParseIP(ip), e.Config.DHCP.LeasesFile)
		if err != nil {
			e.Log.Errorf("Failed to get MAC for IP %s: %s", ip, err.Error())
			e.Templates.ExecuteTemplate(w, "error.tmpl", struct{ Message string }{"Error detected MAC address"})
			return
		}

		// Check if it or the username is blacklisted
		bl, err := IsBlacklisted(e.DB, mac, username)
		if err != nil {
			e.Log.Errorf("There was an error checking the blacklist for MAC %s and user %s", mac, username)
			e.Templates.ExecuteTemplate(w, "error.tmpl", struct{ Message string }{"There was an error registering your device."})
		}
		if bl {
			e.Log.Errorf("Attempted authentication of blacklisted MAC or user %s - %s", mac, username)
			e.Templates.ExecuteTemplate(w, "error.tmpl", struct{ Message string }{"There was an error registering your device. BL"})
			return
		}

		// Register the MAC to the user
		err = Register(e.DB, mac, username, r.FormValue("platform"), ip, r.UserAgent(), "")
		if err != nil {
			e.Log.Errorf("Failed to register MAC address %s to user %s: %s", mac, username, err.Error())
			e.Templates.ExecuteTemplate(w, "error.tmpl", struct{ Message string }{"There was an error registering your device"})
			return
		}
		e.Log.Infof("Successfully registered MAC %s to user %s", mac, username)
		e.Templates.ExecuteTemplate(w, "success.tmpl", struct{ Message string }{"Device successfully registered"})
	}
}
