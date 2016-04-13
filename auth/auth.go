package auth

import (
	"database/sql"
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

	err = bcrypt.CompareHashAndPassword([]byte(storedPass), []byte(password))
	if err == nil {
		return true
	}
	return false
}

func ldapAuth(db *sql.DB, username, password string) bool {
	// Check username and pass against an ldap server
	return false
}

// User it's a user
type User struct {
	ID                int
	Username          string
	DeviceLimit       int
	DefaultExpiration time.Time
	ValidAfter        time.Time
	ValidBefore       time.Time
	NeverExpires      bool
	UserNeverExpires  bool
}

// GetAllUsers will return a slice of Users on success. Returns nil and an error on error.
func GetAllUsers(db *sql.DB) ([]User, error) {
	rows, err := db.Query("SELECT \"id\", \"username\", \"deviceLimit\", \"expires\", \"validAfter\", \"validBefore\" FROM \"user\"")
	if err != nil {
		return nil, err
	}

	var results []User
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

		user := User{
			ID:                id,
			Username:          username,
			DeviceLimit:       deviceLimit,
			DefaultExpiration: time.Unix(expiration, 0),
			ValidAfter:        time.Unix(validAfter, 0),
			ValidBefore:       time.Unix(validBefore, 0),
			NeverExpires:      (expiration == 0),
			UserNeverExpires:  (validAfter == 0 && validBefore == 0),
		}
		results = append(results, user)
	}
	return results, nil
}
