// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"net/http"

	"github.com/lfkeitel/verbose/v4"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	authFunctions["local"] = &localAuthenticator{}
}

type localAuthenticator struct{}

func (l *localAuthenticator) checkLogin(username, password string, r *http.Request) bool {
	e := common.GetEnvironmentFromContext(r)
	user, err := stores.GetUserStore(e).GetUserByUsername(username)
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "auth:local",
		}).Errorf("Error getting user")
		return false
	}

	testPass := user.GetPassword()
	if testPass == "" { // User doesn't have a local password
		return false
	}
	err = bcrypt.CompareHashAndPassword([]byte(testPass), []byte(password))
	if err != nil {
		if err != bcrypt.ErrMismatchedHashAndPassword {
			e.Log.WithFields(verbose.Fields{
				"error":   err,
				"package": "auth:local",
			}).Debug("Bcrypt failed")
		}
		return false
	}

	// If the passwords match, check if the user is still valid
	if user.IsExpired() {
		e.Log.WithFields(verbose.Fields{
			"username": user.Username,
			"package":  "auth:local",
		}).Info("User expired")
		return false
	}
	return true
}
