package dhcp

import (
	"bufio"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/onesimus-systems/packet-guardian/auth"
	"github.com/onesimus-systems/packet-guardian/common"
)

// RegistrationPageHandler handles GET requests to /register
func RegistrationPageHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ip := net.ParseIP(strings.Split(r.RemoteAddr, ":")[0])
		// if reg, _ := IsRegisteredByIP(e.DB, ip, e.Config.DHCP.LeasesFile); reg {
		// 	http.Redirect(w, r, "/manage", http.StatusTemporaryRedirect)
		// 	return
		// }

		data := struct {
			SiteTitle   string
			CompanyName string
			Policy      []string
		}{
			SiteTitle:   e.Config.Core.SiteTitle,
			CompanyName: e.Config.Core.SiteCompanyName,
			Policy:      loadPolicyText(e.Config.Core.RegistrationPolicyFile),
		}
		e.Templates.ExecuteTemplate(w, "register", data)
	}
}

// AutoRegisterHandler handles POST requests to /register
func AutoRegisterHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := ""
		if !auth.IsLoggedIn(e, r) {
			username = r.FormValue("username")
			// Authenticate
			if !auth.IsValidLogin(e.DB, username, r.FormValue("password")) {
				e.Log.Errorf("Failed authentication for %s", username)
				common.NewAPIResponse(common.APIStatusInvalidAuth, "Incorrect username or password", nil).WriteTo(w)
				return
			}
		} else {
			username = e.Sessions.GetSession(r, e.Config.Webserver.SessionName).GetString("username")

			if username == "" {
				common.NewAPIResponse(common.APIStatusInvalidAuth, "Incorrect username or password", nil).WriteTo(w)
			}
		}

		// Get MAC address
		var err error
		mac := r.FormValue("mac-address")
		ip := ""
		if mac == "" {
			// Automatic registration
			ip = strings.Split(r.RemoteAddr, ":")[0]
			mac, err = GetMacFromIP(net.ParseIP(ip), e.Config.DHCP.LeasesFile)
			if err != nil {
				e.Log.Errorf("Failed to get MAC for IP %s: %s", ip, err.Error())
				common.NewAPIResponse(common.APIStatusGenericError, "Error detected MAC address.", nil).WriteTo(w)
				return
			}
		} else {
			// Manual registration
			if !e.Config.Core.AllowManualRegistrations {
				e.Log.Errorf("Unauthorized manual registration attempt for MAC %s from user %s", mac, username)
				common.NewAPIResponse(common.APIStatusGenericError, "Manual registrations are not allowed.", nil).WriteTo(w)
				return
			}
		}

		// Check if the mac or username is blacklisted
		bl, err := IsBlacklisted(e.DB, mac, username)
		if err != nil {
			e.Log.Errorf("There was an error checking the blacklist for MAC %s and user %s", mac, username)
			common.NewAPIResponse(common.APIStatusGenericError, "There was an error registering your device.", nil).WriteTo(w)
			return
		}
		if bl {
			e.Log.Errorf("Attempted authentication of blacklisted MAC or user %s - %s", mac, username)
			common.NewAPIResponse(common.APIStatusGenericError, "There was an error registering your device. Blacklisted username or MAC address", nil).WriteTo(w)
			return
		}

		// Register the MAC to the user
		err = Register(e.DB, mac, username, r.FormValue("platform"), ip, r.UserAgent(), "")
		if err != nil {
			e.Log.Errorf("Failed to register MAC address %s to user %s: %s", mac, username, err.Error())
			common.NewAPIResponse(common.APIStatusGenericError, "Error: "+err.Error(), nil).WriteTo(w)
			return
		}
		e.DHCP.WriteHostFile()
		e.Log.Infof("Successfully registered MAC %s to user %s", mac, username)
		common.NewAPIOK("Registration successful", nil).WriteTo(w)
	}
}

// DeleteHandler handles the path /devices/delete
func DeleteHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := e.Sessions.GetSession(r, e.Config.Webserver.SessionName)
		username := sess.GetString("username")

		devicesStr := r.FormValue("devices")
		devices := strings.Split(devicesStr, ",")
		sql := ""
		sqlParams := make([]interface{}, 0)
		if devicesStr == "" {
			sql = "DELETE FROM \"device\" WHERE \"username\" = ?"
		} else {
			sql = "DELETE FROM \"device\" WHERE (0 = 1"

			for _, device := range devices {
				if isValidMac(device) {
					sql += " OR \"mac\" = ?"
					sqlParams = append(sqlParams, device)
				}
			}
			sql += ") AND \"username\" = ?"
		}
		sqlParams = append(sqlParams, username)
		_, err := e.DB.Exec(sql, sqlParams...)
		if err != nil {
			e.Log.Error(err.Error())
			w.Write(common.NewAPIResponse(common.APIStatusGenericError, "SQL statement failed", nil).Encode())
			return
		}

		e.DHCP.WriteHostFile()
		if devicesStr == "" {
			e.Log.Infof("Successfully deleted all registrations for user %s", username)
		} else {
			for _, mac := range devices {
				e.Log.Infof("Successfully deleted MAC %s for user %s", mac, username)
			}
		}
		w.Write(common.NewAPIResponse(common.APIStatusOK, "Devices deleted successfully", nil).Encode())
	}
}

// RegisterHandler handles the path /devices/register
func RegisterHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		return
	}
}

func loadPolicyText(file string) []string {
	f, err := os.Open(file)
	if err != nil {
		return nil
	}
	defer f.Close()

	policy := make([]string, 0)
	currentParagraph := ""
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		t := strings.TrimSpace(scanner.Text())
		if t == "" {
			policy = append(policy, currentParagraph)
			currentParagraph = ""
			continue
		}
		currentParagraph += " " + t
	}
	policy = append(policy, currentParagraph)
	return policy
}
