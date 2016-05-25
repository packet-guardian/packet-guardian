// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"net/http"
	"strconv"

	"github.com/oec/goradius"
	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/models"
)

func init() {
	authFunctions["radius"] = &radAuthenticator{}
}

type radAuthenticator struct {
	auther *radius.Authenticator
}

func (rad *radAuthenticator) loginUser(r *http.Request, w http.ResponseWriter) bool {
	e := common.GetEnvironmentFromContext(r)
	if rad.auther == nil {
		rad.auther = radius.New(
			e.Config.Auth.Radius.Servers[0],
			strconv.Itoa(e.Config.Auth.Radius.Port),
			e.Config.Auth.Radius.Secret,
		)
	}
	ok, err := rad.auther.Authenticate(r.FormValue("username"), r.FormValue("password"))
	if err != nil {
		e.Log.Errorf("Error authenticating against radius: %s", err.Error())
		return false
	}

	if !ok {
		return false
	}

	user, err := models.GetUserByUsername(e, r.FormValue("username"))
	if err != nil {
		e.Log.WithField("Err", err).Errorf("Error getting user")
		return false
	}
	if user.IsExpired() {
		e.Log.WithField("username", user.Username).Info("User expired")
		return false
	}
	return true
}
