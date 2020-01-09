// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/lfkeitel/verbose/v4"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
)

type UserController struct {
	e       *common.Environment
	users   stores.UserStore
	devices stores.DeviceStore
}

func NewUserController(e *common.Environment, us stores.UserStore, ds stores.DeviceStore) *UserController {
	return &UserController{
		e:       e,
		users:   us,
		devices: ds,
	}
}

func (u *UserController) SaveUserHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sessionUser := models.GetUserFromContext(r)
	username := strings.ToLower(r.FormValue("username"))
	if username == "" {
		common.NewAPIResponse("Username required", nil).WriteResponse(w, http.StatusBadRequest)
		return
	}

	user, err := u.users.GetUserByUsername(username)
	if err != nil {
		u.e.Log.WithFields(verbose.Fields{
			"error":    err,
			"package":  "controllers:api:user",
			"username": username,
		}).Error("Error getting user")
		common.NewAPIResponse("Error saving user", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}

	canCreate := sessionUser.Can(models.CreateUser)
	canEdit := sessionUser.Can(models.EditUser)
	if !(user.IsNew() && canCreate) && !(!user.IsNew() && canEdit) {
		common.NewAPIResponse("Permission denied", nil).WriteResponse(w, http.StatusForbidden)
		return
	}

	// Permission groups
	uiGroup := r.FormValue("ui_group")
	apiGroup := r.FormValue("api_group")
	allowStatusAPI := r.FormValue("allow_status_api") == "1"
	if (user.UIGroup != uiGroup || user.APIGroup != apiGroup || user.AllowStatusAPI != allowStatusAPI) && !sessionUser.Can(models.EditUserPermissions) {
		common.NewAPIResponse("Permission denied", nil).WriteResponse(w, http.StatusForbidden)
		return
	}

	if !common.StringInSlice(uiGroup, []string{"default", "admin", "helpdesk", "readonly"}) {
		common.NewAPIResponse("Unknown ui group", nil).WriteResponse(w, http.StatusBadRequest)
		return
	}
	user.UIGroup = uiGroup

	if !common.StringInSlice(apiGroup, []string{"disable", "readonly-api", "readwrite-api"}) {
		common.NewAPIResponse("Unknown api group", nil).WriteResponse(w, http.StatusBadRequest)
		return
	}
	user.APIGroup = apiGroup
	user.AllowStatusAPI = allowStatusAPI

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
			t, err := time.ParseInLocation(common.TimeFormat, validStart, time.Local)
			if err != nil {
				common.NewAPIResponse("Invalid time format: valid_start", nil).WriteResponse(w, http.StatusBadRequest)
				return
			}
			user.ValidStart = t
		}
		if validEnd != "" {
			t, err := time.ParseInLocation(common.TimeFormat, validEnd, time.Local)
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

	canAutoreg := r.FormValue("can_autoreg")
	if canAutoreg != "" {
		user.CanAutoreg = (canAutoreg == "1")
	}

	// Delegates
	oldDelegates := make([]string, 0, len(user.Delegates))
	for d := range user.Delegates {
		oldDelegates = append(oldDelegates, d)
	}

	delegates := strings.Split(r.FormValue("delegates"), ",")
	newDelegates := make([]string, 0, len(delegates))
	for _, delegate := range delegates {
		if delegate == "" {
			continue
		}

		dsplit := strings.Split(delegate, ":")
		name := dsplit[0]
		permissions, exists := models.DelegatePermissions[dsplit[1]]
		if !exists {
			continue
		}

		user.Delegates[name] = permissions
		newDelegates = append(newDelegates, name)
	}

	deleteDelegates := make([]string, 1)
	for _, old := range oldDelegates {
		if !common.StringInSlice(old, newDelegates) {
			deleteDelegates = append(deleteDelegates, old)
		}
	}

	isNewUser := user.IsNew() // This will always be false after a call to Save()

	if err := user.Save(); err != nil {
		u.e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "controllers:api:user",
		}).Error("Error saving user")
		common.NewAPIResponse("Error saving user", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}

	for _, d := range deleteDelegates {
		u.users.DeleteDelegate(user, d)
	}

	if isNewUser {
		u.e.Log.WithFields(verbose.Fields{
			"package":    "controllers:api:user",
			"action":     "create_user",
			"username":   user.Username,
			"changed-by": sessionUser.Username,
		}).Info("User created")
	} else {
		u.e.Log.WithFields(verbose.Fields{
			"package":    "controllers:api:user",
			"action":     "edit_user",
			"username":   user.Username,
			"changed-by": sessionUser.Username,
		}).Info("User edited")
	}

	if updateDeviceExpirations {
		devices, err := u.devices.GetDevicesForUser(user)
		if err != nil {
			u.e.Log.WithFields(verbose.Fields{
				"error":   err,
				"package": "controllers:api:user",
			}).Error("Error getting devices")
			common.NewAPIResponse("User saved, but devices not updated", nil).WriteResponse(w, http.StatusInternalServerError)
			return
		}
		errOccured := false
		for _, d := range devices {
			d.Expires = user.DeviceExpiration.NextExpiration(u.e, d.DateRegistered)
			if err := d.Save(); err != nil {
				u.e.Log.WithFields(verbose.Fields{
					"error":   err,
					"package": "controllers:api:user",
					"mac":     d.MAC.String(),
				}).Error("Error saving device")
				errOccured = true
			}
		}
		if errOccured {
			common.NewAPIResponse("User saved, but some devices not updated", nil).WriteResponse(w, http.StatusInternalServerError)
			return
		}
	}

	if isNewUser {
		common.NewAPIResponse("User created successfully", nil).WriteResponse(w, http.StatusNoContent)
		return
	}
	common.NewAPIResponse("User saved successfully", nil).WriteResponse(w, http.StatusNoContent)
}

func (u *UserController) DeleteUserHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sessionUser := models.GetUserFromContext(r)
	username := r.FormValue("username")
	if username == "" {
		common.NewAPIResponse("Username required", nil).WriteResponse(w, http.StatusBadRequest)
		return
	}

	user, err := u.users.GetUserByUsername(username)
	if err != nil {
		u.e.Log.WithFields(verbose.Fields{
			"error":    err,
			"package":  "controllers:api:user",
			"username": username,
		}).Error("Error getting user")
		common.NewAPIResponse("Error deleting user", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}

	if err := user.Delete(); err != nil {
		u.e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "controllers:api:user",
		}).Error("Error deleting user")
		common.NewAPIResponse("Error deleting user", nil).WriteResponse(w, http.StatusInternalServerError)
		return
	}

	u.e.Log.WithFields(verbose.Fields{
		"package":    "controllers:api:user",
		"action":     "delete_user",
		"username":   user.Username,
		"changed-by": sessionUser.Username,
	}).Info("User deleted")
	common.NewAPIResponse("User deleted", nil).WriteResponse(w, http.StatusNoContent)
}

func (u *UserController) GetUserHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sessionUser := models.GetUserFromContext(r)
	usernameParam := p.ByName("username")

	user, err := u.users.GetUserByUsername(usernameParam)
	if err != nil {
		http.Error(w, "Error getting user from database", http.StatusInternalServerError)
		return
	}

	if user.Username != sessionUser.Username && !sessionUser.Can(models.ViewUsers) {
		common.NewAPIResponse("Unauthorized", nil).WriteResponse(w, http.StatusUnauthorized)
		return
	}

	common.NewAPIResponse("", user).WriteResponse(w, http.StatusOK)
}
