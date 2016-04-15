package controllers

import (
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/onesimus-systems/packet-guardian/src/auth"
	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/dhcp"
	"github.com/onesimus-systems/packet-guardian/src/models"
	"github.com/onesimus-systems/packet-guardian/src/server/middleware"
)

type Device struct {
	e *common.Environment
}

func NewDeviceController(e *common.Environment) *Device {
	return &Device{e: e}
}

func (d *Device) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/device/register", d.registrationHandler).Methods("POST")
	r.HandleFunc("/devices/delete", middleware.CheckAuthAPI(d.e, d.deleteHandler)).Methods("POST")
}

func (d *Device) registrationHandler(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	username := ""
	if !auth.IsLoggedIn(d.e, r) {
		username = r.FormValue("username")
		// Authenticate
		if !auth.IsValidLogin(d.e.DB, username, r.FormValue("password")) {
			d.e.Log.Errorf("Failed authentication for %s", username)
			common.NewAPIResponse(common.APIStatusInvalidAuth, "Incorrect username or password", nil).WriteTo(w)
			return
		}
	} else {
		formUsername := r.FormValue("username")
		sessUsername := d.e.Sessions.GetSession(r).GetString("username")
		if formUsername == "" || sessUsername == "" {
			common.NewAPIResponse(common.APIStatusInvalidAuth, "Incorrect username or password", nil).WriteTo(w)
			return
		}

		if sessUsername != formUsername && !auth.IsAdminUser(d.e, r) {
			d.e.Log.Errorf("Admin action attempted: Register device for %s attempted by user %s", formUsername, sessUsername)
			common.NewAPIResponse(common.APIStatusNotAdmin, "Only admins can do that", nil).WriteTo(w)
			return
		}
		username = formUsername
	}

	var err error
	limit := d.e.Config.Core.DefaultDeviceLimit
	if user, _ := models.GetUser(d.e.DB, username); user != nil && user.DeviceLimit != -1 {
		limit = user.DeviceLimit
	}

	devices := dhcp.Query{User: username}.Search(d.e)
	if limit != 0 && len(devices) >= limit {
		common.NewAPIResponse(common.APIStatusGenericError, "Device limit reached", nil).WriteTo(w)
		return
	}

	// Get MAC address
	var mac net.HardwareAddr
	macPost := r.FormValue("mac-address")
	ip := net.ParseIP(strings.Split(r.RemoteAddr, ":")[0])
	if macPost == "" {
		// Automatic registration
		mac, err = dhcp.GetMacFromIP(ip)
		if err != nil {
			d.e.Log.Errorf("Failed to get MAC for IP %s: %s", ip, err.Error())
			common.NewAPIResponse(common.APIStatusGenericError, "Error detecting MAC address.", nil).WriteTo(w)
			return
		}
	} else {
		// Manual registration
		if !d.e.Config.Core.AllowManualRegistrations {
			d.e.Log.Errorf("Unauthorized manual registration attempt for MAC %s from user %s", macPost, username)
			common.NewAPIResponse(common.APIStatusGenericError, "Manual registrations are not allowed.", nil).WriteTo(w)
			return
		}
		mac, err = dhcp.FormatMacAddress(macPost)
		if err != nil {
			d.e.Log.Errorf("Error formatting MAC %s", macPost)
			common.NewAPIResponse(common.APIStatusGenericError, "Incorrect MAC address format.", nil).WriteTo(w)
			return
		}
	}

	// Check if the mac or username is blacklisted
	// Administrators bypass the blacklist check
	if !auth.IsAdminUser(d.e, r) {
		bl, err1 := dhcp.IsBlacklisted(d.e.DB, mac.String(), username)
		if err1 != nil {
			d.e.Log.Errorf("There was an error checking the blacklist for MAC %s and user %s: %s", mac.String(), username, err1.Error())
			common.NewAPIResponse(common.APIStatusGenericError, "There was an error registering your devicd.e.", nil).WriteTo(w)
			return
		}
		if bl {
			d.e.Log.Errorf("Attempted registration of blacklisted MAC or user %s - %s", mac.String(), username)
			common.NewAPIResponse(common.APIStatusGenericError, "There was an error registering your devicd.e. Blacklisted username or MAC address", nil).WriteTo(w)
			return
		}
	}

	// Register the MAC to the user
	err = dhcp.Register(d.e.DB, mac, username, r.FormValue("platform"), ip, r.UserAgent(), "")
	if err != nil {
		d.e.Log.Errorf("Failed to register MAC address %s to user %s: %s", mac.String(), username, err.Error())
		// TODO: Look at what error messages are being returned from Register()
		common.NewAPIResponse(common.APIStatusGenericError, "Error: "+err.Error(), nil).WriteTo(w)
		return
	}
	d.e.Log.Infof("Successfully registered MAC %s to user %s", mac.String(), username)

	resp := struct{ Location string }{Location: "/manage"}
	if auth.IsAdminUser(d.e, r) {
		resp.Location = "/admin/user/" + username
	}

	common.NewAPIOK("Registration successful", resp).WriteTo(w)
}

func (d *Device) deleteHandler(w http.ResponseWriter, r *http.Request) {
	sessUsername := d.e.Sessions.GetSession(r).GetString("username")
	formUsername := r.FormValue("username")

	// If the two don't match, check if the user is an admin, if not return an error
	if sessUsername != formUsername && !auth.IsAdminUser(d.e, r) {
		d.e.Log.Errorf("Admin action attempted: Delete device for %s attempted by user %s", formUsername, sessUsername)
		common.NewAPIResponse(common.APIStatusNotAdmin, "Only admins can do that", nil).WriteTo(w)
		return
	}

	var devices []dhcp.Device
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
				d.e.Log.Error(err.Error())
				common.NewAPIResponse(common.APIStatusGenericError, "Error deleting devices", nil).WriteTo(w)
				return
			}
		}
		sql += ") AND \"username\" = ?"
		devices = dhcp.Query{ID: dIDs}.Search(d.e)
	}
	sqlParams = append(sqlParams, formUsername)

	if bl, _ := dhcp.IsBlacklisted(d.e.DB, sqlParams...); bl && !auth.IsAdminUser(d.e, r) {
		d.e.Log.Errorf("Admin action attempted: Delete blacklisted device by user %s", formUsername)
		common.NewAPIResponse(common.APIStatusNotAdmin, "Only admins can do that", nil).WriteTo(w)
		return
	}

	_, err = d.e.DB.Exec(sql, sqlParams...)
	if err != nil {
		d.e.Log.Error(err.Error())
		common.NewAPIResponse(common.APIStatusGenericError, "Error deleting devices", nil).WriteTo(w)
		return
	}

	if all {
		d.e.Log.Infof("Successfully deleted all registrations for user %s by user %s", formUsername, sessUsername)
	} else {
		if devices != nil {
			for i := range devices {
				d.e.Log.Infof("Successfully deleted MAC %s for user %s by user %s", devices[i].MAC, formUsername, sessUsername)
			}
		} else {
			d.e.Log.Error("An error occured that prevents me from listing the deleted MAC addresses")
		}
	}
	common.NewAPIResponse(common.APIStatusOK, "Devices deleted successfully", nil).WriteTo(w)
}
