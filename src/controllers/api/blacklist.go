// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/lfkeitel/verbose/v4"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
)

var errInvalidMAC = errors.New("Incorrect MAC address format")

type Blacklist struct {
	e       *common.Environment
	users   stores.UserStore
	devices stores.DeviceStore
}

func NewBlacklistController(e *common.Environment, us stores.UserStore, ds stores.DeviceStore) *Blacklist {
	return &Blacklist{
		e:       e,
		users:   us,
		devices: ds,
	}
}

func (b *Blacklist) BlacklistUserHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	username := p.ByName("username")
	if username == "" {
		common.NewAPIResponse("No username given", nil).WriteResponse(w, http.StatusBadRequest)
		return
	}

	user, err := b.users.GetUserByUsername(username)
	if err != nil {
		b.e.Log.WithFields(verbose.Fields{
			"username": username,
			"error":    err,
			"package":  "controllers:api:blacklist",
		}).Error("Error getting user")
		common.NewAPIResponse("Error blacklisting user", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" && !user.IsBlacklisted() {
		user.Blacklist()
	} else if r.Method == "DELETE" {
		user.Unblacklist()
	}

	if err := user.SaveToBlacklist(); err != nil {
		b.e.Log.WithFields(verbose.Fields{
			"username": user.Username,
			"error":    err,
			"package":  "controllers:api:blacklist",
		}).Error("Error blacklisting user")
		common.NewAPIResponse("Error blacklisting user", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" {
		b.e.Log.WithFields(verbose.Fields{
			"package":    "controllers:api:blacklist",
			"action":     "blacklist",
			"changed-by": models.GetUserFromContext(r).Username,
			"username":   user.Username,
		}).Info("User added to blacklisted")
		common.NewEmptyAPIResponse().WriteResponse(w, http.StatusNoContent)
	} else if r.Method == "DELETE" {
		b.e.Log.WithFields(verbose.Fields{
			"package":    "controllers:api:blacklist",
			"action":     "unblacklist",
			"changed-by": models.GetUserFromContext(r).Username,
			"username":   user.Username,
		}).Info("User removed from blacklist")
		common.NewEmptyAPIResponse().WriteResponse(w, http.StatusNoContent)
	}
}

func (b *Blacklist) BlacklistDeviceHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sessionUser := models.GetUserFromContext(r)

	//blacklistAll := (r.FormValue("mac") == "")
	macStr := r.FormValue("mac")
	addToBlacklist := (r.Method == "POST")

	finishedWithErrors := false
	devices := b.buildDeviceList(w, r, macStr, addToBlacklist)
	if devices == nil {
		return
	}

	// Blacklist selected devices
	for _, device := range devices {
		device.SetBlacklist(addToBlacklist)
		if err := device.SaveToBlacklist(); err != nil {
			b.e.Log.WithFields(verbose.Fields{
				"error":   err,
				"mac":     device.MAC.String(),
				"package": "controllers:api:blacklist",
			}).Error("Error blacklisting device")
			finishedWithErrors = true
			continue
		}

		if device.IsBlacklisted() {
			b.e.Log.WithFields(verbose.Fields{
				"package":    "controllers:api:blacklist",
				"action":     "blacklist",
				"mac":        device.MAC.String(),
				"changed-by": sessionUser.Username,
				"username":   device.GetUsername(),
			}).Info("Device added to blacklist")
		} else {
			b.e.Log.WithFields(verbose.Fields{
				"package":    "controllers:api:blacklist",
				"action":     "unblacklist",
				"mac":        device.MAC.String(),
				"changed-by": sessionUser.Username,
				"username":   device.GetUsername(),
			}).Info("Device removed from blacklist")
		}
	}

	if finishedWithErrors {
		common.NewAPIResponse("Finished but with errors", nil).WriteResponse(w, http.StatusNoContent)
		return
	}

	if addToBlacklist {
		common.NewAPIResponse("Devices blacklisted successful", nil).WriteResponse(w, http.StatusNoContent)
		return
	}

	common.NewAPIResponse("Devices removed from blacklist successful", nil).WriteResponse(w, http.StatusNoContent)
}

func (b *Blacklist) buildDeviceList(w http.ResponseWriter, r *http.Request, macStr string, addToBlacklist bool) []*models.Device {
	if macStr != "" {
		devices, err := b.getDevicesFromList(strings.Split(macStr, ","), addToBlacklist)
		if err != nil {
			if err == errInvalidMAC {
				common.NewAPIResponse(err.Error(), nil).WriteResponse(w, http.StatusBadRequest)
				return nil
			}
			b.e.Log.WithFields(verbose.Fields{
				"error":   err,
				"package": "controllers:api:blacklist",
			}).Error("Error blacklisting devices")
			common.NewAPIResponse("Error blacklisting devices", nil).WriteResponse(w, http.StatusInternalServerError)
			return nil
		}
		return devices
	}

	username := r.FormValue("username")
	if username == "" {
		common.NewAPIResponse("Username required to delete all devices", nil).WriteResponse(w, http.StatusBadRequest)
		return nil
	}

	var user *models.User
	user, err := b.users.GetUserByUsername(username)
	if err != nil {
		b.e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "controllers:api:blacklist",
		}).Error("Error blacklisting devices")
		common.NewAPIResponse("Error blacklisting devices", nil).WriteResponse(w, http.StatusInternalServerError)
		return nil
	}

	devices, err := b.devices.GetDevicesForUser(user)
	if err != nil {
		b.e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "controllers:api:blacklist",
		}).Error("Error blacklisting devices")
		common.NewAPIResponse("Error blacklisting devices", nil).WriteResponse(w, http.StatusInternalServerError)
		return nil
	}
	return devices
}

func (b *Blacklist) getDevicesFromList(l []string, add bool) ([]*models.Device, error) {
	var devices []*models.Device
	// Build list of devices to blacklist
	for _, deviceMAC := range l {
		mac, err := net.ParseMAC(deviceMAC)
		if err != nil {
			return nil, errInvalidMAC
		}

		device, err := b.devices.GetDeviceByMAC(mac)
		if err != nil {
			return nil, err
		}

		// Device is already in the state we want
		if device.IsBlacklisted() == add {
			continue
		}

		devices = append(devices, device)
	}
	return devices, nil
}

func (b *Blacklist) GetBlacklistHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	device := p.ByName("mac")

	if device == "" {
		// Return all blacklisted entities
	}

	// Return specific device
}
