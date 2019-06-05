// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"encoding/json"
	"time"

	"github.com/packet-guardian/packet-guardian/src/common"
)

type UserExpiration int
type UserDeviceLimit int

const (
	UserDeviceExpirationNever    UserExpiration = 0
	UserDeviceExpirationGlobal   UserExpiration = 1
	UserDeviceExpirationSpecific UserExpiration = 2
	UserDeviceExpirationDuration UserExpiration = 3
	UserDeviceExpirationDaily    UserExpiration = 4
	UserDeviceExpirationRolling  UserExpiration = 5

	UserDeviceLimitGlobal    UserDeviceLimit = -1
	UserDeviceLimitUnlimited UserDeviceLimit = 0
)

func (ue UserExpiration) string() string {
	switch ue {
	case UserDeviceExpirationNever:
		return "never"
	case UserDeviceExpirationGlobal:
		return "global"
	case UserDeviceExpirationSpecific:
		return "specific-datetime"
	case UserDeviceExpirationDuration:
		return "duration"
	case UserDeviceExpirationDaily:
		return "daily-time"
	case UserDeviceExpirationRolling:
		return "rolling-duration"
	}
	return ""
}

var globalDeviceExpiration *UserDeviceExpiration

type UserDeviceExpiration struct {
	Mode  UserExpiration
	Value int64 // Daily and Duration, time in seconds. Specific, unix epoch
}

func (ude *UserDeviceExpiration) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Mode  string `json:"mode"`
		Value int64  `json:"value"`
	}{
		Mode:  ude.Mode.string(),
		Value: ude.Value,
	})
}

func GetGlobalDefaultExpiration(e *common.Environment) *UserDeviceExpiration {
	g := &UserDeviceExpiration{} // Defaults to never
	switch e.Config.Registration.DefaultDeviceExpirationType {
	case "date":
		d, err := time.ParseInLocation("2006-01-02", e.Config.Registration.DefaultDeviceExpiration, time.Local)
		if err != nil {
			e.Log.Error("Incorrect default device expiration date format")
			break
		}
		g.Value = d.Unix()
		g.Mode = UserDeviceExpirationSpecific
	case "duration":
		d, err := time.ParseDuration(e.Config.Registration.DefaultDeviceExpiration)
		if err != nil {
			e.Log.Error("Incorrect default device expiration duration format")
			break
		}
		g.Value = int64(d / time.Second)
		g.Mode = UserDeviceExpirationDuration
	case "daily":
		secs, err := common.ParseTime(e.Config.Registration.DefaultDeviceExpiration)
		if err != nil {
			e.Log.Error("Incorrect default device expiration time format")
			break
		}
		g.Value = secs
		g.Mode = UserDeviceExpirationDaily
	case "rolling":
		g.Value = 0
		g.Mode = UserDeviceExpirationRolling
	}
	return g
}

func (e *UserDeviceExpiration) String() string {
	switch e.Mode {
	case UserDeviceExpirationNever:
		return "Never"
	case UserDeviceExpirationGlobal:
		return "Global"
	case UserDeviceExpirationRolling:
		return "Rolling"
	case UserDeviceExpirationSpecific:
		return time.Unix(e.Value, 0).Format(common.TimeFormat)
	case UserDeviceExpirationDuration:
		return (time.Duration(e.Value) * time.Second).String()
	}

	// Daily
	year, month, day := time.Now().Date()
	bod := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	bod = bod.Add(time.Duration(e.Value) * time.Second)
	return "Daily at " + bod.Format("15:04")
}

func (e *UserDeviceExpiration) NextExpiration(env *common.Environment, base time.Time) time.Time {
	switch e.Mode {
	case UserDeviceExpirationGlobal:
		if globalDeviceExpiration != nil {
			return globalDeviceExpiration.NextExpiration(env, base)
		}
		// Build the global default, typically it's the most used mode
		globalDeviceExpiration = GetGlobalDefaultExpiration(env)
		return globalDeviceExpiration.NextExpiration(env, base)
	case UserDeviceExpirationRolling:
		return time.Unix(1, 0)
	case UserDeviceExpirationSpecific:
		return time.Unix(e.Value, 0)
	case UserDeviceExpirationDuration:
		return base.Add(time.Duration(e.Value) * time.Second)
	case UserDeviceExpirationDaily:
		year, month, day := base.Date()
		bod := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
		bod = bod.Add(time.Duration(e.Value) * time.Second)
		if bod.Before(base) { // If the time has passed today, rollover to tomorrow
			bod = bod.Add(time.Duration(24) * time.Hour)
		}
		return bod
	}

	// Default to never
	return time.Unix(0, 0)
}
