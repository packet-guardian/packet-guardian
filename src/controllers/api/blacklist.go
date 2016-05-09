package api

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/models"
)

type Blacklist struct {
	e *common.Environment
}

func NewBlacklistController(e *common.Environment) *Blacklist {
	return &Blacklist{e: e}
}

func (b *Blacklist) BlacklistUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username, ok := vars["username"]
	if !ok {
		common.NewAPIResponse("No username given", nil).WriteResponse(w, http.StatusBadRequest)
		return
	}

	user, err := models.GetUserByUsername(b.e, username)
	if err != nil {
		b.e.Log.Errorf("Error getting user: %s", err.Error())
		common.NewAPIResponse("Error blacklisting user", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" {
		user.Blacklist()
	} else if r.Method == "DELETE" {
		user.Unblacklist()
	}

	if err := user.SaveToBlacklist(); err != nil {
		b.e.Log.Errorf("Error blacklisting user: %s", err.Error())
		common.NewAPIResponse("Error blacklisting user", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" {
		b.e.Log.Infof("Admin %s blacklisted user %s", models.GetUserFromContext(r).Username, user.Username)
		common.NewEmptyAPIResponse().WriteResponse(w, http.StatusNoContent)
	} else if r.Method == "DELETE" {
		b.e.Log.Infof("Admin %s unblacklisted user %s", models.GetUserFromContext(r).Username, user.Username)
		common.NewEmptyAPIResponse().WriteResponse(w, http.StatusNoContent)
	}

}

func (b *Blacklist) BlacklistDeviceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username, ok := vars["username"]
	if !ok {
		common.NewAPIResponse("No username given", nil).WriteResponse(w, http.StatusBadRequest)
		return
	}

	user, err := models.GetUserByUsername(b.e, username)
	if err != nil {
		b.e.Log.Errorf("Error getting user: %s", err.Error())
		common.NewAPIResponse("Error blacklisting user", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}

	blacklistAll := (r.FormValue("mac") == "")
	macsToBlacklist := strings.Split(r.FormValue("mac"), ",")
	usersDevices, err := models.GetDevicesForUser(b.e, user)
	if err != nil {
		b.e.Log.Errorf("Error blacklisting devices: %s", err.Error())
		common.NewAPIResponse("Error blacklisting devices", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}

	finishedWithErrors := false
	for _, device := range usersDevices {
		if !blacklistAll && !common.StringInSlice(device.MAC.String(), macsToBlacklist) {
			continue
		}

		device.IsBlacklisted = (r.Method == "POST")
		if err := device.Save(); err != nil {
			b.e.Log.Errorf("Error blacklisting device %s: %s", device.MAC.String(), err.Error())
			finishedWithErrors = true
			continue
		}

		if device.IsBlacklisted {
			b.e.Log.Infof("Blacklisted device %s for user %s", device.MAC.String(), user.Username)
		} else {
			b.e.Log.Infof("Removed device %s from blacklist for user %s", device.MAC.String(), user.Username)
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
