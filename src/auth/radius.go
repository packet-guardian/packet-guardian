// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"net/http"
	"strconv"

	"github.com/lfkeitel/verbose"
	"github.com/oec/goradius"
	"github.com/usi-lfkeitel/packet-guardian/src/common"
	"github.com/usi-lfkeitel/packet-guardian/src/models"
)

func init() {
	authFunctions["radius"] = &radAuthenticator{}
}

type radAuthenticator struct {
	auther *radius.Authenticator
}

func (rad *radAuthenticator) checkLogin(username, password string, r *http.Request) bool {
	e := common.GetEnvironmentFromContext(r)
	if rad.auther == nil {
		rad.auther = radius.New(
			e.Config.Auth.Radius.Servers[0],
			strconv.Itoa(e.Config.Auth.Radius.Port),
			e.Config.Auth.Radius.Secret,
		)
	}
	ok, err := rad.auther.Authenticate(username, password)
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "auth:radius",
		}).Error("Error authenticating against radius server")
		return false
	}

	if !ok {
		return false
	}

	user, err := models.GetUserByUsername(e, username)
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "auth:radius",
		}).Error("Error getting user")
		return false
	}
	if user.IsExpired() {
		e.Log.WithFields(verbose.Fields{
			"username": user.Username,
			"package":  "auth:radius",
		}).Info("User expired")
		return false
	}

	return true
}
