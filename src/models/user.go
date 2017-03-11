// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/usi-lfkeitel/packet-guardian/src/common"
)

// User it's a user
type User struct {
	e                *common.Environment
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
	blacklist        *blacklistItem
	Rights           Permission
}

// NewUser creates a new base user
func NewUser(e *common.Environment) *User {
	// User with the following attributes:
	// Device limit is global
	// Device Expiration is global
	// User never expires
	// User can manage their devices
	u := &User{}
	u.e = e
	u.blacklist = newBlacklistItem(getBlacklistStore(e))
	u.DeviceLimit = UserDeviceLimitGlobal
	u.DeviceExpiration = &UserDeviceExpiration{Mode: UserDeviceExpirationGlobal}
	u.ValidStart = time.Unix(0, 0)
	u.ValidEnd = time.Unix(0, 0)
	u.ValidForever = true
	u.CanManage = true
	u.CanAutoreg = true
	u.Rights = ViewOwn | ManageOwnRights
	// Load extra rights as set in the configuration
	u.LoadRights()
	return u
}

func GetUserByUsername(e *common.Environment, username string) (*User, error) {
	if username == "" {
		return NewUser(e), nil
	}

	username = strings.ToLower(username)

	sql := `WHERE "username" = ?`
	users, err := getUsersFromDatabase(e, sql, username)
	if users == nil || len(users) == 0 {
		u := NewUser(e)
		u.Username = username
		u.LoadRights()
		return u, err
	}
	users[0].LoadRights()
	return users[0], nil
}

func GetAllUsers(e *common.Environment) ([]*User, error) {
	sql := `ORDER BY "username"`
	if e.DB.Driver == "sqlite" {
		sql += " COLLATE NOCASE"
	}
	sql += " ASC"
	return getUsersFromDatabase(e, sql)
}

func SearchUsersByField(e *common.Environment, field, pattern string) ([]*User, error) {
	sql := `WHERE "` + field + `" LIKE ?`
	return getUsersFromDatabase(e, sql, pattern)
}

func getUsersFromDatabase(e *common.Environment, where string, values ...interface{}) ([]*User, error) {
	sql := `SELECT "id", "username", "password", "device_limit", "default_expiration", "expiration_type", "can_manage", "can_autoreg", "valid_forever", "valid_start", "valid_end" FROM "user" ` + where
	rows, err := e.DB.Query(sql, values...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*User
	for rows.Next() {
		var id int
		var username string
		var password string
		var deviceLimit int
		var defaultExpiration int64
		var expirationType int
		var canManage bool
		var canAutoreg bool
		var validForever bool
		var validStart int64
		var validEnd int64

		err := rows.Scan(
			&id,
			&username,
			&password,
			&deviceLimit,
			&defaultExpiration,
			&expirationType,
			&canManage,
			&canAutoreg,
			&validForever,
			&validStart,
			&validEnd,
		)
		if err != nil {
			continue
		}

		user := NewUser(e)
		user.e = e
		user.blacklist = newBlacklistItem(getBlacklistStore(e))
		user.ID = id
		user.Username = username
		user.HasPassword = (password != "")
		user.DeviceLimit = UserDeviceLimit(deviceLimit)
		user.ValidStart = time.Unix(validStart, 0)
		user.ValidEnd = time.Unix(validEnd, 0)
		user.ValidForever = validForever
		user.CanManage = canManage
		user.CanAutoreg = canAutoreg
		user.Rights = ViewOwn

		if canManage {
			user.Rights = user.Rights.With(ManageOwnRights)
		}
		if canAutoreg {
			user.Rights = user.Rights.With(AutoRegOwn)
		}
		user.DeviceExpiration = &UserDeviceExpiration{
			Mode:  UserExpiration(expirationType),
			Value: defaultExpiration,
		}
		results = append(results, user)
	}
	return results, nil
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

func (u *User) GetPassword() string {
	result := u.e.DB.QueryRow(`SELECT "password" FROM "user" WHERE "username" = ?`, u.Username)
	err := result.Scan(&u.Password)
	if err != nil {
		return ""
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
	return u.blacklist.isBlacklisted(u.Username)
}

func (u *User) IsExpired() bool {
	if u.ValidForever {
		return false
	}

	now := time.Now()
	return (u.ValidStart.After(now) || u.ValidEnd.Before(now))
}

func (u *User) Blacklist() {
	u.blacklist.blacklist()
}

func (u *User) Unblacklist() {
	u.blacklist.unblacklist()
}

func (u *User) SaveToBlacklist() error {
	return u.blacklist.save(u.Username)
}

func (u *User) Save() error {
	if u.ID == 0 {
		return u.saveNew()
	}
	return u.updateExisting()
}

func (u *User) updateExisting() error {
	sql := `UPDATE "user" SET "device_limit" = ?, "default_expiration" = ?, "expiration_type" = ?, "can_manage" = ?, "can_autoreg" = ?, "valid_forever" = ?, "valid_start" = ?, "valid_end" = ?`

	if u.savePassword {
		sql += ", \"password\" = ?"
	}

	sql += " WHERE \"id\" = ?"

	var err error
	if u.savePassword {
		_, err = u.e.DB.Exec(
			sql,
			u.DeviceLimit,
			u.DeviceExpiration.Value,
			u.DeviceExpiration.Mode,
			u.CanManage,
			u.CanAutoreg,
			u.ValidForever,
			u.ValidStart.Unix(),
			u.ValidEnd.Unix(),
			u.Password,
			u.ID,
		)
	} else {
		_, err = u.e.DB.Exec(
			sql,
			u.DeviceLimit,
			u.DeviceExpiration.Value,
			u.DeviceExpiration.Mode,
			u.CanManage,
			u.CanAutoreg,
			u.ValidForever,
			u.ValidStart.Unix(),
			u.ValidEnd.Unix(),
			u.ID,
		)
	}
	if err != nil {
		return err
	}
	return u.blacklist.save(u.Username)
}

func (u *User) saveNew() error {
	if u.Username == "" {
		return errors.New("Username cannot be empty")
	}

	sql := `INSERT INTO "user" ("username", "password", "device_limit", "default_expiration", "expiration_type", "can_manage", "can_autoreg", "valid_forever", "valid_start", "valid_end") VALUES (?,?,?,?,?,?,?,?,?,?)`

	result, err := u.e.DB.Exec(
		sql,
		u.Username,
		u.Password,
		u.DeviceLimit,
		u.DeviceExpiration.Value,
		u.DeviceExpiration.Mode,
		u.CanManage,
		u.CanAutoreg,
		u.ValidForever,
		u.ValidStart.Unix(),
		u.ValidEnd.Unix(),
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	u.ID = int(id)
	return u.blacklist.save(u.Username)
}

func (u *User) Delete() error {
	if u.ID == 0 {
		return nil
	}

	sql := `DELETE FROM "user" WHERE "id" = ?`
	_, err := u.e.DB.Exec(sql, u.ID)
	return err
}
