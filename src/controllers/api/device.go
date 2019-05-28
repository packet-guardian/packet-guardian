// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/lfkeitel/verbose/v4"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
	"github.com/packet-guardian/useragent"
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
		formUser, err = stores.GetUserStore(d.e).GetUserByUsername(formUsername)
		if err != nil {
			d.e.Log.WithFields(verbose.Fields{
				"error":    err,
				"package":  "controllers:api:device",
				"username": formUsername,
			}).Error("Error getting user")
			common.NewAPIResponse("Error registering device", nil).WriteResponse(w, http.StatusInternalServerError)
			return
		}
	}

	// CreateDevice is the administrative permision
	if !sessionUser.Can(models.CreateDevice) {
		httpCode, err := d.checkCanRegister(formUser)
		if err != nil {
			common.NewAPIResponse(err.Error(), nil).WriteResponse(w, httpCode)
			return
		}
	}

	// Get MAC address
	ip := common.GetIPFromContext(r)
	mac, httpCode, err := d.getRegMACAddress(manual, ip, macPost, sessionUser)
	if err != nil {
		common.NewAPIResponse(err.Error(), nil).WriteResponse(w, httpCode)
		return
	}

	// Get device from database
	device, err := stores.GetDeviceStore(d.e).GetDeviceByMAC(mac)
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

	// Validate platform, we don't want someone to submit an inappropriate value
	platform := ""
	if manual {
		if common.StringInSlice(r.FormValue("platform"), d.e.Config.Registration.ManualRegPlatforms) {
			platform = r.FormValue("platform")
		}
	} else {
		platform = useragent.ParseUserAgent(r.UserAgent()).String()
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

func (d *Device) checkCanRegister(formUser *models.User) (int, error) {
	if formUser.IsBlacklisted() {
		d.e.Log.WithFields(verbose.Fields{
			"package":  "controllers:api:device",
			"username": formUser.Username,
		}).Error("Attempted registration by blacklisted user")
		return http.StatusForbidden, errors.New("Username blacklisted")
	}

	// Get and enforce the device limit
	limit := models.UserDeviceLimit(d.e.Config.Registration.DefaultDeviceLimit)
	if formUser.DeviceLimit != models.UserDeviceLimitGlobal {
		limit = formUser.DeviceLimit
	}

	// If user's limit is unlimited, bypass device count
	if limit != models.UserDeviceLimitUnlimited {
		deviceCount, err := stores.GetDeviceStore(d.e).GetDeviceCountForUser(formUser)
		if err != nil {
			d.e.Log.WithFields(verbose.Fields{
				"package": "controllers:api:device",
				"error":   err,
			}).Error("Error getting device count")
		}
		if deviceCount >= int(limit) {
			return http.StatusConflict, errors.New("Device limit reached")
		}
	}
	return 0, nil
}

func (d *Device) getRegMACAddress(manual bool, ip net.IP, macPost string, sessionUser *models.User) (net.HardwareAddr, int, error) {
	if manual {
		// Manual registration
		// if manual registeration are not allowed and not admin
		if !d.e.Config.Registration.AllowManualRegistrations && !sessionUser.Can(models.CreateDevice) {
			return nil, http.StatusForbidden, errors.New("Manual registrations not allowed")
		}
		mac, err := common.FormatMacAddress(macPost)
		if err != nil {
			return nil, http.StatusBadRequest, errors.New("Incorrect MAC address format")
		}
		return mac, 0, nil
	}

	// Automatic registration
	lease, err := stores.GetLeaseStore(d.e).GetLeaseByIP(ip)
	if err != nil {
		d.e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "controllers:api:device",
			"ip":      ip.String(),
		}).Error("Error getting MAC for IP")
		return nil, http.StatusInternalServerError, errors.New("Failed detecting MAC address")
	}

	if lease.ID == 0 {
		d.e.Log.WithFields(verbose.Fields{
			"package": "controllers:api:device",
			"ip":      ip.String(),
		}).Notice("Attempted auto reg from non-leased device")
		return nil, http.StatusInternalServerError, errors.New("Error detecting MAC address")
	}
	return lease.MAC, 0, nil
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
		formUser, err = stores.GetUserStore(d.e).GetUserByUsername(username)
		if err != nil {
			d.e.Log.WithFields(verbose.Fields{
				"error":    err,
				"package":  "controllers:api:device",
				"username": username,
			}).Error("Error getting user")
			common.NewAPIResponse("Server error", nil).WriteResponse(w, http.StatusInternalServerError)
			return
		}
	}

	if !sessionUser.Can(models.DeleteOwn) {
		common.NewAPIResponse("Permission denied", nil).WriteResponse(w, http.StatusForbidden)
		return
	}

	deleteAll := (r.FormValue("mac") == "")
	macsToDelete := strings.Split(r.FormValue("mac"), ",")
	usersDevices, err := stores.GetDeviceStore(d.e).GetDevicesForUser(formUser)
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

	user, err := stores.GetUserStore(d.e).GetUserByUsername(username)
	if err != nil {
		d.e.Log.WithFields(verbose.Fields{
			"error":    err,
			"package":  "controllers:api:device",
			"username": username,
		}).Error("Error getting user")
		common.NewAPIResponse("Server error", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}

	devicesToReassign := strings.Split(devices, ",")
	for _, devMacStr := range devicesToReassign {
		devMacStr = strings.TrimSpace(devMacStr)
		mac, err := common.FormatMacAddress(devMacStr)
		if err != nil {
			common.NewAPIResponse("Malformed MAC address "+devMacStr, nil).WriteResponse(w, http.StatusBadRequest)
			return
		}
		dev, err := stores.GetDeviceStore(d.e).GetDeviceByMAC(mac)
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

	device, err := stores.GetDeviceStore(d.e).GetDeviceByMAC(mac)
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
		// Check admin privileges
		if !sessionUser.Can(models.EditDevice) {
			common.NewAPIResponse("Permission denied", nil).WriteResponse(w, http.StatusUnauthorized)
			return
		}
	} else {
		// Check user privileges
		deviceUser, err := stores.GetUserStore(d.e).GetUserByUsername(device.Username)
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

	device, err := stores.GetDeviceStore(d.e).GetDeviceByMAC(mac)
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

func (d *Device) GetDeviceHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sessionUser := models.GetUserFromContext(r)
	macParam := p.ByName("mac")

	mac, err := net.ParseMAC(macParam)
	if err != nil {
		http.Error(w, "Invalid mac address", http.StatusBadRequest)
		return
	}

	device, err := stores.GetDeviceStore(d.e).GetDeviceByMAC(mac)
	if err != nil {
		http.Error(w, "Error getting device from database", http.StatusInternalServerError)
		return
	}

	if device.Username != sessionUser.Username && !sessionUser.Can(models.ViewDevices) {
		common.NewAPIResponse("Unauthorized", nil).WriteResponse(w, http.StatusUnauthorized)
		return
	}

	if device.ID == 0 {
		common.NewAPIResponse("Device not found", nil).WriteResponse(w, http.StatusNotFound)
		return
	}

	common.NewAPIResponse("Registration successful", device).WriteResponse(w, http.StatusOK)
}

func (d *Device) EditFlaggedHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sessionUser := models.GetUserFromContext(r)
	var err error
	mac, err := net.ParseMAC(p.ByName("mac"))
	if err != nil {
		common.NewAPIResponse("Invalid MAC address", nil).WriteResponse(w, http.StatusBadRequest)
	}

	device, err := stores.GetDeviceStore(d.e).GetDeviceByMAC(mac)
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

	flagged := r.FormValue("flagged")
	if flagged != "" {
		device.Flagged = (flagged == "1" || flagged == "true")
	}

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
		"flagged":    device.Flagged,
		"package":    "controllers:api:device",
		"action":     "edit_flagged_device",
	}).Info("Device flagged status changed")
	common.NewAPIResponse("Device saved successfully", nil).WriteResponse(w, http.StatusOK)
}
