// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"net/http"
	"strings"

	"github.com/lfkeitel/verbose"
	"github.com/usi-lfkeitel/packet-guardian/src/common"
	"github.com/usi-lfkeitel/packet-guardian/src/models"
)

type authenticator interface {
	loginUser(r *http.Request, w http.ResponseWriter) bool
}

var authFunctions = make(map[string]authenticator)

// LoginUser will verify the username and password against several login methods
// If one method succeeds, true will be returned. False otherwise.
func LoginUser(r *http.Request, w http.ResponseWriter) bool {
	e := common.GetEnvironmentFromContext(r)
	if r.FormValue("password") == "" || r.FormValue("username") == "" {
		return false
	}

	for _, method := range e.Config.Auth.AuthMethod {
		if authMethod, ok := authFunctions[method]; ok {
			if authMethod.loginUser(r, w) {
				sess := common.GetSessionFromContext(r)
				sess.Set("loggedin", true)
				sess.Set("username", strings.ToLower(r.FormValue("username")))
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
	e.Log.WithField("username", r.FormValue("username")).Info("Failed login")
	return false
}

func IsLoggedIn(r *http.Request) bool {
	return common.GetSessionFromContext(r).GetBool("loggedin")
}

func LogoutUser(r *http.Request, w http.ResponseWriter) {
	sess := common.GetSessionFromContext(r)
	if sess.GetBool("loggedin") {
		sess.Set("loggedin", false)
		sess.Set("username", "")
		sess.Save(r, w)
	}

	e := common.GetEnvironmentFromContext(r)
	user := models.GetUserFromContext(r)
	e.Log.WithFields(verbose.Fields{
		"username": user.Username,
		"method":   sess.GetString("_authMethod"),
	}).Info("Logged out user")
}
