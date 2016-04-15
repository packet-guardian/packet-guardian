package models

import (
	"errors"
	"time"

	"github.com/onesimus-systems/packet-guardian/src/common"

	"golang.org/x/crypto/bcrypt"
)

// User it's a user
type User struct {
	ID                int
	Username          string
	Password          string
	ClearPassword     bool
	DeviceLimit       int
	RawExpiration     int64
	DefaultExpiration time.Time
	ValidAfter        time.Time
	ValidBefore       time.Time
	NeverExpires      bool
	UserNeverExpires  bool
	Existed           bool
}

// NewUser creates a new base user
func NewUser() *User {
	// User with the following attributes:
	// Device limit is global
	// Expiration is global
	// User never expires
	return &User{
		Existed:           false,
		DeviceLimit:       -1,
		RawExpiration:     1,
		DefaultExpiration: time.Unix(1, 0),
		ValidAfter:        time.Unix(0, 0),
		ValidBefore:       time.Unix(0, 0),
		UserNeverExpires:  true,
	}
}

// TODO: Consildate the below two functions into one

// GetUser by username
func GetUser(db *common.DatabaseAccessor, username string) (*User, error) {
	stmt, err := db.Prepare("SELECT \"id\", \"deviceLimit\", \"expires\", \"validAfter\", \"validBefore\" FROM \"user\" WHERE \"username\" = ?")
	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(username)
	var result *User
	var id int
	var deviceLimit int
	var expiration int64
	var validAfter int64
	var validBefore int64

	err = row.Scan(&id, &deviceLimit, &expiration, &validAfter, &validBefore)
	if err != nil {
		return nil, err
	}

	result = &User{
		ID:                id,
		Username:          username,
		DeviceLimit:       deviceLimit,
		RawExpiration:     expiration,
		DefaultExpiration: time.Unix(expiration, 0),
		ValidAfter:        time.Unix(validAfter, 0),
		ValidBefore:       time.Unix(validBefore, 0),
		NeverExpires:      (expiration == 0),
		UserNeverExpires:  (validAfter == 0 && validBefore == 0),
		Existed:           true,
	}
	return result, nil
}

// GetAllUsers will return a slice of Users on success. Returns nil and an error on error.
func GetAllUsers(db *common.DatabaseAccessor) ([]*User, error) {
	rows, err := db.Query("SELECT \"id\", \"username\", \"deviceLimit\", \"expires\", \"validAfter\", \"validBefore\" FROM \"user\"")
	if err != nil {
		return nil, err
	}

	var results []*User
	for rows.Next() {
		var id int
		var username string
		var deviceLimit int
		var expiration int64
		var validAfter int64
		var validBefore int64

		err := rows.Scan(&id, &username, &deviceLimit, &expiration, &validAfter, &validBefore)
		if err != nil {
			continue
		}

		user := &User{
			ID:                id,
			Username:          username,
			DeviceLimit:       deviceLimit,
			RawExpiration:     expiration,
			DefaultExpiration: time.Unix(expiration, 0),
			ValidAfter:        time.Unix(validAfter, 0),
			ValidBefore:       time.Unix(validBefore, 0),
			NeverExpires:      (expiration == 0),
			UserNeverExpires:  (validAfter == 0 && validBefore == 0),
			Existed:           true,
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
	return nil
}

// Save the user to a database
func (u *User) Save(db *common.DatabaseAccessor) error {
	if !u.Existed {
		return u.saveNew(db)
	}

	baseSQL := "UPDATE \"user\" SET \"username\" = ?, \"deviceLimit\" = ?, \"expires\" = ?, \"validAfter\" = ?, \"validBefore\" = ?"
	passParam := ""
	if u.ClearPassword {
		baseSQL += ", \"password\" = ?"
	} else if u.Password != "" {
		baseSQL += ", \"password\" = ?"
		passParam = u.Password
	}
	baseSQL += " WHERE \"id\" = ?"

	var err error
	if u.ClearPassword || u.Password != "" {
		_, err = db.Exec(
			baseSQL,
			u.Username,
			u.DeviceLimit,
			u.DefaultExpiration.Unix(),
			u.ValidAfter.Unix(),
			u.ValidBefore.Unix(),
			passParam,
			u.ID,
		)
	} else {
		_, err = db.Exec(
			baseSQL,
			u.Username,
			u.DeviceLimit,
			u.DefaultExpiration.Unix(),
			u.ValidAfter.Unix(),
			u.ValidBefore.Unix(),
			u.ID,
		)
	}
	return err
}

func (u *User) saveNew(db *common.DatabaseAccessor) error {
	if u.Username == "" {
		return errors.New("Username cannot be empty")
	}

	sql := "INSERT INTO \"user\" (\"username\", \"password\", \"deviceLimit\", \"expires\", \"validAfter\", \"validBefore\") "
	sql += "VALUES (?,?,?,?,?,?)"
	_, err := db.Exec(
		sql,
		u.Username,
		u.Password,
		u.DeviceLimit,
		u.DefaultExpiration.Unix(),
		u.ValidAfter.Unix(),
		u.ValidBefore.Unix(),
	)
	return err
}
