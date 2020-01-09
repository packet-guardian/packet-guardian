// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"encoding/json"
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
	ID               int    `json:"-"`
	Username         string `json:"username"`
	Password         string `json:"-"`
	HasPassword      bool   `json:"has_password"`
	savePassword     bool
	ClearPassword    bool                  `json:"-"`
	DeviceLimit      UserDeviceLimit       `json:"device_limit"`
	DeviceExpiration *UserDeviceExpiration `json:"device_expiration"`
	ValidStart       time.Time             `json:"-"`
	ValidEnd         time.Time             `json:"-"`
	ValidForever     bool                  `json:"valid_forever"`
	CanManage        bool                  `json:"can_manage"`
	CanAutoreg       bool                  `json:"can_autoreg"`
	blacklist        BlacklistItem
	Rights           Permission `json:"-"`

	UIGroup        string                `json:"-"`
	APIGroup       string                `json:"-"`
	AllowStatusAPI bool                  `json:"-"`
	Delegates      map[string]Permission `json:"delegates"`
}

// NewUser creates a new base user
func NewUser(e *common.Environment, us UserStore, b BlacklistItem, username string) *User {
	// User with the following attributes:
	// Device limit is global
	// Device Expiration is global
	// User never expires
	// User can manage their devices
	u := &User{
		e:                e,
		blacklist:        b,
		store:            us,
		Username:         username,
		DeviceLimit:      UserDeviceLimitGlobal,
		DeviceExpiration: &UserDeviceExpiration{Mode: UserDeviceExpirationGlobal},
		ValidStart:       time.Unix(0, 0),
		ValidEnd:         time.Unix(0, 0),
		ValidForever:     true,
		CanManage:        true,
		CanAutoreg:       true,
		Rights:           ViewOwn | ManageOwnRights,
		UIGroup:          "default",
		APIGroup:         "disabled",
		Delegates:        make(map[string]Permission),
	}
	// Load extra rights as set in the configuration
	u.LoadRights()
	return u
}

func (u *User) LoadRights() {
	u.Rights = u.Rights.With(uiPermissions[u.UIGroup])
	u.Rights = u.Rights.With(apiPermissions[u.APIGroup])
	if u.AllowStatusAPI {
		u.Rights = u.Rights.With(apiPermissions["status-api"])
	}

	if u.IsBlacklisted() {
		u.Rights = u.Rights.Without(ManageOwnRights)
	}
}

func (u *User) MarshalJSON() ([]byte, error) {
	type Alias User
	return json.Marshal(&struct {
		*Alias
		ValidStart  time.Time `json:"valid_start"`
		ValidEnd    time.Time `json:"valid_end"`
		Blacklisted bool      `json:"blacklisted"`
	}{
		Alias:       (*Alias)(u),
		ValidStart:  u.ValidStart.UTC(),
		ValidEnd:    u.ValidEnd.UTC(),
		Blacklisted: u.IsBlacklisted(),
	})
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

func (u *User) DelegateCan(username string, p Permission) bool {
	return u.Delegates[username].Can(p)
}

func (u *User) DelegateCanEither(username string, p Permission) bool {
	return u.Delegates[username].CanEither(p)
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
