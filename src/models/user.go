package models

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

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
	}
}

func GetUserByUsername(e *common.Environment, username string) (*User, error) {
	sql := "WHERE \"username\" = ?"
	users, err := getUsersFromDatabase(e, sql, username)
	if err != nil {
		return NewUser(e), err
	}
	return users[0], nil
}

func GetUserByID(e *common.Environment, id int) (*User, error) {
	sql := "WHERE \"id\" = ?"
	users, err := getUsersFromDatabase(e, sql, id)
	if err != nil {
		return NewUser(e), err
	}
	return users[0], nil
}

func GetAllUsers(e *common.Environment) ([]*User, error) {
	return getUsersFromDatabase(e, "")
}

func getUsersFromDatabase(e *common.Environment, where string, values ...interface{}) ([]*User, error) {
	sql := "SELECT \"id\", \"username\", \"password\", \"device_limit\", \"default_expiration\", \"expiration_type\", \"can_manage\", \"valid_forever\", \"valid_start\", \"valid_end\" FROM \"user\" " + where
	rows, err := e.DB.Query(sql, values)
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
		}
		results = append(results, user)
	}
	return results, nil
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
	return err
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
	return err
}
