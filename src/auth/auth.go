package auth

import (
	"net/http"

	"github.com/onesimus-systems/packet-guardian/src/common"
)

type authFunc func(db *common.DatabaseAccessor, username, password string) bool

var authFunctions = make(map[string]authFunc)

func init() {
	authFunctions["local"] = normalAuth
	authFunctions["ldap"] = ldapAuth
}

// IsValidLogin will verify the username and password against several login methods
// If one method succeeds, true will be returned. False otherwise.
func IsValidLogin(db *common.DatabaseAccessor, username, password string) bool {
	// Check the user and pass against the defined auth functions
	// For right now we're only doing local authentication
	return authFunctions["local"](db, username, password)
}

func IsLoggedIn(r *http.Request) bool {
	return common.GetSessionFromContext(r).GetBool("loggedin")
}
