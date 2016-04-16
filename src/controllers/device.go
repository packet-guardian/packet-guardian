package controllers

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/context"
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
	// Check authentication and get User models
	var formUser *model.User // The user to whom the device is being registered
	sessionUser := context.Get(r, "sessionUser").(*models.User)
	if !auth.IsLoggedIn(d.e, r) {
		username = r.FormValue("username")
		// Authenticate
		if !auth.IsValidLogin(d.e.DB, username, r.FormValue("password")) {
			d.e.Log.Errorf("Failed authentication for %s", username)
			common.NewAPIResponse(common.APIStatusInvalidAuth, "Incorrect username or password", nil).WriteTo(w)
			return
		}

		formUser = sessionUser
	} else {
		formUsername := r.FormValue("username")
		if formUsername == "" {
			common.NewAPIResponse(common.APIStatusInvalidAuth, "Incorrect username or password", nil).WriteTo(w)
			return
		}

		if sessionUser.Username != formUsername && !sessionUser.IsAdmin() {
			d.e.Log.Errorf("Admin action attempted: Register device for %s attempted by user %s", formUsername, sessUsername)
			common.NewAPIResponse(common.APIStatusNotAdmin, "Only admins can do that", nil).WriteTo(w)
			return
		}

		if formUsername == sessUsername {
			formUser = sessionUser
		} else {
			formUser, err := models.GetUserByUsername(formUsername)
			if err != nil {
				d.e.Log.Error(err.Error())
				// TODO: Show error page to user
				return
			}
		}
	}

	// Get and enforce the device limit
	var err error
	limit := d.e.Config.Core.DefaultDeviceLimit
	if formUser.DeviceLimit != -1 {
		limit = formUser.DeviceLimit
	}

	deviceCount, err := models.GetDeviceCountForUser(e, u)
	if err != nil {
		d.e.Log.Errorf("Error getting device count: %s", err.Error())
	}
	// A limit of 0 means unlimited
	if limit != 0 && deviceCount >= limit {
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
			d.e.Log.Errorf("Unauthorized manual registration attempt for MAC %s from user %s", macPost, formUser.Username)
			common.NewAPIResponse(common.APIStatusGenericError, "Manual registrations are not allowed.", nil).WriteTo(w)
			return
		}
		mac, err = common.FormatMacAddress(macPost)
		if err != nil {
			d.e.Log.Errorf("Error formatting MAC %s", macPost)
			common.NewAPIResponse(common.APIStatusGenericError, "Incorrect MAC address format.", nil).WriteTo(w)
			return
		}
	}

	// Get device from database
	device, err := models.GetDeviceByMAC(e, mac)
	if err != nil {
		d.e.Log.Errorf("Error getting device: %s", err.Error())
	}

	// Check if device is already registered
	if device.ID != 0 {
		d.e.Log.Errorf("Attempted duplicate registration of MAC %s to user %s", mac.String(), formUser.Username)
		common.NewAPIResponse(common.APIStatusGenericError, "Device already registered", nil).WriteTo(w)
		return
	}

	// Check if the mac or username is blacklisted
	// Administrators bypass the blacklist check
	if !sessionUser.IsAdmin() {
		if device.Blacklisted {
			d.e.Log.Errorf("Attempted registration of blacklisted MAC %s by user %s", mac.String(), formUser.Username)
			common.NewAPIResponse(common.APIStatusGenericError, "Failed to register device: MAC address blacklisted", nil).WriteTo(w)
			return
		}
		if formUser.IsBlacklisted() {
			d.e.Log.Errorf("Attempted registration by blacklisted user %s", formUser.Username)
			common.NewAPIResponse(common.APIStatusGenericError, "Failed to register device: Username blacklisted", nil).WriteTo(w)
			return
		}
	}

	// Fill in device information
	device.Username = formUser.Username
	device.RegisteredFrom = ip
	device.Platform = r.FormValue("platform")
	device.Expired = time.Unix(0, 0)
	device.DateRegistered = time.Now()
	device.UserAgent = r.UserAgent()

	// Save new device
	if err := device.Save(); err != nil {
		d.e.Log.Errorf("Error registering device: %s", err.Error())
		common.NewAPIResponse(common.APIStatusGenericError, "Error registering device", nil).WriteTo(w)
	}
	d.e.Log.Infof("Successfully registered MAC %s to user %s", mac.String(), formUser.Username)

	// Redirect client as needed
	resp := struct{ Location string }{Location: "/manage"}
	if sessionUser.IsAdmin() {
		resp.Location = "/admin/user/" + formUser.Username
	}

	common.NewAPIOK("Registration successful", resp).WriteTo(w)
}

func (d *Device) deleteHandler(w http.ResponseWriter, r *http.Request) {

}
