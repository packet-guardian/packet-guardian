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

	UserDeviceLimitGlobal    UserDeviceLimit = -1
	UserDeviceLimitUnlimited UserDeviceLimit = 0
)

// User it's a user
type User struct {
	e                    *common.Environment
	ID                   int
	Username             string
	Password             string
	HasPassword          bool
	ClearPassword        bool
	DeviceLimit          UserDeviceLimit
	DeviceExpirationType UserExpiration
	DeviceExpiration     time.Time
	ValidStart           time.Time
	ValidEnd             time.Time
	ValidForever         bool
	CanManage            bool
	blacklistCached      bool
	blacklistSave        bool
	blacklisted          bool

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
		e:                    e,
		DeviceLimit:          UserDeviceLimitGlobal,
		DeviceExpirationType: UserDeviceExpirationGlobal,
		ValidForever:         true,
		CanManage:            true,
		isAdmin:              -1,
		isHelpDesk:           -1,
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
	sql := "SELECT \"id\", \"username\", \"password\", \"device_limit\", \"default_expiration\", \"expiration_type\", \"can_manage\", \"valid_forever\", \"valid_start\", \"valid_end\" FROM \"user\" " + where
	rows, err := e.DB.Query(sql, values...)
	if err != nil {
		return nil, err
	}

	var results []*User
	for rows.Next() {
		var id int
		var username string
		var password string
		var deviceLimit UserDeviceLimit
		var defaultExpiration int64
		var expirationType UserExpiration
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
			e:                    e,
			ID:                   id,
			Username:             username,
			HasPassword:          (password != ""),
			DeviceLimit:          deviceLimit,
			DeviceExpirationType: expirationType,
			DeviceExpiration:     time.Unix(defaultExpiration, 0),
			ValidStart:           time.Unix(validStart, 0),
			ValidEnd:             time.Unix(validEnd, 0),
			ValidForever:         validForever,
			CanManage:            canManage,
			isAdmin:              -1,
			isHelpDesk:           -1,
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

// NewPassword will hash s and set it as the password for User u.
func (u *User) NewPassword(s string) error {
	pass, err := bcrypt.GenerateFromPassword([]byte(s), 0)
	if err != nil {
		return err
	}
	u.Password = string(pass)
	u.HasPassword = true
	return nil
}

func (u *User) RemovePassword() {
	u.Password = ""
	u.HasPassword = false
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

	sql := "SELECT \"id\" FROM \"blacklist\" WHERE \"value\" = ?"
	var id int
	row := u.e.DB.QueryRow(sql, u.Username)
	err := row.Scan(&id)
	u.blacklisted = (err == nil)
	u.blacklistCached = true
	return u.blacklisted
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
	sql := "UPDATE \"user\" SET \"device_limit\" = ?, \"default_expiration\" = ?, \"expiration_type\" = ?, \"can_manage\" = ?, \"valid_forever\" = ?, \"valid_start\" = ?, \"valid_end\" = ?"

	if u.HasPassword && u.Password != "" {
		sql += ", \"password = ?"
	}

	sql += " WHERE \"id\" = ?"

	var err error
	if u.HasPassword && u.Password != "" {
		_, err = u.e.DB.Exec(
			sql,
			u.DeviceLimit,
			u.DeviceExpiration,
			u.DeviceExpirationType,
			u.CanManage,
			u.ValidForever,
			u.ValidStart,
			u.ValidEnd,
			u.Password,
		)
	} else {
		_, err = u.e.DB.Exec(
			sql,
			u.DeviceLimit,
			u.DeviceExpiration,
			u.DeviceExpirationType,
			u.CanManage,
			u.ValidForever,
			u.ValidStart,
			u.ValidEnd,
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

	sql := "INSERT INTO \"user\" (\"username\", \"password\", \"device_limit\", \"default_expiration\", \"expiration_type\", \"can_manage\", \"valid_forever\", \"valid_start\", \"valid_end\") VALUES (?,?,?,?,?,?,?,?)"

	_, err := u.e.DB.Exec(
		sql,
		u.Username,
		u.Password,
		u.DeviceLimit,
		u.DeviceExpiration,
		u.DeviceExpirationType,
		u.CanManage,
		u.ValidForever,
		u.ValidStart,
		u.ValidEnd,
		u.Password,
	)
	if err != nil {
		return err
	}
	return u.writeToBlacklist()
}

func (u *User) writeToBlacklist() error {
	// We only need to do something if the blacklist setting was changed
	if !u.blacklistSave {
		return nil
	}

	// If blacklisted, insert into database
	if u.blacklisted {
		sql := "INSERT INTO \"blacklist\" (\"value\") VALUES (?)"
		_, err := u.e.DB.Exec(sql, u.Username)
		return err
	}

	// Otherwise remove them from the blacklist
	sql := "DELETE FROM \"blacklist\" WHERE \"value\" = ?"
	_, err := u.e.DB.Exec(sql, u.Username)
	return err
}

func (u *User) Delete() error {
	sql := "DELETE FROM \"user\" WHERE \"id\" = ?"
	_, err := u.e.DB.Exec(sql, u.ID)
	return err
}
