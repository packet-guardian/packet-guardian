package auth

import (
	"net/http"

	"github.com/dragonrider23/verbose"
	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/models"
)

type authenticator interface {
	loginUser(r *http.Request, w http.ResponseWriter) bool
	logoutUser(r *http.Request, w http.ResponseWriter)
	isLoggedIn(r *http.Request) bool
}

var authFunctions = make(map[string]authenticator)

// LoginUser will verify the username and password against several login methods
// If one method succeeds, true will be returned. False otherwise.
func LoginUser(r *http.Request, w http.ResponseWriter) bool {
	e := common.GetEnvironmentFromContext(r)
	if r.FormValue("password") == "" || r.FormValue("username") == "" {
		return false
	}

	e.Log.Debug(e.Config.Auth.AuthMethod)
	for _, method := range e.Config.Auth.AuthMethod {
		if authMethod, ok := authFunctions[method]; ok {
			if authMethod.loginUser(r, w) {
				sess := common.GetSessionFromContext(r)
				sess.Set("_authMethod", method)
				sess.Save(r, w)
				e.Log.WithFields(verbose.Fields{
					"username": r.FormValue("username"),
					"method":   method,
				}).Info("Logged in user")
				return true
			}
		}
	}
	return false
}

func IsLoggedIn(r *http.Request) bool {
	method := common.GetSessionFromContext(r).GetString("_authMethod")
	if m, ok := authFunctions[method]; ok {
		return m.isLoggedIn(r)
	}
	return false
}

func LogoutUser(r *http.Request, w http.ResponseWriter) {
	method := common.GetSessionFromContext(r).GetString("_authMethod")
	m, ok := authFunctions[method]
	if !ok {
		return
	}
	m.logoutUser(r, w)

	e := common.GetEnvironmentFromContext(r)
	user := models.GetUserFromContext(r)
	e.Log.WithFields(verbose.Fields{
		"username": user.Username,
		"method":   method,
	}).Info("Logged out user")
}
