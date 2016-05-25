// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"net/http"

	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/models"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	authFunctions["local"] = &localAuthenticator{}
}

type localAuthenticator struct{}

func (l *localAuthenticator) loginUser(r *http.Request, w http.ResponseWriter) bool {
	e := common.GetEnvironmentFromContext(r)
	user, err := models.GetUserByUsername(e, r.FormValue("username"))
	if err != nil {
		e.Log.WithField("Err", err).Errorf("Error getting user")
		return false
	}

	testPass := user.GetPassword()
	if testPass == "" { // User doesn't have a local password
		return false
	}
	err = bcrypt.CompareHashAndPassword([]byte(testPass), []byte(r.FormValue("password")))
	if err != nil {
		if err != bcrypt.ErrMismatchedHashAndPassword {
			e.Log.WithField("Err", err).Debug("Bcrypt failed")
		}
		return false
	}

	// If the passwords match, check if the user is still valid
	if user.IsExpired() {
		e.Log.WithField("username", user.Username).Info("User expired")
		return false
	}
	return true
}
