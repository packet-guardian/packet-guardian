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

func (b *Blacklist) BlacklistHandler(w http.ResponseWriter, r *http.Request) {
	reqType := mux.Vars(r)["type"]
	if reqType == "user" {
		b.blacklistUser(w, r)
	} else if reqType == "device" {
		b.blacklistDevice(w, r)
	} else {
		common.NewAPIResponse(common.APIStatusMalformedRequest, "Invalid blacklist type", nil).WriteTo(w)
	}
}

func (b *Blacklist) blacklistUser(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	if username == "" {
		common.NewAPIResponse(common.APIStatusMalformedRequest, "No username given", nil).WriteTo(w)
		return
	}

	user, err := models.GetUserByUsername(b.e, username)
	if err != nil {
		b.e.Log.Errorf("Error getting user: %s", err.Error())
		common.NewAPIResponse(common.APIStatusGenericError, "Error blacklisting user", nil).WriteTo(w)
		return
	}

	if r.Method == "POST" {
		user.Blacklist()
	} else if r.Method == "DELETE" {
		user.Unblacklist()
	}

	if err := user.SaveToBlacklist(); err != nil {
		b.e.Log.Errorf("Error blacklisting user: %s", err.Error())
		common.NewAPIResponse(common.APIStatusGenericError, "Error blacklisting user", nil).WriteTo(w)
		return
	}

	if r.Method == "POST" {
		b.e.Log.Infof("Admin %s blacklisted user %s", models.GetUserFromContext(r).Username, user.Username)
		common.NewAPIOK("User blacklisted", nil).WriteTo(w)
	} else if r.Method == "DELETE" {
		b.e.Log.Infof("Admin %s unblacklisted user %s", models.GetUserFromContext(r).Username, user.Username)
		common.NewAPIOK("User removed from blacklist", nil).WriteTo(w)
	}

}

func (b *Blacklist) blacklistDevice(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	user, err := models.GetUserByUsername(b.e, username)
	if err != nil {
		b.e.Log.Errorf("Error getting user: %s", err.Error())
		common.NewAPIResponse(common.APIStatusGenericError, "Error blacklisting user", nil).WriteTo(w)
		return
	}

	blacklistAll := (r.FormValue("mac") == "")
	macsToBlacklist := strings.Split(r.FormValue("mac"), ",")
	usersDevices, err := models.GetDevicesForUser(b.e, user)
	if err != nil {
		b.e.Log.Errorf("Error blacklisting devices: %s", err.Error())
		common.NewAPIResponse(common.APIStatusGenericError, "Error blacklisting devices", nil).WriteTo(w)
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
		common.NewAPIResponse(common.APIStatusGenericError, "Finished but with errors", nil).WriteTo(w)
		return
	}

	if r.Method == "POST" {
		common.NewAPIOK("Devices blacklisted successful", nil).WriteTo(w)
		return
	}

	common.NewAPIOK("Devices removed from blacklist successful", nil).WriteTo(w)
}
