package dhcp

import (
	"bufio"
	"html/template"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/onesimus-systems/packet-guardian/auth"
	"github.com/onesimus-systems/packet-guardian/common"
)

const (
	nonAdminAutoReg       = "na-auto"
	nonAdminManReg        = "na-man"
	nonAdminManRegNologin = "na-man-nologin"
	adminReg              = "admin"
)

// RegistrationPageHandler handles GET requests to /register
func RegistrationPageHandler(e *common.Environment, w http.ResponseWriter, r *http.Request) {
	flash := ""
	man := (r.FormValue("manual") == "1")
	loggedIn := auth.IsLoggedIn(e, r)
	ip := net.ParseIP(strings.Split(r.RemoteAddr, ":")[0])
	reg, _ := IsRegisteredByIP(e.DB, ip)
	if !man && reg {
		http.Redirect(w, r, "/manage", http.StatusTemporaryRedirect)
		return
	}

	formType := nonAdminAutoReg
	if man {
		if !e.Config.Core.AllowManualRegistrations {
			flash = "Manual registrations are not allowed"
		} else {
			formType = nonAdminManReg
			if loggedIn {
				formType = nonAdminManRegNologin
			}
		}
	}

	username := e.Sessions.GetSession(r, e.Config.Webserver.SessionName).GetString("username")
	if r.FormValue("username") != "" && auth.IsAdminUser(e, r) {
		username = r.FormValue("username")
		formType = "admin"
		flash = ""
	}

	// TODO: Don't show the registration page at all if blacklisted
	if formType == nonAdminAutoReg {
		ip := net.ParseIP(strings.Split(r.RemoteAddr, ":")[0])
		mac, err := GetMacFromIP(ip)
		if err != nil {
			e.Log.Errorf("Failed to get MAC from IP for %s", ip.String())
		} else {
			bl, err := IsBlacklisted(e.DB, mac.String())
			if err != nil {
				e.Log.Errorf("There was an error checking the blacklist for MAC %s", mac.String())
			}
			if bl {
				flash = "The device appears to be blacklisted"
			}
		}
	}

	data := struct {
		SiteTitle    string
		CompanyName  string
		Policy       []template.HTML
		Type         string
		Username     string
		FlashMessage string
	}{
		SiteTitle:    e.Config.Core.SiteTitle,
		CompanyName:  e.Config.Core.SiteCompanyName,
		Policy:       loadPolicyText(e.Config.Core.RegistrationPolicyFile),
		Type:         formType,
		Username:     username,
		FlashMessage: flash,
	}

	if err := e.Templates.ExecuteTemplate(w, "register", data); err != nil {
		e.Log.Error(err.Error())
	}
}

// RegistrationHandler handles POST requests to /register
func RegistrationHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			RegistrationPageHandler(e, w, r)
			return
		}

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
				common.NewAPIResponse(common.APIStatusNotAdmin, "Only admins can do that", nil).WriteTo(w)
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
		// Administrators bypass the blacklist check
		if !auth.IsAdminUser(e, r) {
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
		}

		// Register the MAC to the user
		err = Register(e.DB, mac, username, r.FormValue("platform"), ip, r.UserAgent(), "")
		if err != nil {
			e.Log.Errorf("Failed to register MAC address %s to user %s: %s", mac.String(), username, err.Error())
			// TODO: Look at what error messages are being returned from Register()
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
			common.NewAPIResponse(common.APIStatusNotAdmin, "Only admins can do that", nil).WriteTo(w)
			return
		}

		var devices []Device
		var sqlParams []interface{}
		var sql string
		var err error

		deviceIDs := strings.Split(r.FormValue("devices"), ",")
		all := (len(deviceIDs) == 1 && deviceIDs[0] == "")

		if all {
			sql = "DELETE FROM \"device\" WHERE \"username\" = ?"
		} else {
			sql = "DELETE FROM \"device\" WHERE (0 = 1"
			dIDs := make([]int, len(deviceIDs))

			for i, deviceID := range deviceIDs {
				sql += " OR \"id\" = ?"
				sqlParams = append(sqlParams, deviceID)
				if in, err := strconv.Atoi(deviceID); err == nil {
					dIDs[i] = in
				} else {
					e.Log.Error(err.Error())
					common.NewAPIResponse(common.APIStatusGenericError, "Error deleting devices", nil).WriteTo(w)
					return
				}
			}
			sql += ") AND \"username\" = ?"
			devices = Query{ID: dIDs}.Search(e)
		}
		sqlParams = append(sqlParams, formUsername)

		if bl, _ := IsBlacklisted(e.DB, sqlParams...); bl && !auth.IsAdminUser(e, r) {
			e.Log.Errorf("Admin action attempted: Delete blacklisted device by user %s", formUsername)
			common.NewAPIResponse(common.APIStatusNotAdmin, "Only admins can do that", nil).WriteTo(w)
			return
		}

		_, err = e.DB.Exec(sql, sqlParams...)
		if err != nil {
			e.Log.Error(err.Error())
			common.NewAPIResponse(common.APIStatusGenericError, "Error deleting devices", nil).WriteTo(w)
			return
		}

		if all {
			e.Log.Infof("Successfully deleted all registrations for user %s by user %s", formUsername, sessUsername)
		} else {
			if devices != nil {
				for i := range devices {
					e.Log.Infof("Successfully deleted MAC %s for user %s by user %s", devices[i].MAC, formUsername, sessUsername)
				}
			} else {
				e.Log.Error("An error occured that prevents me from listing the deleted MAC addresses")
			}
		}
		common.NewAPIResponse(common.APIStatusOK, "Devices deleted successfully", nil).WriteTo(w)
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
