package auth

import (
	"net/http"

	"github.com/onesimus-systems/packet-guardian/src/common"
)

type authFunc func(e *common.Environment, username, password string) bool

var authFunctions = make(map[string]authFunc)

// IsValidLogin will verify the username and password against several login methods
// If one method succeeds, true will be returned. False otherwise.
func IsValidLogin(e *common.Environment, username, password string) bool {
	if password == "" || username == "" {
		return false
	}

	for _, method := range e.Config.Auth.AuthMethod {
		if authMethod, ok := authFunctions[method]; ok {
			if authMethod(e, username, password) {
				return true
			}
		}
	}
	return false
}

func IsLoggedIn(r *http.Request) bool {
	return common.GetSessionFromContext(r).GetBool("loggedin")
}
