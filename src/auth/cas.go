// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/dragonlibs/cas"
	"github.com/lfkeitel/verbose"

	"github.com/usi-lfkeitel/packet-guardian/src/common"
	"github.com/usi-lfkeitel/packet-guardian/src/models"
)

func init() {
	authFunctions["cas"] = &casAuthenticator{}
}

type casAuthenticator struct {
	client *cas.Client
}

func (c *casAuthenticator) checkLogin(username, password string, r *http.Request) bool {
	e := common.GetEnvironmentFromContext(r)
	if c.client == nil {
		casUrlStr := strings.TrimRight(e.Config.Auth.CAS.Server, "/") + "/" // Ensure server ends in /
		casUrl, err := url.Parse(casUrlStr)
		if err != nil {
			e.Log.WithFields(verbose.Fields{
				"error":   err,
				"url":     casUrlStr,
				"package": "auth:cas",
			}).Error("Failed to parse CAS url")
			return false
		}
		c.client = &cas.Client{
			URL: casUrl,
		}
	}

	_, err := c.client.AuthenticateUser(username, password, r)
	if err == cas.InvalidCredentials {
		return false
	}
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "auth:cas",
		}).Error("Error communicating with CAS server")
		return false
	}

	user, err := models.GetUserByUsername(e, username)
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "auth:cas",
		}).Error("Error getting user")
		return false
	}
	if user.IsExpired() {
		e.Log.WithFields(verbose.Fields{
			"username": user.Username,
			"package":  "auth:cas",
		}).Info("User expired")
		return false
	}

	return true
}
