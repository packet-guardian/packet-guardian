package dhcp

import (
	"bufio"
	"html/template"
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
		man := (r.FormValue("manual") == "1")
		loggedIn := auth.IsLoggedIn(e, r)
		ip := net.ParseIP(strings.Split(r.RemoteAddr, ":")[0])
		reg, _ := IsRegisteredByIP(e.DB, ip, e.Config.DHCP.LeasesFile)
		if !man && reg {
			http.Redirect(w, r, "/manage", http.StatusTemporaryRedirect)
			return
		}
		if !reg && !man {
			loggedIn = false
		}

		data := struct {
			SiteTitle     string
			CompanyName   string
			Policy        []template.HTML
			Manual        bool
			ShowLoginForm bool
		}{
			SiteTitle:     e.Config.Core.SiteTitle,
			CompanyName:   e.Config.Core.SiteCompanyName,
			Policy:        loadPolicyText(e.Config.Core.RegistrationPolicyFile),
			Manual:        man,
			ShowLoginForm: !loggedIn,
		}
		e.Templates.ExecuteTemplate(w, "register", data)
	}
}

// AutoRegisterHandler handles POST requests to /register
func AutoRegisterHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check authentication
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
		var mac net.HardwareAddr
		macPost := r.FormValue("mac-address")
		ip := net.ParseIP(strings.Split(r.RemoteAddr, ":")[0])
		if macPost == "" {
			// Automatic registration
			mac, err = GetMacFromIP(ip, e.Config.DHCP.LeasesFile)
			if err != nil {
				e.Log.Errorf("Failed to get MAC for IP %s: %s", ip, err.Error())
				common.NewAPIResponse(common.APIStatusGenericError, "Error detecting MAC address.", nil).WriteTo(w)
				return
			}
		} else {
			// Manual registration
			if !e.Config.Core.AllowManualRegistrations {
				e.Log.Errorf("Unauthorized manual registration attempt for MAC %s from user %s", macPost, username)
				common.NewAPIResponse(common.APIStatusGenericError, "Manual registrations are not allowed.", nil).WriteTo(w)
				return
			}
			mac, err = formatMacAddress(macPost)
			if err != nil {
				e.Log.Errorf("Error formatting MAC %s", macPost)
				common.NewAPIResponse(common.APIStatusGenericError, "Incorrect MAC address format.", nil).WriteTo(w)
				return
			}
		}

		// Check if the mac or username is blacklisted
		bl, err := IsBlacklisted(e.DB, mac.String(), username)
		if err != nil {
			e.Log.Errorf("There was an error checking the blacklist for MAC %s and user %s", mac.String(), username)
			common.NewAPIResponse(common.APIStatusGenericError, "There was an error registering your device.", nil).WriteTo(w)
			return
		}
		if bl {
			e.Log.Errorf("Attempted registration of blacklisted MAC or user %s - %s", mac.String(), username)
			common.NewAPIResponse(common.APIStatusGenericError, "There was an error registering your device. Blacklisted username or MAC address", nil).WriteTo(w)
			return
		}

		// Register the MAC to the user
		err = Register(e.DB, mac, username, r.FormValue("platform"), ip, r.UserAgent(), "")
		if err != nil {
			e.Log.Errorf("Failed to register MAC address %s to user %s: %s", mac.String(), username, err.Error())
			common.NewAPIResponse(common.APIStatusGenericError, "Error: "+err.Error(), nil).WriteTo(w)
			return
		}
		e.DHCP.WriteHostFile()
		e.Log.Infof("Successfully registered MAC %s to user %s", mac.String(), username)
		common.NewAPIOK("Registration successful", nil).WriteTo(w)
	}
}

// DeleteHandler handles the path /devices/delete
func DeleteHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := e.Sessions.GetSession(r, e.Config.Webserver.SessionName).GetString("username")
		devices := strings.Split(r.FormValue("devices"), ",")
		all := (len(devices) == 1 && devices[0] == "")
		sql := ""
		sqlParams := make([]interface{}, 0)
		if all {
			sql = "DELETE FROM \"device\" WHERE \"username\" = ?"
		} else {
			sql = "DELETE FROM \"device\" WHERE (0 = 1"

			for _, device := range devices {
				if mac, err := formatMacAddress(device); err == nil {
					sql += " OR \"mac\" = ?"
					sqlParams = append(sqlParams, mac.String())
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
		if all {
			e.Log.Infof("Successfully deleted all registrations for user %s", username)
		} else {
			for _, mac := range devices {
				e.Log.Infof("Successfully deleted MAC %s for user %s", mac, username)
			}
		}
		w.Write(common.NewAPIResponse(common.APIStatusOK, "Devices deleted successfully", nil).Encode())
	}
}

func loadPolicyText(file string) []template.HTML {
	f, err := os.Open(file)
	if err != nil {
		return nil
	}
	defer f.Close()

	policy := make([]template.HTML, 0)
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
