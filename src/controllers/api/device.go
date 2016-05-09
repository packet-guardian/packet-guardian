package api

import (
	"net"
	"net/http"
	"strings"
	"time"

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
	var err error
	sessionUser := models.GetUserFromContext(r)
	formUsername := r.FormValue("username")
	if formUsername == "" {
		common.NewAPIResponse("No username given", nil).WriteResponse(w, http.StatusBadRequest)
		return
	}

	if sessionUser.Username != formUsername && !sessionUser.IsAdmin() {
		d.e.Log.Errorf("Admin action attempted: Register device for %s attempted by user %s", formUsername, sessionUser.Username)
		common.NewAPIResponse("Only admins can do that", nil).WriteResponse(w, http.StatusForbidden)
		return
	}

	if formUsername == sessionUser.Username {
		formUser = sessionUser
	} else {
		var err error
		formUser, err = models.GetUserByUsername(d.e, formUsername)
		if err != nil {
			d.e.Log.Errorf("Failed to get user from database %s: %s", formUsername, err.Error())
			common.NewAPIResponse("Error registering device", nil).WriteResponse(w, http.StatusInternalServerError)
			return
		}
	}

	if !sessionUser.IsAdmin() { // Admins can register above the limit
		// Get and enforce the device limit
		limit := models.UserDeviceLimit(d.e.Config.Registration.DefaultDeviceLimit)
		if formUser.DeviceLimit != models.UserDeviceLimitGlobal {
			limit = formUser.DeviceLimit
		}

		deviceCount, err := models.GetDeviceCountForUser(d.e, formUser)
		if err != nil {
			d.e.Log.Errorf("Error getting device count: %s", err.Error())
		}
		if limit != models.UserDeviceLimitUnlimited && deviceCount >= int(limit) {
			common.NewAPIResponse("Device limit reached", nil).WriteResponse(w, http.StatusConflict)
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
			d.e.Log.Noticef("Unauthorized manual registration attempt for MAC %s from user %s", macPost, formUser.Username)
			common.NewAPIResponse("Manual registrations are not allowed", nil).WriteResponse(w, http.StatusForbidden)
			return
		}
		mac, err = common.FormatMacAddress(macPost)
		if err != nil {
			common.NewAPIResponse("Incorrect MAC address format", nil).WriteResponse(w, http.StatusBadRequest)
			return
		}
	} else {
		// Automatic registration
		lease, err := dhcp.GetLeaseByIP(d.e, ip)
		if err != nil {
			d.e.Log.Errorf("Failed to get MAC for IP %s: %s", ip, err.Error())
			common.NewEmptyAPIResponse().WriteResponse(w, http.StatusInternalServerError)
			return
		} else if lease.ID == 0 {
			d.e.Log.Errorf("Attempted automatic registration on non-leased device %s", ip)
			common.NewAPIResponse("Error detecting MAC address", nil).WriteResponse(w, http.StatusBadRequest)
			return
		}
		mac = lease.MAC
	}

	// Get device from database
	device, err := models.GetDeviceByMAC(d.e, mac)
	if err != nil {
		d.e.Log.Errorf("Error getting device: %s", err.Error())
	}

	// Check if device is already registered
	if device.ID != 0 {
		d.e.Log.Noticef("Attempted duplicate registration of MAC %s to user %s", mac.String(), formUser.Username)
		common.NewAPIResponse("Device already registered", nil).WriteResponse(w, http.StatusConflict)
		return
	}

	// Check if the username is blacklisted
	// Administrators bypass the blacklist check
	if !sessionUser.IsAdmin() && formUser.IsBlacklisted() {
		d.e.Log.Noticef("Attempted registration by blacklisted user %s", formUser.Username)
		common.NewAPIResponse("Username blacklisted", nil).WriteResponse(w, http.StatusForbidden)
		return
	}

	// Validate platform, we don't want someone to submit an inappropiate value
	platform := ""
	if manual {
		if common.StringInSlice(r.FormValue("platform"), d.e.Config.Registration.ManualRegPlatforms) {
			platform = r.FormValue("platform")
		}
	} else {
		platform = common.ParseUserAgent(r.UserAgent())
	}

	// Fill in device information
	device.Username = formUser.Username
	device.RegisteredFrom = ip
	device.Platform = platform
	device.Expires = formUser.DeviceExpiration.NextExpiration(d.e)
	device.DateRegistered = time.Now()
	if !manual {
		device.UserAgent = r.UserAgent()
	} else {
		device.UserAgent = "Manual"
	}

	// Save new device
	if err := device.Save(); err != nil {
		d.e.Log.Errorf("Error registering device: %s", err.Error())
		common.NewAPIResponse("Error registering device", nil).WriteResponse(w, http.StatusInternalServerError)
	}
	d.e.Log.Infof("Successfully registered MAC %s to user %s", mac.String(), formUser.Username)

	// Redirect client as needed
	resp := struct{ Location string }{Location: "/manage"}
	if sessionUser.IsAdmin() {
		resp.Location = "/admin/manage/" + formUser.Username
	}

	common.NewAPIResponse("Registration successful", resp).WriteResponse(w, http.StatusOK)
}

func (d *Device) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	sessionUser := models.GetUserFromContext(r)
	formUser := sessionUser
	if r.FormValue("username") != sessionUser.Username {
		if !sessionUser.IsAdmin() {
			common.NewAPIResponse("Admin Error", nil).WriteResponse(w, http.StatusForbidden)
			return
		}
		var err error
		formUser, err = models.GetUserByUsername(d.e, r.FormValue("username"))
		if err != nil {
			d.e.Log.Errorf("Error getting user: %s", err.Error())
			common.NewAPIResponse("Server error", nil).WriteResponse(w, http.StatusInternalServerError)
			return
		}
	}

	deleteAll := (r.FormValue("mac") == "")
	macsToDelete := strings.Split(r.FormValue("mac"), ",")
	usersDevices, err := models.GetDevicesForUser(d.e, formUser)
	if err != nil {
		d.e.Log.Errorf("Error deleting devices: %s", err.Error())
		common.NewAPIResponse("Error deleting devices", nil).WriteResponse(w, http.StatusInternalServerError)
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
		common.NewAPIResponse("Finished but with errors", nil).WriteResponse(w, http.StatusOK)
		return
	}

	common.NewAPIResponse("Devices deleted successful", nil).WriteResponse(w, http.StatusNoContent)
}
