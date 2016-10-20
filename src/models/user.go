// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"container/list"
	"errors"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/usi-lfkeitel/packet-guardian/src/common"
)

var appUserPool *userPool

func init() {
	appUserPool = newUserPool()
}

// When changing the User struct, make sure to update the
// clean() method to reflect any new/editied fields.

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
	blacklistCached  bool
	blacklistSave    bool
	blacklisted      bool
	Rights           Permission
	pool             *userPool
}

// NewUser creates a new base user
func NewUser(e *common.Environment) *User {
	// User with the following attributes:
	// Device limit is global
	// Device Expiration is global
	// User never expires
	// User can manage their devices
	u := appUserPool.getUser()
	u.e = e
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

	sql := "WHERE \"username\" = ?"
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
	sql := "WHERE \"" + field + "\" LIKE ?"
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

		user := appUserPool.getUser()
		user.e = e
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
		return
	}
	if common.StringInSlice(u.Username, u.e.Config.Auth.HelpDeskUsers) {
		u.Rights = u.Rights.With(HelpDeskRights)
		return
	}
	if common.StringInSlice(u.Username, u.e.Config.Auth.ReadOnlyUsers) {
		u.Rights = u.Rights.With(ReadOnlyRights)
		return
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
	if u.blacklistCached {
		return u.blacklisted
	}

	sql := `SELECT "id" FROM "blacklist" WHERE "value" = ?`
	var id int
	row := u.e.DB.QueryRow(sql, u.Username)
	err := row.Scan(&id)
	u.blacklisted = (err == nil)
	u.blacklistCached = true
	return u.blacklisted
}

func (u *User) IsExpired() bool {
	if u.ValidForever {
		return false
	}

	now := time.Now()
	return (u.ValidStart.After(now) || u.ValidEnd.Before(now))
}

func (u *User) Blacklist() {
	u.blacklistCached = true
	u.blacklisted = true
	u.blacklistSave = true
}

func (u *User) Unblacklist() {
	u.blacklistCached = true
	u.blacklisted = false
	u.blacklistSave = true
}

func (u *User) SaveToBlacklist() error {
	return u.writeToBlacklist()
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
	return u.writeToBlacklist()
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
	return u.writeToBlacklist()
}

func (u *User) writeToBlacklist() error {
	// We only need to do something if the blacklist setting was changed
	if !u.blacklistSave {
		return nil
	}

	// If blacklisted, insert into database
	if u.blacklisted {
		sql := `INSERT INTO "blacklist" ("value") VALUES (?)`
		_, err := u.e.DB.Exec(sql, u.Username)
		return err
	}

	// Otherwise remove them from the blacklist
	sql := `DELETE FROM "blacklist" WHERE "value" = ?`
	_, err := u.e.DB.Exec(sql, u.Username)
	return err
}

func (u *User) Delete() error {
	if u.ID == 0 {
		return nil
	}

	sql := `DELETE FROM "user" WHERE "id" = ?`
	_, err := u.e.DB.Exec(sql, u.ID)
	return err
}

// clean sets the User object to all field defaults
func (u *User) clean() {
	u.ID = 0
	u.Username = ""
	u.Password = ""
	u.HasPassword = false
	u.savePassword = false
	u.ClearPassword = false
	u.DeviceLimit = UserDeviceLimitUnlimited
	u.DeviceExpiration = nil
	u.ValidStart = time.Unix(0, 0)
	u.ValidEnd = time.Unix(0, 0)
	u.ValidForever = false
	u.CanManage = false
	u.CanAutoreg = false
	u.blacklistCached = false
	u.blacklistSave = false
	u.blacklisted = false
	u.Rights = Permission(0)
}

func (u *User) Release() {
	u.pool.release(u)
}

func ReleaseUsers(u []*User) {
	for _, user := range u {
		user.Release()
	}
}

// A userPool is a collection of user objects than can be used instead of
// creating new objects.
type userPool struct {
	*sync.RWMutex
	l *list.List
}

func newUserPool() *userPool {
	return &userPool{
		RWMutex: &sync.RWMutex{},
		l:       list.New(),
	}
}

func (p *userPool) getUser() *User {
	p.Lock()
	defer p.Unlock()

	e := p.l.Front()
	if e == nil { // Nothing in the list
		return &User{pool: p}
	}
	p.l.Remove(e)
	return e.Value.(*User)
}

func (p *userPool) release(u *User) {
	u.clean()
	p.Lock()
	p.l.PushBack(u)
	p.Unlock()
}
