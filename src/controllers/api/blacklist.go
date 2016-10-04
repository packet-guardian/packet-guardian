// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/lfkeitel/verbose"
	"github.com/usi-lfkeitel/packet-guardian/src/common"
	"github.com/usi-lfkeitel/packet-guardian/src/models"
)

type Blacklist struct {
	e *common.Environment
}

func NewBlacklistController(e *common.Environment) *Blacklist {
	return &Blacklist{e: e}
}

func (b *Blacklist) BlacklistUserHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if !models.GetUserFromContext(r).Can(models.ManageBlacklist) {
		common.NewAPIResponse("Permission denied", nil).WriteResponse(w, http.StatusForbidden)
		return
	}

	username := p.ByName("username")
	if username == "" {
		common.NewAPIResponse("No username given", nil).WriteResponse(w, http.StatusBadRequest)
		return
	}

	user, err := models.GetUserByUsername(b.e, username)
	if err != nil {
		b.e.Log.WithFields(verbose.Fields{
			"username": username,
			"error":    err,
			"package":  "controllers:api:blacklist",
		}).Error("Error getting user")
		common.NewAPIResponse("Error blacklisting user", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}
	defer user.Release()

	if r.Method == "POST" {
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
	if !sessionUser.Can(models.ManageBlacklist) {
		common.NewAPIResponse("Permission denied", nil).WriteResponse(w, http.StatusForbidden)
		return
	}

	username := p.ByName("username")
	if username == "" {
		common.NewAPIResponse("No username given", nil).WriteResponse(w, http.StatusBadRequest)
		return
	}

	user, err := models.GetUserByUsername(b.e, username)
	if err != nil {
		b.e.Log.WithFields(verbose.Fields{
			"username": username,
			"error":    err,
			"package":  "controllers:api:blacklist",
		}).Error("Error getting user")
		common.NewAPIResponse("Error blacklisting user", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}
	defer user.Release()

	blacklistAll := (r.FormValue("mac") == "")
	macsToBlacklist := strings.Split(r.FormValue("mac"), ",")
	usersDevices, err := models.GetDevicesForUser(b.e, user)
	if err != nil {
		b.e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "controllers:api:blacklist",
		}).Error("Error blacklisting devices")
		common.NewAPIResponse("Error blacklisting devices", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}

	finishedWithErrors := false
	for _, device := range usersDevices {
		if !blacklistAll && !common.StringInSlice(device.MAC.String(), macsToBlacklist) {
			continue
		}

		// Device is already in the state we want
		if device.IsBlacklisted() == (r.Method == "POST") {
			continue
		}

		device.SetBlacklist(r.Method == "POST")
		if err := device.Save(); err != nil {
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
				"username":   user.Username,
			}).Info("Device added to blacklist")
		} else {
			b.e.Log.WithFields(verbose.Fields{
				"package":    "controllers:api:blacklist",
				"action":     "unblacklist",
				"mac":        device.MAC.String(),
				"changed-by": sessionUser.Username,
				"username":   user.Username,
			}).Info("Device removed from blacklist")
		}
	}

	if finishedWithErrors {
		common.NewAPIResponse("Finished but with errors", nil).WriteResponse(w, http.StatusNoContent)
		return
	}

	if r.Method == "POST" {
		common.NewAPIResponse("Devices blacklisted successful", nil).WriteResponse(w, http.StatusNoContent)
		return
	}

	common.NewAPIResponse("Devices removed from blacklist successful", nil).WriteResponse(w, http.StatusNoContent)
}
