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
		reg, _ := IsRegisteredByIP(e.DB, ip)
		if !man && reg {
			http.Redirect(w, r, "/manage", http.StatusTemporaryRedirect)
			return
		}

		formType := "na-auto"
		if man {
			formType = "na-man"
			if loggedIn {
				formType = "na-man-nologin"
			}
		}

		username := e.Sessions.GetSession(r, e.Config.Webserver.SessionName).GetString("username")
		if r.FormValue("username") != "" && auth.IsAdminUser(e, r) {
			username = r.FormValue("username")
			formType = "admin"
		}

		data := struct {
			SiteTitle   string
			CompanyName string
			Policy      []template.HTML
			Type        string
			Username    string
		}{
			SiteTitle:   e.Config.Core.SiteTitle,
			CompanyName: e.Config.Core.SiteCompanyName,
			Policy:      loadPolicyText(e.Config.Core.RegistrationPolicyFile),
			Type:        formType,
			Username:    username,
		}
		e.Templates.ExecuteTemplate(w, "register", data)
	}
}

// RegistrationHandler handles POST requests to /register
func RegistrationHandler(e *common.Environment) http.HandlerFunc {
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
			formUsername := r.FormValue("username")
			sessUsername := e.Sessions.GetSession(r, e.Config.Webserver.SessionName).GetString("username")
			if formUsername == "" || sessUsername == "" {
				common.NewAPIResponse(common.APIStatusInvalidAuth, "Incorrect username or password", nil).WriteTo(w)
				return
			}

			if sessUsername != formUsername && !auth.IsAdminUser(e, r) {
				e.Log.Errorf("Admin action attempted: Register device for %s attempted by user %s", formUsername, sessUsername)
				w.Write(common.NewAPIResponse(common.APIStatusInvalidAuth, "Only admins can do that", nil).Encode())
				return
			}
			username = formUsername
		}

		// Get MAC address
		var err error
		var mac net.HardwareAddr
		macPost := r.FormValue("mac-address")
		ip := net.ParseIP(strings.Split(r.RemoteAddr, ":")[0])
		if macPost == "" {
			// Automatic registration
			mac, err = GetMacFromIP(ip)
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
			mac, err = FormatMacAddress(macPost)
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
		e.Log.Infof("Successfully registered MAC %s to user %s", mac.String(), username)

		resp := struct{ Location string }{Location: "/manage"}
		if auth.IsAdminUser(e, r) {
			resp.Location = "/admin/user/" + username
		}

		common.NewAPIOK("Registration successful", resp).WriteTo(w)
	}
}

// DeleteHandler handles the path /devices/delete
func DeleteHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessUsername := e.Sessions.GetSession(r, e.Config.Webserver.SessionName).GetString("username")
		formUsername := r.FormValue("username")

		// If the two don't match, check if the user is an admin, if not return an error
		if sessUsername != formUsername && !auth.IsAdminUser(e, r) {
			e.Log.Errorf("Admin action attempted: Delete device for %s attempted by user %s", formUsername, sessUsername)
			w.Write(common.NewAPIResponse(common.APIStatusInvalidAuth, "Only admins can do that", nil).Encode())
			return
		}

		devices := strings.Split(r.FormValue("devices"), ",")
		all := (len(devices) == 1 && devices[0] == "")
		sql := ""
		var sqlParams []interface{}
		if all {
			sql = "DELETE FROM \"device\" WHERE \"username\" = ?"
		} else {
			sql = "DELETE FROM \"device\" WHERE (0 = 1"

			for _, device := range devices {
				if mac, err := FormatMacAddress(device); err == nil {
					sql += " OR \"mac\" = ?"
					sqlParams = append(sqlParams, mac.String())
				}
			}
			sql += ") AND \"username\" = ?"
		}
		sqlParams = append(sqlParams, formUsername)
		_, err := e.DB.Exec(sql, sqlParams...)
		if err != nil {
			e.Log.Error(err.Error())
			w.Write(common.NewAPIResponse(common.APIStatusGenericError, "Error deleting devices", nil).Encode())
			return
		}

		if all {
			e.Log.Infof("Successfully deleted all registrations for user %s by user %s", formUsername, sessUsername)
		} else {
			for _, mac := range devices {
				e.Log.Infof("Successfully deleted MAC %s for user %s by user %s", mac, formUsername, sessUsername)
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
