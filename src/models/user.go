package models

import (
	"errors"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/gorilla/context"
	"github.com/onesimus-systems/packet-guardian/src/common"
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
	} else {
		year, month, day := time.Now().Date()
		bod := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
		bod = bod.Add(time.Duration(e.Value) * time.Second)
		return "Daily at " + bod.Format("15:04")
	}
}

func (e *UserDeviceExpiration) NextExpiration(env *common.Environment) time.Time {
	if e.Mode == UserDeviceExpirationGlobal {
		if globalDeviceExpiration != nil {
			return globalDeviceExpiration.NextExpiration(env)
		}
		// Build the global default, typically it's the most used mode
		globalDeviceExpiration = &UserDeviceExpiration{} // Defaults to never
		switch env.Config.Registration.DefaultDeviceExpirationType {
		case "date":
			d, err := time.ParseInLocation("2006-01-02", env.Config.Registration.DefaultDeviceExpiration, time.Local)
			if err != nil {
				env.Log.Error("Incorrect default device expiration date format")
				break
			}
			globalDeviceExpiration.Value = d.Unix()
		case "duration":
			d, err := time.ParseDuration(env.Config.Registration.DefaultDeviceExpiration)
			if err != nil {
				env.Log.Error("Incorrect default device expiration duration format")
				break
			}
			globalDeviceExpiration.Value = int64(d / time.Second)
		case "daily":
			secs, err := common.ParseTime(env.Config.Registration.DefaultDeviceExpiration)
			if err != nil {
				env.Log.Error("Incorrect default device expiration time format")
				break
			}
			globalDeviceExpiration.Value = secs
		}
		return globalDeviceExpiration.NextExpiration(env)
	} else if e.Mode == UserDeviceExpirationSpecific {
		return time.Unix(e.Value, 0)
	} else if e.Mode == UserDeviceExpirationDuration {
		return time.Now().Add(time.Duration(e.Value) * time.Second)
	} else if e.Mode == UserDeviceExpirationDaily {
		now := time.Now()
		year, month, day := now.Date()
		bod := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
		bod = bod.Add(time.Duration(e.Value) * time.Second)
		if bod.Before(now) { // If the time has passed today, rollover to tomorrow
			bod = bod.Add(time.Duration(24) * time.Hour)
		}
		return bod
	} else if e.Mode == UserDeviceExpirationRolling {
		return time.Unix(1, 0)
	} else { // Default to never
		return time.Unix(0, 0)
	}
}

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
	blacklistCached  bool
	blacklistSave    bool
	blacklisted      bool

	isAdmin    int
	isHelpDesk int
}

// NewUser creates a new base user
func NewUser(e *common.Environment) *User {
	// User with the following attributes:
	// Device limit is global
	// Device Expiration is global
	// User never expires
	// User can manage their devices
	return &User{
		e:                e,
		DeviceLimit:      UserDeviceLimitGlobal,
		DeviceExpiration: &UserDeviceExpiration{Mode: UserDeviceExpirationGlobal},
		ValidStart:       time.Unix(0, 0),
		ValidEnd:         time.Unix(0, 0),
		ValidForever:     true,
		CanManage:        true,
		isAdmin:          -1,
		isHelpDesk:       -1,
	}
}

func GetUserByUsername(e *common.Environment, username string) (*User, error) {
	if username == "" {
		return NewUser(e), nil
	}

	sql := "WHERE \"username\" = ?"
	users, err := getUsersFromDatabase(e, sql, username)
	if users == nil || len(users) == 0 {
		u := NewUser(e)
		u.Username = username
		return u, err
	}
	return users[0], nil
}

func GetUserByID(e *common.Environment, id int) (*User, error) {
	sql := "WHERE \"id\" = ?"
	users, err := getUsersFromDatabase(e, sql, id)
	if users == nil || len(users) == 0 {
		return NewUser(e), err
	}
	return users[0], nil
}

func GetAllUsers(e *common.Environment) ([]*User, error) {
	return getUsersFromDatabase(e, "")
}

func SearchUsersByField(e *common.Environment, field, pattern string) ([]*User, error) {
	sql := "WHERE \"" + field + "\" LIKE ?"
	return getUsersFromDatabase(e, sql, pattern)
}

func getUsersFromDatabase(e *common.Environment, where string, values ...interface{}) ([]*User, error) {
	sql := `SELECT "id", "username", "password", "device_limit", "default_expiration", "expiration_type", "can_manage", "valid_forever", "valid_start", "valid_end" FROM "user" ` + where
	rows, err := e.DB.Query(sql, values...)
	if err != nil {
		return nil, err
	}

	var results []*User
	for rows.Next() {
		var id int
		var username string
		var password string
		var deviceLimit int
		var defaultExpiration int64
		var expirationType int
		var canManage bool
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
			&validForever,
			&validStart,
			&validEnd,
		)
		if err != nil {
			continue
		}

		user := &User{
			e:            e,
			ID:           id,
			Username:     username,
			HasPassword:  (password != ""),
			DeviceLimit:  UserDeviceLimit(deviceLimit),
			ValidStart:   time.Unix(validStart, 0),
			ValidEnd:     time.Unix(validEnd, 0),
			ValidForever: validForever,
			CanManage:    canManage,
			isAdmin:      -1,
			isHelpDesk:   -1,
		}
		user.DeviceExpiration = &UserDeviceExpiration{
			Mode:  UserExpiration(expirationType),
			Value: defaultExpiration,
		}
		results = append(results, user)
	}
	return results, nil
}

func GetUserFromContext(r *http.Request) *User {
	if rv := context.Get(r, common.SessionUserKey); rv != nil {
		return rv.(*User)
	}
	return nil
}

func SetUserToContext(r *http.Request, u *User) {
	context.Set(r, common.SessionUserKey, u)
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

func (u *User) IsAdmin() bool {
	if u.isAdmin != -1 {
		return (u.isAdmin == 1)
	}
	u.isAdmin = 0
	if common.StringInSlice(u.Username, u.e.Config.Auth.AdminUsers) {
		u.isAdmin = 1
	}
	return (u.isAdmin == 1)
}

func (u *User) IsHelpDesk() bool {
	if u.isHelpDesk != -1 {
		return (u.isHelpDesk == 1)
	}
	u.isHelpDesk = 0
	if common.StringInSlice(u.Username, u.e.Config.Auth.HelpDeskUsers) {
		u.isHelpDesk = 1
	}
	return (u.isHelpDesk == 1)
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
	sql := `UPDATE "user" SET "device_limit" = ?, "default_expiration" = ?, "expiration_type" = ?, "can_manage" = ?, "valid_forever" = ?, "valid_start" = ?, "valid_end" = ?`

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

	sql := `INSERT INTO "user" ("username", "password", "device_limit", "default_expiration", "expiration_type", "can_manage", "valid_forever", "valid_start", "valid_end") VALUES (?,?,?,?,?,?,?,?,?)`

	result, err := u.e.DB.Exec(
		sql,
		u.Username,
		u.Password,
		u.DeviceLimit,
		u.DeviceExpiration.Value,
		u.DeviceExpiration.Mode,
		u.CanManage,
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
