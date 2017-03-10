// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/lfkeitel/verbose"
	"github.com/usi-lfkeitel/packet-guardian/src/common"
	"github.com/usi-lfkeitel/packet-guardian/src/models"
)

type Device struct {
	e *common.Environment
}

func NewDeviceController(e *common.Environment) *Device {
	return &Device{e: e}
}

func (d *Device) RegistrationHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Check authentication and get User models
	macPost := r.FormValue("mac-address")
	manual := (macPost != "")
	var formUser *models.User // The user to whom the device is being registered
	var err error
	sessionUser := models.GetUserFromContext(r)
	formUsername := strings.ToLower(r.FormValue("username"))
	if formUsername == "" {
		common.NewAPIResponse("No username given", nil).WriteResponse(w, http.StatusBadRequest)
		return
	}

	if sessionUser.Username != formUsername && !sessionUser.Can(models.CreateDevice) {
		d.e.Log.WithFields(verbose.Fields{
			"package":    "controllers:api:device",
			"username":   formUser,
			"changed-by": sessionUser.Username,
		}).Notice("Admin register action attempted`")
		common.NewAPIResponse("Permission denied", nil).WriteResponse(w, http.StatusForbidden)
		return
	}

	if formUsername == sessionUser.Username {
		if manual && !sessionUser.Can(models.CreateOwn) {
			common.NewAPIResponse("Cannot manually register device - Permission denied", nil).WriteResponse(w, http.StatusForbidden)
			return
		}
		if !manual && !sessionUser.Can(models.AutoRegOwn) {
			common.NewAPIResponse("Cannot automatically register device - Permission denied", nil).WriteResponse(w, http.StatusForbidden)
			return
		}
		formUser = sessionUser
	} else {
		var err error
		formUser, err = models.GetUserByUsername(d.e, formUsername)
		if err != nil {
			d.e.Log.WithFields(verbose.Fields{
				"error":    err,
				"package":  "controllers:api:device",
				"username": formUsername,
			}).Error("Error getting user")
			common.NewAPIResponse("Error registering device", nil).WriteResponse(w, http.StatusInternalServerError)
			return
		}
		// Be careful with this, if this goes outside this if, it may release the sessionUser prematurely.
		defer formUser.Release()
	}

	// CreateDevice is the administrative permision
	if !sessionUser.Can(models.CreateDevice) {
		if formUser.IsBlacklisted() {
			d.e.Log.WithFields(verbose.Fields{
				"package":  "controllers:api:device",
				"username": formUser.Username,
			}).Error("Attempted registration by blacklisted user")
			common.NewAPIResponse("Username blacklisted", nil).WriteResponse(w, http.StatusForbidden)
			return
		}

		// Get and enforce the device limit
		limit := models.UserDeviceLimit(d.e.Config.Registration.DefaultDeviceLimit)
		if formUser.DeviceLimit != models.UserDeviceLimitGlobal {
			limit = formUser.DeviceLimit
		}

		deviceCount, err := models.GetDeviceCountForUser(d.e, formUser)
		if err != nil {
			d.e.Log.WithFields(verbose.Fields{
				"package": "controllers:api:device",
				"error":   err,
			}).Error("Error getting device count")
		}
		if limit != models.UserDeviceLimitUnlimited && deviceCount >= int(limit) {
			common.NewAPIResponse("Device limit reached", nil).WriteResponse(w, http.StatusConflict)
			return
		}
	}

	// Get MAC address
	var mac net.HardwareAddr
	ip := common.GetIPFromContext(r)
	if manual {
		// Manual registration
		// if manual registeration are not allowed and not admin
		if !d.e.Config.Registration.AllowManualRegistrations && !sessionUser.Can(models.CreateDevice) {
			common.NewAPIResponse("Manual registrations not allowed", nil).WriteResponse(w, http.StatusForbidden)
			return
		}
		mac, err = common.FormatMacAddress(macPost)
		if err != nil {
			common.NewAPIResponse("Incorrect MAC address format", nil).WriteResponse(w, http.StatusBadRequest)
			return
		}
	} else {
		// Automatic registration
		lease, err := models.GetLeaseStore(d.e).GetLeaseByIP(ip)
		if err != nil {
			d.e.Log.WithFields(verbose.Fields{
				"error":   err,
				"package": "controllers:api:device",
				"ip":      ip.String(),
			}).Error("Error getting MAC for IP")
			common.NewAPIResponse("Failed detecting MAC address", nil).WriteResponse(w, http.StatusInternalServerError)
			return
		} else if lease.ID == 0 {
			d.e.Log.WithFields(verbose.Fields{
				"package": "controllers:api:device",
				"ip":      ip.String(),
			}).Notice("Attempted auto reg from non-leased device")
			common.NewAPIResponse("Error detecting MAC address", nil).WriteResponse(w, http.StatusBadRequest)
			return
		}
		mac = lease.MAC
	}

	// Get device from database
	device, err := models.GetDeviceByMAC(d.e, mac)
	if err != nil {
		d.e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "controllers:api:device",
			"mac":     mac.String(),
		}).Error("Error getting device")
		common.NewAPIResponse("Failed loading device", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}

	// Check if device is already registered
	if device.ID != 0 {
		d.e.Log.WithFields(verbose.Fields{
			"package":    "controllers:api:device",
			"mac":        mac.String(),
			"changed-by": sessionUser.Username,
			"username":   formUser.Username,
		}).Notice("Attempted duplicate registration")
		common.NewAPIResponse("Device already registered", nil).WriteResponse(w, http.StatusConflict)
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
	device.Description = r.FormValue("description")
	device.RegisteredFrom = ip
	device.Platform = platform
	device.Expires = formUser.DeviceExpiration.NextExpiration(d.e, time.Now())
	device.DateRegistered = time.Now()
	device.LastSeen = time.Now()
	device.UserAgent = r.UserAgent()
	if manual {
		device.UserAgent = "Manual"
	}

	// Save new device
	if err := device.Save(); err != nil {
		d.e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "controllers:api:device",
		}).Error("Error saving device")
		common.NewAPIResponse("Error saving device", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}
	d.e.Log.WithFields(verbose.Fields{
		"package":    "controllers:api:device",
		"mac":        mac.String(),
		"changed-by": sessionUser.Username,
		"username":   formUser.Username,
		"action":     "register_device",
		"manual":     manual,
	}).Info("Device registered")

	// Redirect client as needed
	resp := struct{ Location string }{Location: "/manage"}
	if sessionUser.Can(models.ViewDevices) {
		resp.Location = "/admin/manage/user/" + formUser.Username
	}

	common.NewAPIResponse("Registration successful", resp).WriteResponse(w, http.StatusOK)
}

func (d *Device) DeleteHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sessionUser := models.GetUserFromContext(r)
	formUser := sessionUser
	username := p.ByName("username")
	if username == "" {
		common.NewAPIResponse("No username given", nil).WriteResponse(w, http.StatusBadRequest)
		return
	}
	if username != sessionUser.Username {
		if !sessionUser.Can(models.DeleteDevice) {
			common.NewAPIResponse("Admin Error", nil).WriteResponse(w, http.StatusForbidden)
			return
		}
		var err error
		formUser, err = models.GetUserByUsername(d.e, username)
		if err != nil {
			d.e.Log.WithFields(verbose.Fields{
				"error":    err,
				"package":  "controllers:api:device",
				"username": username,
			}).Error("Error getting user")
			common.NewAPIResponse("Server error", nil).WriteResponse(w, http.StatusInternalServerError)
			return
		}
		defer formUser.Release()
	}

	if !sessionUser.Can(models.DeleteOwn) {
		common.NewAPIResponse("Permission denied", nil).WriteResponse(w, http.StatusForbidden)
		return
	}

	deleteAll := (r.FormValue("mac") == "")
	macsToDelete := strings.Split(r.FormValue("mac"), ",")
	usersDevices, err := models.GetDevicesForUser(d.e, formUser)
	if err != nil {
		d.e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "controllers:api:device",
		}).Error("Error getting devices")
		common.NewAPIResponse("Error deleting devices", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}

	finishedWithErrors := false
	for _, device := range usersDevices {
		if !deleteAll && !common.StringInSlice(device.MAC.String(), macsToDelete) {
			continue
		}

		// Protect blacklisted devices
		if device.IsBlacklisted() && !sessionUser.Can(models.ManageBlacklist) {
			d.e.Log.WithFields(verbose.Fields{
				"package":    "controllers:api:device",
				"mac":        device.MAC.String(),
				"changed-by": sessionUser.Username,
				"username":   formUser.Username,
			}).Notice("Attempted deleting a blacklisted device")
			continue
		}

		if err := device.Delete(); err != nil {
			d.e.Log.WithFields(verbose.Fields{
				"error":   err,
				"package": "controllers:api:device",
				"mac":     device.MAC.String(),
			}).Error("Error deleting device")
			finishedWithErrors = true
			continue
		}

		d.e.Log.WithFields(verbose.Fields{
			"package":    "controllers:api:device",
			"mac":        device.MAC.String(),
			"changed-by": sessionUser.Username,
			"username":   formUser.Username,
			"action":     "delete_device",
		}).Notice("Device deleted")
	}

	if finishedWithErrors {
		common.NewAPIResponse("Finished but with errors", nil).WriteResponse(w, http.StatusOK)
		return
	}

	common.NewAPIResponse("Devices deleted successful", nil).WriteResponse(w, http.StatusNoContent)
}

func (d *Device) ReassignHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sessionUser := models.GetUserFromContext(r)
	if !sessionUser.Can(models.ReassignDevice) {
		common.NewEmptyAPIResponse().WriteResponse(w, http.StatusForbidden)
		return
	}

	username := r.FormValue("username")
	if username == "" {
		common.NewAPIResponse("Username required", nil).WriteResponse(w, http.StatusBadRequest)
		return
	}

	devices := r.FormValue("macs")
	if devices == "" {
		common.NewAPIResponse("At least one MAC address is required", nil).WriteResponse(w, http.StatusBadRequest)
		return
	}

	user, err := models.GetUserByUsername(d.e, username)
	if err != nil {
		d.e.Log.WithFields(verbose.Fields{
			"error":    err,
			"package":  "controllers:api:device",
			"username": username,
		}).Error("Error getting user")
		common.NewAPIResponse("Server error", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}
	defer user.Release()

	devicesToReassign := strings.Split(devices, ",")
	for _, devMacStr := range devicesToReassign {
		devMacStr = strings.TrimSpace(devMacStr)
		mac, err := common.FormatMacAddress(devMacStr)
		if err != nil {
			common.NewAPIResponse("Malformed MAC address "+devMacStr, nil).WriteResponse(w, http.StatusBadRequest)
			return
		}
		dev, err := models.GetDeviceByMAC(d.e, mac)
		if err != nil {
			d.e.Log.WithFields(verbose.Fields{
				"error":   err,
				"package": "controllers:api:device",
				"mac":     mac.String(),
			}).Error("Error getting device")
			common.NewAPIResponse("Server error", nil).WriteResponse(w, http.StatusInternalServerError)
			return
		}
		if dev.ID == 0 { // Device doesn't exist
			d.e.Log.WithFields(verbose.Fields{
				"error":      err,
				"package":    "controllers:api:device",
				"mac":        dev.MAC.String(),
				"changed-by": sessionUser.Username,
			}).Error("Attempted reassigning unregistered device")
			common.NewAPIResponse("Device "+devMacStr+" isn't registered", nil).WriteResponse(w, http.StatusBadRequest)
			return
		}
		// Protect blacklisted devices
		if dev.IsBlacklisted() && !sessionUser.Can(models.ManageBlacklist) {
			d.e.Log.WithFields(verbose.Fields{
				"package":    "controllers:api:device",
				"changed-by": sessionUser.Username,
				"mac":        dev.MAC.String(),
			}).Error("Attempted reassigning a blacklisted device")
			continue
		}
		originalUser := dev.Username
		dev.Username = user.Username
		// Change expiration to reflect new owner
		dev.Expires = user.DeviceExpiration.NextExpiration(d.e, time.Now())
		if err := dev.Save(); err != nil {
			d.e.Log.WithFields(verbose.Fields{
				"error":   err,
				"package": "controllers:api:device",
			}).Error("Error saving device")
			common.NewAPIResponse("Error saving device", nil).WriteResponse(w, http.StatusInternalServerError)
			return
		}
		d.e.Log.WithFields(verbose.Fields{
			"changed-by":   sessionUser.Username,
			"new-username": username,
			"old-username": originalUser,
			"mac":          mac.String(),
			"action":       "reassign_device",
		}).Info("Reassigned device to another user")
	}

	common.NewAPIResponse("Devices reassigned successfully", nil).WriteResponse(w, http.StatusOK)
}

func (d *Device) EditDescriptionHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sessionUser := models.GetUserFromContext(r)
	mac, err := net.ParseMAC(p.ByName("mac"))
	if err != nil {
		common.NewAPIResponse("Invalid MAC address", nil).WriteResponse(w, http.StatusBadRequest)
	}

	device, err := models.GetDeviceByMAC(d.e, mac)
	if err != nil {
		d.e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "controllers:api:device",
			"mac":     mac.String(),
		}).Error("Error getting device")
		common.NewAPIResponse("Server error", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}

	if device.Username != sessionUser.Username {
		// Check admin privilages
		if !sessionUser.Can(models.EditDevice) {
			common.NewAPIResponse("Permission denied", nil).WriteResponse(w, http.StatusUnauthorized)
			return
		}
	} else {
		// Check user privilages
		deviceUser, err := models.GetUserByUsername(d.e, device.Username)
		if err != nil {
			d.e.Log.WithFields(verbose.Fields{
				"error":    err,
				"package":  "controllers:api:device",
				"username": device.Username,
			}).Error("Error getting user")
			common.NewAPIResponse("Server error", nil).WriteResponse(w, http.StatusInternalServerError)
			return
		}

		if !deviceUser.Can(models.EditOwn) {
			common.NewAPIResponse("Permission denied", nil).WriteResponse(w, http.StatusUnauthorized)
			return
		}
	}

	device.Description = r.FormValue("description")
	if err := device.Save(); err != nil {
		d.e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "controllers:api:device",
		}).Error("Error saving device")
		common.NewAPIResponse("Error saving device", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}

	d.e.Log.WithFields(verbose.Fields{
		"mac":        device.MAC.String(),
		"username":   device.Username,
		"changed-by": sessionUser.Username,
		"package":    "controllers:api:device",
		"action":     "edit_desc_device",
	}).Info("Device description changed")
	common.NewAPIResponse("Device saved successfully", nil).WriteResponse(w, http.StatusOK)
}

func (d *Device) EditExpirationHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sessionUser := models.GetUserFromContext(r)
	var err error
	mac, err := net.ParseMAC(p.ByName("mac"))
	if err != nil {
		common.NewAPIResponse("Invalid MAC address", nil).WriteResponse(w, http.StatusBadRequest)
	}

	device, err := models.GetDeviceByMAC(d.e, mac)
	if err != nil {
		d.e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "controllers:api:device",
			"mac":     mac.String(),
		}).Error("Error getting device")
		common.NewAPIResponse("Server error", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}

	if !sessionUser.Can(models.EditDevice) {
		common.NewAPIResponse("Permission denied", nil).WriteResponse(w, http.StatusUnauthorized)
		return
	}

	expType := r.FormValue("type")
	expValue := r.FormValue("value")

	newExpire := &models.UserDeviceExpiration{}
	switch expType {
	case "global":
		newExpire.Mode = models.UserDeviceExpirationGlobal
	case "never":
		newExpire.Mode = models.UserDeviceExpirationNever
	case "rolling":
		newExpire.Mode = models.UserDeviceExpirationRolling
	case "specific":
		newExpire.Mode = models.UserDeviceExpirationSpecific
		expTime, err := time.ParseInLocation(common.TimeFormat, expValue, time.Local)
		if err != nil {
			d.e.Log.WithFields(verbose.Fields{
				"error":   err,
				"package": "controllers:api:device",
			}).Error("Error parsing time")
			common.NewAPIResponse("Error saving device", nil).WriteResponse(w, http.StatusInternalServerError)
			return
		}
		newExpire.Value = expTime.Unix()
	default:
		common.NewAPIResponse("Invalid expiration type", nil).WriteResponse(w, http.StatusBadRequest)
		return
	}
	newExpireResp := newExpire.String()

	device.Expires = newExpire.NextExpiration(d.e, time.Now())
	if err := device.Save(); err != nil {
		d.e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "controllers:api:device",
		}).Error("Error saving device")
		common.NewAPIResponse("Error saving device", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}

	if newExpire.Mode == models.UserDeviceExpirationGlobal {
		newExpireResp = models.GetGlobalDefaultExpiration(d.e).String()
	} else if newExpire.Mode == models.UserDeviceExpirationSpecific {
		newExpireResp = device.Expires.Format(common.TimeFormat)
	}

	d.e.Log.WithFields(verbose.Fields{
		"mac":        device.MAC.String(),
		"username":   device.Username,
		"changed-by": sessionUser.Username,
		"expiration": newExpireResp,
		"package":    "controllers:api:device",
		"action":     "edit_exp_device",
	}).Info("Device expiration changed")
	resp := map[string]string{"newExpiration": newExpireResp}
	common.NewAPIResponse("Device saved successfully", resp).WriteResponse(w, http.StatusOK)
}
