package auth

import (
	"strconv"

	"github.com/oec/goradius"
	"github.com/onesimus-systems/packet-guardian/src/common"
)

func init() {
	authFunctions["radius"] = radiusAuth
}

var radAuther *radius.Authenticator

func radiusAuth(e *common.Environment, username, password string) bool {
	if radAuther == nil {
		radAuther = radius.New(
			e.Config.Auth.Radius.Servers[0],
			strconv.Itoa(e.Config.Auth.Radius.Port),
			e.Config.Auth.Radius.Secret,
		)
	}
	ok, err := radAuther.Authenticate(username, password)
	if err != nil {
		e.Log.Errorf("Error authenticating against radius: %s", err.Error())
	}
	return ok
}
