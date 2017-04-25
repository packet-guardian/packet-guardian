// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/packet-guardian/packet-guardian/src/common"
)

type UserStore interface {
	Save(*User) error
	Delete(*User) error
	GetPassword(string) (string, error)
}

// User it's a user
type User struct {
	e                *common.Environment
	store            UserStore
	ID               int
	Username         string
	Password         string
	HasPassword      bool
	savePassword     bool
	ClearPassword    bool
	DeviceLimit      UserDeviceLimit
	DeviceExpiration *UserDeviceExpiration
	ValidStart       time.Time
	ValidEnd         time.Time
	ValidForever     bool
	CanManage        bool
	CanAutoreg       bool
	blacklist        BlacklistItem
	Rights           Permission
}

// NewUser creates a new base user
func NewUser(e *common.Environment, us UserStore, b BlacklistItem) *User {
	// User with the following attributes:
	// Device limit is global
	// Device Expiration is global
	// User never expires
	// User can manage their devices
	u := &User{
		e:                e,
		blacklist:        b,
		store:            us,
		DeviceLimit:      UserDeviceLimitGlobal,
		DeviceExpiration: &UserDeviceExpiration{Mode: UserDeviceExpirationGlobal},
		ValidStart:       time.Unix(0, 0),
		ValidEnd:         time.Unix(0, 0),
		ValidForever:     true,
		CanManage:        true,
		CanAutoreg:       true,
		Rights:           ViewOwn | ManageOwnRights,
	}
	// Load extra rights as set in the configuration
	u.LoadRights()
	return u
}

func (u *User) LoadRights() {
	if common.StringInSlice(u.Username, u.e.Config.Auth.AdminUsers) {
		u.Rights = u.Rights.With(AdminRights)
		u.Rights = u.Rights.Without(APIRead)
		u.Rights = u.Rights.Without(APIWrite)
	}
	if common.StringInSlice(u.Username, u.e.Config.Auth.HelpDeskUsers) {
		u.Rights = u.Rights.With(HelpDeskRights)
	}
	if common.StringInSlice(u.Username, u.e.Config.Auth.ReadOnlyUsers) {
		u.Rights = u.Rights.With(ReadOnlyRights)
	}
	if common.StringInSlice(u.Username, u.e.Config.Auth.APIReadOnlyUsers) {
		u.Rights = u.Rights.With(APIRead)
	}
	if common.StringInSlice(u.Username, u.e.Config.Auth.APIReadWriteUsers) {
		u.Rights = u.Rights.With(APIRead)
		u.Rights = u.Rights.With(APIWrite)
	}
}

func (u *User) IsNew() bool {
	return (u.ID == 0 || u.Username == "")
}

func (u *User) Can(p Permission) bool {
	return u.Rights.Can(p)
}

func (u *User) CanEither(p Permission) bool {
	return u.Rights.CanEither(p)
}

func (u *User) NeedToSavePassword() bool {
	return u.savePassword
}

func (u *User) GetPassword() string {
	if u.Password == "" {
		u.Password, _ = u.store.GetPassword(u.Username)
	}
	return u.Password
}

// NewPassword will hash s and set it as the password for User u.
func (u *User) NewPassword(s string) error {
	pass, err := bcrypt.GenerateFromPassword([]byte(s), 0)
	if err != nil {
		return err
	}
	u.Password = string(pass)
	u.HasPassword = true
	u.savePassword = true
	return nil
}

func (u *User) RemovePassword() {
	u.Password = ""
	u.HasPassword = false
	u.savePassword = true
}

func (u *User) IsBlacklisted() bool {
	return u.blacklist.IsBlacklisted(u.Username)
}

func (u *User) IsExpired() bool {
	if u.ValidForever {
		return false
	}

	now := time.Now()
	return (u.ValidStart.After(now) || u.ValidEnd.Before(now))
}

func (u *User) Blacklist() {
	u.blacklist.Blacklist()
}

func (u *User) Unblacklist() {
	u.blacklist.Unblacklist()
}

func (u *User) SaveToBlacklist() error {
	return u.blacklist.Save(u.Username)
}

func (u *User) Save() error {
	if err := u.store.Save(u); err != nil {
		return err
	}
	return u.SaveToBlacklist()
}

func (u *User) Delete() error {
	return u.store.Delete(u)
}
