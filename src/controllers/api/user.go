// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/models"
)

type User struct {
	e *common.Environment
}

func NewUserController(e *common.Environment) *User {
	return &User{e: e}
}

func (u *User) UserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		u.saveUserHandler(w, r)
	} else if r.Method == "DELETE" {
		u.deleteUserHandler(w, r)
	}
}

func (u *User) saveUserHandler(w http.ResponseWriter, r *http.Request) {
	sessionUser := models.GetUserFromContext(r)
	username := strings.ToLower(r.FormValue("username"))
	if username == "" {
		common.NewAPIResponse("Username required", nil).WriteResponse(w, http.StatusBadRequest)
		return
	}

	user, err := models.GetUserByUsername(u.e, username)
	if err != nil {
		u.e.Log.Errorf("Error saving user: %s", err.Error())
		common.NewAPIResponse("Error saving user", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}

	canCreate := sessionUser.Can(models.CreateUser)
	canEdit := sessionUser.Can(models.EditUser)
	if !(user.IsNew() && canCreate) && !(!user.IsNew() && canEdit) {
		common.NewAPIResponse("Permission denied", nil).WriteResponse(w, http.StatusForbidden)
		return
	}

	// Password
	password := r.FormValue("password")
	if password != "" {
		if password == "-1" {
			user.RemovePassword()
		} else {
			user.NewPassword(password)
		}
	}

	// Device limit
	limitStr := r.FormValue("device_limit")
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			u.e.Log.Errorf("Error saving user: %s", err.Error())
			common.NewAPIResponse("device_limit must be a number", nil).WriteResponse(w, http.StatusBadRequest)
			return
		}
		user.DeviceLimit = models.UserDeviceLimit(limit)
	}

	// Default device expiration
	expTypeStr := r.FormValue("expiration_type")
	devExpiration := r.FormValue("device_expiration")
	updateDeviceExpirations := false
	if expTypeStr != "" || devExpiration != "" {
		if expTypeStr == "" && devExpiration != "" {
			common.NewAPIResponse("Error saving user: Expiration type not given", nil).WriteResponse(w, http.StatusBadRequest)
			return
		}

		expType, err := strconv.Atoi(expTypeStr)
		if err != nil {
			u.e.Log.Errorf("Error saving user: %s", err.Error())
			common.NewAPIResponse("Expiration type must be an integer", nil).WriteResponse(w, http.StatusBadRequest)
			return
		}
		if user.DeviceExpiration.Mode != models.UserExpiration(expType) {
			user.DeviceExpiration.Mode = models.UserExpiration(expType)
			updateDeviceExpirations = true
		}
		if user.DeviceExpiration.Mode == models.UserDeviceExpirationGlobal ||
			user.DeviceExpiration.Mode == models.UserDeviceExpirationNever ||
			user.DeviceExpiration.Mode == models.UserDeviceExpirationRolling {
			user.DeviceExpiration.Value = 0
		} else if user.DeviceExpiration.Mode == models.UserDeviceExpirationSpecific {
			t, err := time.ParseInLocation(common.TimeFormat, devExpiration, time.Local)
			if err != nil {
				common.NewAPIResponse("Invalid time format", nil).WriteResponse(w, http.StatusBadRequest)
				return
			}
			if user.DeviceExpiration.Value != t.Unix() {
				user.DeviceExpiration.Value = t.Unix()
				updateDeviceExpirations = true
			}
		} else if user.DeviceExpiration.Mode == models.UserDeviceExpirationDaily {
			secs, err := common.ParseTime(devExpiration)
			if err != nil {
				common.NewAPIResponse("Invalid time format", nil).WriteResponse(w, http.StatusBadRequest)
				return
			}
			if user.DeviceExpiration.Value != secs {
				user.DeviceExpiration.Value = secs
				updateDeviceExpirations = true
			}
		} else if user.DeviceExpiration.Mode == models.UserDeviceExpirationDuration {
			d, err := time.ParseDuration(devExpiration)
			if err != nil {
				common.NewAPIResponse("Invalid duration", nil).WriteResponse(w, http.StatusBadRequest)
				return
			}
			// time.Duration's Second() returns a float that contains the nanoseconds as well
			// For sanity we don't care and don't want the nanoseconds
			dur := int64(d / time.Second)
			if user.DeviceExpiration.Value != dur {
				user.DeviceExpiration.Value = dur
				updateDeviceExpirations = true
			}
		} else {
			common.NewAPIResponse("Invalid device expiration type", nil).WriteResponse(w, http.StatusBadRequest)
			return
		}
	}

	// Valid dates
	validStart := r.FormValue("valid_start")
	validEnd := r.FormValue("valid_end")
	if validStart == "0" && validEnd == "0" {
		user.ValidForever = true
		user.ValidStart = time.Unix(0, 0)
		user.ValidEnd = user.ValidStart
	} else {
		user.ValidForever = false
		if validStart != "" {
			t, err := time.Parse(common.TimeFormat, validStart)
			if err != nil {
				common.NewAPIResponse("Invalid time format: valid_start", nil).WriteResponse(w, http.StatusBadRequest)
				return
			}
			user.ValidStart = t
		}
		if validEnd != "" {
			t, err := time.Parse(common.TimeFormat, validEnd)
			if err != nil {
				common.NewAPIResponse("Invalid time format: valid_end", nil).WriteResponse(w, http.StatusBadRequest)
				return
			}
			user.ValidEnd = t
		}
		if user.ValidEnd.Before(user.ValidStart) {
			common.NewAPIResponse("valid_start must be before valid_end", nil).WriteResponse(w, http.StatusBadRequest)
			return
		}
	}

	canManage := r.FormValue("can_manage")
	if canManage != "" {
		user.CanManage = (canManage == "1")
	}

	isNewUser := user.IsNew() // This will always be false after a call to Save()

	if err := user.Save(); err != nil {
		u.e.Log.Errorf("Error saving user: %s", err.Error())
		common.NewAPIResponse("Error saving user", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}

	if isNewUser {
		u.e.Log.Infof("Admin %s created user %s", models.GetUserFromContext(r).Username, user.Username)
	} else {
		u.e.Log.Infof("Admin %s edited user %s", models.GetUserFromContext(r).Username, user.Username)
	}

	if updateDeviceExpirations {
		devices, err := models.GetDevicesForUser(u.e, user)
		if err != nil {
			u.e.Log.WithField("Err", err).Error("Failed to update user's device expirations")
			common.NewAPIResponse("User saved, but devices not updated", nil).WriteResponse(w, http.StatusInternalServerError)
			return
		}
		for _, d := range devices {
			d.Expires = user.DeviceExpiration.NextExpiration(u.e, d.DateRegistered)
			if err := d.Save(); err != nil {
				u.e.Log.WithField("Err", err).Error("Failed to save device")
				common.NewAPIResponse("User saved, but some devices not updated", nil).WriteResponse(w, http.StatusInternalServerError)
				return
			}
		}
	}

	if isNewUser {
		common.NewAPIResponse("User created successfully", nil).WriteResponse(w, http.StatusNoContent)
		return
	}
	common.NewAPIResponse("User saved successfully", nil).WriteResponse(w, http.StatusNoContent)
}

func (u *User) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	sessionUser := models.GetUserFromContext(r)
	if !sessionUser.Can(models.DeleteUser) {
		common.NewAPIResponse("Permission denied", nil).WriteResponse(w, http.StatusForbidden)
		return
	}

	username := r.FormValue("username")
	if username == "" {
		common.NewAPIResponse("Username required", nil).WriteResponse(w, http.StatusBadRequest)
		return
	}

	user, err := models.GetUserByUsername(u.e, username)
	if err != nil {
		u.e.Log.Errorf("Error getting user for deletion: %s", err.Error())
		common.NewAPIResponse("Error deleting user", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}

	if err := user.Delete(); err != nil {
		u.e.Log.Errorf("Error deleting user: %s", err.Error())
		common.NewAPIResponse("Error deleting user", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}
	u.e.Log.Infof("Admin %s deleted user %s", models.GetUserFromContext(r).Username, user.Username)
	common.NewAPIResponse("User deleted", nil).WriteResponse(w, http.StatusNoContent)
}
