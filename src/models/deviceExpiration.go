// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
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

var globalDeviceExpiration *UserDeviceExpiration

type UserDeviceExpiration struct {
	Mode  UserExpiration
	Value int64 // Daily and Duration, time in seconds. Specific, unix epoch
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
	if e.Mode == UserDeviceExpirationNever {
		return "Never"
	} else if e.Mode == UserDeviceExpirationGlobal {
		return "Global"
	} else if e.Mode == UserDeviceExpirationRolling {
		return "Rolling"
	} else if e.Mode == UserDeviceExpirationSpecific {
		return time.Unix(e.Value, 0).Format(common.TimeFormat)
	} else if e.Mode == UserDeviceExpirationDuration {
		return (time.Duration(e.Value) * time.Second).String()
	}

	year, month, day := time.Now().Date()
	bod := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	bod = bod.Add(time.Duration(e.Value) * time.Second)
	return "Daily at " + bod.Format("15:04")
}

func (e *UserDeviceExpiration) NextExpiration(env *common.Environment, base time.Time) time.Time {
	if e.Mode == UserDeviceExpirationGlobal {
		if globalDeviceExpiration != nil {
			return globalDeviceExpiration.NextExpiration(env, base)
		}
		// Build the global default, typically it's the most used mode
		globalDeviceExpiration = GetGlobalDefaultExpiration(env)
		return globalDeviceExpiration.NextExpiration(env, base)
	} else if e.Mode == UserDeviceExpirationSpecific {
		return time.Unix(e.Value, 0)
	} else if e.Mode == UserDeviceExpirationDuration {
		return base.Add(time.Duration(e.Value) * time.Second)
	} else if e.Mode == UserDeviceExpirationDaily {
		year, month, day := base.Date()
		bod := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
		bod = bod.Add(time.Duration(e.Value) * time.Second)
		if bod.Before(base) { // If the time has passed today, rollover to tomorrow
			bod = bod.Add(time.Duration(24) * time.Hour)
		}
		return bod
	} else if e.Mode == UserDeviceExpirationRolling {
		return time.Unix(1, 0)
	}

	// Default to never
	return time.Unix(0, 0)
}
