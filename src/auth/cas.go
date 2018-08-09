// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/lfkeitel/verbose"
	"github.com/packet-guardian/cas-auth"

	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
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
		casURLStr := strings.TrimRight(e.Config.Auth.CAS.Server, "/") + "/" // Ensure server ends in /
		casURL, err := url.Parse(casURLStr)
		if err != nil {
			e.Log.WithFields(verbose.Fields{
				"error":   err,
				"url":     casURLStr,
				"package": "auth:cas",
			}).Error("Failed to parse CAS url")
			return false
		}

		c.client = &cas.Client{
			URL: casURL,
		}

		if e.Config.Auth.CAS.ServiceURL != "" {
			serviceURL, err := url.Parse(e.Config.Auth.CAS.ServiceURL)
			if err != nil {
				e.Log.WithFields(verbose.Fields{
					"error":   err,
					"url":     e.Config.Auth.CAS.ServiceURL,
					"package": "auth:cas",
				}).Notice("Failed to parse CAS request url, using default")
			} else {
				c.client.ServiceURL = serviceURL
			}
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

	user, err := stores.GetUserStore(e).GetUserByUsername(username)
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
