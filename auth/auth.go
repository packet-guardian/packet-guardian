package auth

import (
	"database/sql"

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
	return true
}
