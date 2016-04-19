package api

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/onesimus-systems/packet-guardian/src/auth"
	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/dhcp"
	"github.com/onesimus-systems/packet-guardian/src/models"
)

type Device struct {
	e *common.Environment
}

func NewDeviceController(e *common.Environment) *Device {
	return &Device{e: e}
}

func (d *Device) RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	// Check authentication and get User models
	var formUser *models.User // The user to whom the device is being registered
	sessionUser := models.GetUserFromContext(r)
	if !auth.IsLoggedIn(r) {
		username := r.FormValue("username")
		// Authenticate
		if !auth.IsValidLogin(d.e.DB, username, r.FormValue("password")) {
			d.e.Log.Errorf("Failed authentication for %s", username)
			common.NewAPIResponse(common.APIStatusInvalidAuth, "Incorrect username or password", nil).WriteTo(w)
			return
		}

		// User authenticated successfully
		formUser, err := models.GetUserByUsername(d.e, username)
		if err != nil {
			d.e.Log.Errorf("Failed to get user from database %s: %s", username, err.Error())
			common.NewAPIResponse(common.APIStatusInvalidAuth, "Error registering device", nil).WriteTo(w)
			return
		}
		sessionUser = formUser
	} else {
		formUsername := r.FormValue("username")
		if formUsername == "" {
			common.NewAPIResponse(common.APIStatusInvalidAuth, "No username given", nil).WriteTo(w)
			return
		}

		if sessionUser.Username != formUsername && !sessionUser.IsAdmin() {
			d.e.Log.Errorf("Admin action attempted: Register device for %s attempted by user %s", formUsername, sessionUser.Username)
			common.NewAPIResponse(common.APIStatusInsufficientPrivilages, "Only admins can do that", nil).WriteTo(w)
			return
		}

		if formUsername == sessionUser.Username {
			formUser = sessionUser
		} else {
			var err error
			formUser, err = models.GetUserByUsername(d.e, formUsername)
			if err != nil {
				d.e.Log.Errorf("Failed to get user from database %s: %s", formUsername, err.Error())
				common.NewAPIResponse(common.APIStatusInvalidAuth, "Error registering device", nil).WriteTo(w)
				return
			}
		}
	}

	// Get and enforce the device limit
	var err error
	if !sessionUser.IsAdmin() { // Admins can register above the limit
		limit := d.e.Config.Registration.DefaultDeviceLimit
		if formUser.DeviceLimit != models.UserDeviceLimitGlobal {
			limit = int(formUser.DeviceLimit)
		}

		deviceCount, err := models.GetDeviceCountForUser(d.e, formUser)
		if err != nil {
			d.e.Log.Errorf("Error getting device count: %s", err.Error())
		}
		// A limit of 0 means unlimited
		if limit != 0 && deviceCount >= limit {
			common.NewAPIResponse(common.APIStatusGenericError, "Device limit reached", nil).WriteTo(w)
			return
		}
	}

	// Get MAC address
	var mac net.HardwareAddr
	macPost := r.FormValue("mac-address")
	manual := (macPost != "")
	ip := net.ParseIP(strings.Split(r.RemoteAddr, ":")[0])
	if manual {
		// Manual registration
		if !d.e.Config.Registration.AllowManualRegistrations {
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
	} else {
		// Automatic registration
		mac, err = dhcp.GetMacFromIP(ip)
		if err != nil {
			d.e.Log.Errorf("Failed to get MAC for IP %s: %s", ip, err.Error())
			common.NewAPIResponse(common.APIStatusGenericError, "Error detecting MAC address.", nil).WriteTo(w)
			return
		}
	}

	// Get device from database
	device, err := models.GetDeviceByMAC(d.e, mac)
	if err != nil {
		d.e.Log.Errorf("Error getting device: %s", err.Error())
	}

	// Check if device is already registered
	if device.ID != 0 {
		d.e.Log.Errorf("Attempted duplicate registration of MAC %s to user %s", mac.String(), formUser.Username)
		common.NewAPIResponse(common.APIStatusGenericError, "Device already registered", nil).WriteTo(w)
		return
	}

	// Check if the username is blacklisted
	// Administrators bypass the blacklist check
	if !sessionUser.IsAdmin() && formUser.IsBlacklisted() {
		d.e.Log.Errorf("Attempted registration by blacklisted user %s", formUser.Username)
		common.NewAPIResponse(common.APIStatusGenericError, "Failed to register device: Username blacklisted", nil).WriteTo(w)
		return
	}

	// Validate platform, we don't want someone to submit an inappropiate value
	platform := ""
	if manual && !common.StringInSlice(r.FormValue("platform"), d.e.Config.Registration.ManualRegPlatforms) {
		platform = r.FormValue("platform")
	}

	// Fill in device information
	device.Username = formUser.Username
	device.RegisteredFrom = ip
	device.Platform = platform
	device.Expires = time.Unix(0, 0)
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
		resp.Location = "/admin/manage/" + formUser.Username
	}

	common.NewAPIOK("Registration successful", resp).WriteTo(w)
}

func (d *Device) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	sessionUser := models.GetUserFromContext(r)
	formUser := sessionUser
	if r.FormValue("username") != sessionUser.Username {
		if !sessionUser.IsAdmin() {
			common.NewAPIResponse(common.APIStatusGenericError, "Admin Error", nil).WriteTo(w)
			return
		}
		var err error
		formUser, err = models.GetUserByUsername(d.e, r.FormValue("username"))
		if err != nil {
			d.e.Log.Errorf("Error getting user: %s", err.Error())
			common.NewAPIResponse(common.APIStatusGenericError, "Server error", nil).WriteTo(w)
			return
		}
	}

	deleteAll := (r.FormValue("mac") == "")
	macsToDelete := strings.Split(r.FormValue("mac"), ",")
	usersDevices, err := models.GetDevicesForUser(d.e, formUser)
	if err != nil {
		d.e.Log.Errorf("Error deleting devices: %s", err.Error())
		common.NewAPIResponse(common.APIStatusGenericError, "Error deleting devices", nil).WriteTo(w)
		return
	}

	finishedWithErrors := false
	for _, device := range usersDevices {
		if !deleteAll && !common.StringInSlice(device.MAC.String(), macsToDelete) {
			continue
		}

		if err := device.Delete(); err != nil {
			d.e.Log.Errorf("Error deleting device %s: %s", device.MAC.String(), err.Error())
			finishedWithErrors = true
			continue
		}

		d.e.Log.Infof("Deleted device %s for user %s", device.MAC.String(), formUser.Username)
	}

	if finishedWithErrors {
		common.NewAPIResponse(common.APIStatusGenericError, "Finished but with errors", nil).WriteTo(w)
		return
	}

	common.NewAPIOK("Devices deleted successful", nil).WriteTo(w)
}
