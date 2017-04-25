// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"net/http"
	"strings"

	"github.com/lfkeitel/verbose"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models"
)

type authenticator interface {
	checkLogin(username, password string, r *http.Request) bool
}

var authFunctions = make(map[string]authenticator)

// LoginUser will verify the username and password against several login methods
// If one method succeeds, true will be returned. False otherwise.
func LoginUser(r *http.Request, w http.ResponseWriter) bool {
	if r.FormValue("password") == "" || r.FormValue("username") == "" {
		return false
	}

	e := common.GetEnvironmentFromContext(r)
	username := strings.ToLower(r.FormValue("username"))
	for _, method := range e.Config.Auth.AuthMethod {
		if authMethod, ok := authFunctions[method]; ok {
			if authMethod.checkLogin(username, r.FormValue("password"), r) {
				sess := common.GetSessionFromContext(r)
				sess.Set("loggedin", true)
				sess.Set("username", username)
				sess.Set("_authMethod", method)
				sess.Save(r, w)
				e.Log.WithFields(verbose.Fields{
					"username": username,
					"method":   method,
					"action":   "login",
					"package":  "auth",
				}).Info("Logged in user")
				return true
			}
		}
	}
	e.Log.WithFields(verbose.Fields{
		"username": username,
		"package":  "auth",
	}).Info("Failed login")
	return false
}

func CheckLogin(username, password string, r *http.Request) bool {
	if password == "" || username == "" {
		return false
	}

	e := common.GetEnvironmentFromContext(r)
	username = strings.ToLower(username)
	for _, method := range e.Config.Auth.AuthMethod {
		if authMethod, ok := authFunctions[method]; ok {
			if authMethod.checkLogin(username, password, r) {
				e.Log.WithFields(verbose.Fields{
					"username": username,
					"method":   method,
					"action":   "login",
					"package":  "auth",
				}).Info("Logged in user")
				return true
			}
		}
	}
	e.Log.WithFields(verbose.Fields{
		"username": username,
		"package":  "auth",
	}).Info("Failed login")
	return false
}

func IsLoggedIn(r *http.Request) bool {
	return common.GetSessionFromContext(r).GetBool("loggedin")
}

func LogoutUser(r *http.Request, w http.ResponseWriter) {
	sess := common.GetSessionFromContext(r)
	if !sess.GetBool("loggedin") {
		return
	}

	sess.Set("loggedin", false)
	sess.Set("username", "")
	sess.Save(r, w)

	e := common.GetEnvironmentFromContext(r)
	user := models.GetUserFromContext(r)
	e.Log.WithFields(verbose.Fields{
		"username": user.Username,
		"method":   sess.GetString("_authMethod"),
		"action":   "logout",
		"package":  "auth",
	}).Info("Logged out user")
}
