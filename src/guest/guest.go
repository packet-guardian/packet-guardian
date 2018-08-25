// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package guest

import (
	"bytes"
	"errors"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/lfkeitel/verbose/v4"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	guestCodeChars  = "ABCDEFGHJKLMNPQRTUVWXYZ0123456789"
	guestCodeLength = 6
)

// GenerateGuestCode will create a 6 character verification code for guest registrations.
// Possibly confusing letters have been removed. In particular, the letters I, S, and O.
func GenerateGuestCode() string {
	code := bytes.Buffer{}
	for i := 0; i < guestCodeLength; i++ {
		code.WriteByte(guestCodeChars[rand.Intn(len(guestCodeChars))])
	}
	return code.String()
}

// RegisterDevice will register the device for a guest. It is a simplified form of the
// full registration function found in controllers.api.Device.RegistrationHandler().
func RegisterDevice(e *common.Environment, name, credential string, r *http.Request) error {
	// Build guest user model
	guest, err := stores.GetUserStore(e).GetUserByUsername(credential)
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"error":    err,
			"package":  "guest",
			"username": credential,
		}).Error("Error getting guest")
		return err
	}
	guest.DeviceLimit = models.UserDeviceLimit(e.Config.Guest.DeviceLimit)
	guest.DeviceExpiration = &models.UserDeviceExpiration{}

	guest.DeviceExpiration.Mode, guest.DeviceExpiration.Value, err = calcDeviceExpirationModeValue(e.Config.Guest.DeviceExpirationType, e.Config.Guest.DeviceExpiration)
	e.Log.WithFields(verbose.Fields{
		"error":   err,
		"package": "guest",
	}).Error("Error parsing device expiration")

	// Get and enforce the device limit
	deviceCount, err := stores.GetDeviceStore(e).GetDeviceCountForUser(guest)
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"package": "guest",
			"error":   err,
		}).Error("Error getting device count")
	}
	if guest.DeviceLimit != models.UserDeviceLimitUnlimited &&
		deviceCount >= int(guest.DeviceLimit) {
		return errors.New("Device limit reached")
	}

	// Get MAC address
	var mac net.HardwareAddr
	ip := common.GetIPFromContext(r)

	// Automatic registration
	lease, err := stores.GetLeaseStore(e).GetLeaseByIP(ip)
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "guest",
			"ip":      ip.String(),
		}).Error("Error getting MAC for IP")
		return errors.New("Internal Server Error")
	} else if lease.ID == 0 {
		e.Log.WithFields(verbose.Fields{
			"package": "guest",
			"ip":      ip.String(),
		}).Notice("Attempted auto reg from non-leased device")
		return errors.New("Error detecting MAC address")
	}
	mac = lease.MAC

	// Get device from database
	device, err := stores.GetDeviceStore(e).GetDeviceByMAC(mac)
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "guest",
			"mac":     mac.String(),
		}).Error("Error getting device")
		return errors.New("Database error")
	}

	// Check if device is already registered
	if device.ID != 0 {
		e.Log.WithFields(verbose.Fields{
			"package":  "guest",
			"mac":      mac.String(),
			"username": credential,
		}).Notice("Attempted duplicate registration")
		return errors.New("This device is already registered")
	}

	// Validate platform, we don't want someone to submit an inappropriate value
	platform := common.ParseUserAgent(r.UserAgent())

	// Fill in device information
	device.Username = credential
	device.Description = "Guest - " + name
	device.RegisteredFrom = ip
	device.Platform = platform
	device.Expires = guest.DeviceExpiration.NextExpiration(e, time.Now())
	device.DateRegistered = time.Now()
	device.LastSeen = time.Now()
	device.UserAgent = r.UserAgent()

	// Save new device
	if err := device.Save(); err != nil {
		e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "guest",
		}).Error("Error saving device")
		return errors.New("Error registering device")
	}
	e.Log.WithFields(verbose.Fields{
		"package":  "guest",
		"mac":      mac.String(),
		"name":     name,
		"username": credential,
		"action":   "register_guest_device",
	}).Info("Device registered")
	return nil
}

// TODO: Create tests for this
func calcDeviceExpirationModeValue(expType, expTimeStr string) (models.UserExpiration, int64, error) {
	switch expType {
	case "never":
		return models.UserDeviceExpirationNever, 0, nil
	case "date":
		expTime, err := time.ParseInLocation(common.TimeFormat, expTimeStr, time.Local)
		return models.UserDeviceExpirationSpecific, expTime.Unix(), err
	case "duration":
		dur, err := time.ParseDuration(expTimeStr)
		return models.UserDeviceExpirationDuration, int64(dur / time.Second), err
	case "daily":
		expTime, err := common.ParseTime(expTimeStr)
		return models.UserDeviceExpirationDaily, expTime, err
	default:
		return 0, 0, errors.New(expType + " is not a valid device expiration type")
	}
}
