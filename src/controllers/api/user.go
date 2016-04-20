package api

import (
	"net/http"
	"strconv"
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
	common.NewAPIOK("User created", nil).WriteTo(w)
	return
	username := r.FormValue("username")
	if username == "" {
		// TODO: Return api error
		return
	}

	user, err := models.GetUserByUsername(u.e, username)
	if err != nil {
		u.e.Log.Errorf("Error saving user: %s", err.Error())
		// TODO: Return error
		return
	}

	// Password
	password := r.FormValue("password")
	if password != "" {
		if password == "-1" {
			user.RemovePassword()
		} else if len(password) > 8 {
			user.NewPassword(password)
		} else {
			u.e.Log.Error("Error saving user: password too short")
			// TODO: Return error
			return
		}
	}

	// Device limit
	limitStr := r.FormValue("device_limit")
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			u.e.Log.Errorf("Error saving user: %s", err.Error())
			// TODO: Return error
			return
		}
		user.DeviceLimit = models.UserDeviceLimit(limit)
	}

	// Default device expiration
	expTypeStr := r.FormValue("expiration_type")
	devExpiration := r.FormValue("device_expiration")
	if expTypeStr != "" && devExpiration != "" {
		expType, err := strconv.Atoi(expTypeStr)
		if err != nil {
			u.e.Log.Errorf("Error saving user: %s", err.Error())
			// TODO: Return error
			return
		}
		user.DeviceExpiration.Mode = models.UserExpiration(expType)
		if user.DeviceExpiration.Mode == models.UserDeviceExpirationGlobal ||
			user.DeviceExpiration.Mode == models.UserDeviceExpirationNever {
			user.DeviceExpiration.Value = 0
		} else if user.DeviceExpiration.Mode == models.UserDeviceExpirationSpecific {
			t, err := time.Parse(common.TimeFormat, devExpiration)
			if err != nil {
				u.e.Log.Errorf("Error saving user: %s", err.Error())
				// TODO: Return error
				return
			}
			user.DeviceExpiration.Value = t.Unix()
		} else if user.DeviceExpiration.Mode == models.UserDeviceExpirationDaily {
			t, err := time.ParseInLocation("15:04", devExpiration, time.UTC)
			if err != nil {
				u.e.Log.Errorf("Error saving user: %s", err.Error())
				// TODO: Return error
				return
			}
			user.DeviceExpiration.Value = t.Unix()
		} else if user.DeviceExpiration.Mode == models.UserDeviceExpirationDuration {
			d, err := time.ParseDuration(devExpiration)
			if err != nil {
				u.e.Log.Errorf("Error saving user: %s", err.Error())
				// TODO: Return error
				return
			}
			// time.Duration's Second() returns a float that contains the nanoseconds as well
			// For sanity we don't care and don't want the nanoseconds
			user.DeviceExpiration.Value = int64(d / time.Second)
		} else {
			u.e.Log.Error("Error saving user: Invalid device expiration type")
			// TODO: Return error
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
		if validStart != "" {
			t, err := time.Parse(common.TimeFormat, validStart)
			if err != nil {
				u.e.Log.Errorf("Error saving user: %s", err.Error())
				// TODO: Return error
				return
			}
			user.ValidStart = t
		}
		if validEnd != "" {
			t, err := time.Parse(common.TimeFormat, validEnd)
			if err != nil {
				u.e.Log.Errorf("Error saving user: %s", err.Error())
				// TODO: Return error
				return
			}
			user.ValidEnd = t
		}
		if user.ValidEnd.Before(user.ValidStart) {
			u.e.Log.Error("Error saving user: Valid end date before start date")
			// TODO: return error, end can't be before start
			return
		}
	}

	canManage := r.FormValue("can_manage")
	if canManage != "" {
		user.CanManage = (canManage == "1")
	}

	if err := user.Save(); err != nil {
		u.e.Log.Errorf("Error saving user: %s", err.Error())
		// TODO: Return error
		return
	}
	common.NewAPIOK("User created", nil).WriteTo(w)
}

func (u *User) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	user, err := models.GetUserByUsername(u.e, username)
	if err != nil {
		u.e.Log.Errorf("Error getting user for deletion: %s", err.Error())
		common.NewAPIResponse(common.APIStatusGenericError, "Error deleting user", nil).WriteTo(w)
		return
	}

	if err := user.Delete(); err != nil {
		u.e.Log.Errorf("Error deleting user: %s", err.Error())
		common.NewAPIResponse(common.APIStatusGenericError, "Error deleting user", nil).WriteTo(w)
		return
	}
	u.e.Log.Infof("Deleted user %s", username)
	common.NewAPIOK("User deleted", nil).WriteTo(w)
}
