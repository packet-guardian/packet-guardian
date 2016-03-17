package auth

import (
	"net/http"

	"github.com/onesimus-systems/net-guardian/common"
)

type authFunc func(username, password string) bool

var authFunctions = make(map[string]authFunc)

// CheckLogin will verify the username and password against several login methods
// If one method succeeds, true will be returned. False otherwise.
func CheckLogin(username, password string) bool {
	// Check the user and pass against the defined auth functions
	// For right now we're only doing local authentication
	return authFunctions["local"](username, password)
}

// IsLoggedIn checks if the session associated with r is an authenticated session
func IsLoggedIn(r *http.Request) bool {
	session := common.GetSession(r, common.Config.Webserver.SessionName)
	return session.GetBool("loggedin", false)
}

// LoginPageHandler is a net/http handler for the login page
func LoginPageHandler(w http.ResponseWriter, r *http.Request) {

}
