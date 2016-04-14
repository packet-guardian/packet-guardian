package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/onesimus-systems/packet-guardian/common"

	"golang.org/x/crypto/bcrypt"
)

type authFunc func(db *sql.DB, username, password string) bool

var authFunctions = make(map[string]authFunc)

func init() {
	authFunctions["local"] = normalAuth
	authFunctions["ldap"] = ldapAuth
}

// IsValidLogin will verify the username and password against several login methods
// If one method succeeds, true will be returned. False otherwise.
func IsValidLogin(db *sql.DB, username, password string) bool {
	// Check the user and pass against the defined auth functions
	// For right now we're only doing local authentication
	return authFunctions["local"](db, username, password)
}

// IsLoggedIn checks if a user is logged in
func IsLoggedIn(e *common.Environment, r *http.Request) bool {
	sess := e.Sessions.GetSession(r, e.Config.Webserver.SessionName)
	return sess.GetBool("loggedin", false)
}

// IsAdminUser checks if a user is an administrator
func IsAdminUser(e *common.Environment, r *http.Request) bool {
	username := e.Sessions.GetSession(r, e.Config.Webserver.SessionName).GetString("username", "")
	return common.StringInSlice(username, e.Config.Auth.AdminUsers)
}

// LogoutUser will set loggedin to false and delete the session
func LogoutUser(e *common.Environment, w http.ResponseWriter, r *http.Request) {
	sess := e.Sessions.GetSession(r, e.Config.Webserver.SessionName)
	sess.Set("loggedin", false)
	sess.Delete(r, w)
}

func normalAuth(db *sql.DB, username, password string) bool {
	if password == "" || username == "" {
		return false
	}

	stmt, err := db.Prepare("SELECT \"password\" FROM \"user\" WHERE \"username\" = ?")
	if err != nil {
		return false
	}
	user := stmt.QueryRow(username)

	var storedPass string
	err = user.Scan(&storedPass)
	if err != nil {
		return false
	}

	if storedPass == "" {
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPass), []byte(password))
	return (err == nil)
}

func ldapAuth(db *sql.DB, username, password string) bool {
	// Check username and pass against an ldap server
	return false
}

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
func GetUser(db *sql.DB, username string) (*User, error) {
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
func GetAllUsers(db *sql.DB) ([]*User, error) {
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
func (u *User) Save(db *sql.DB) error {
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

func (u *User) saveNew(db *sql.DB) error {
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
