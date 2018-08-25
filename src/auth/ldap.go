// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"net/http"

	ldapc "github.com/lfkeitel/go-ldap-client"
	"github.com/lfkeitel/verbose/v4"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
	"gopkg.in/ldap.v2"
)

func init() {
	authFunctions["ldap"] = &ldapAuthenticator{}
}

type ldapAuthenticator struct {
	client *ldapc.LDAPClient
}

func (l *ldapAuthenticator) checkLogin(username, password string, r *http.Request) bool {
	e := common.GetEnvironmentFromContext(r)
	// TODO: Support full LDAP servers and not just AD
	// TODO: Support multiple LDAP servers, not just one
	if l.client == nil {
		l.client = &ldapc.LDAPClient{
			Host:         e.Config.Auth.LDAP.Servers[0],
			Port:         389,
			UseSSL:       e.Config.Auth.LDAP.VerifySSLCert,
			ADDomainName: e.Config.Auth.LDAP.DomainName,
		}
	}
	defer l.client.Close()

	ok, _, err := l.client.Authenticate(username, password)
	if err != nil && !ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
		e.Log.WithFields(verbose.Fields{
			"error":    err,
			"username": username,
			"package":  "auth:ldap",
		}).Error("Error authenticating with LDAP server")
	}

	if !ok {
		return false
	}

	user, err := stores.GetUserStore(e).GetUserByUsername(username)
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "auth:ldap",
		}).Error("Error getting user")
		return false
	}
	if user.IsExpired() {
		e.Log.WithFields(verbose.Fields{
			"username": user.Username,
			"package":  "auth:ldap",
		}).Info("User expired")
		return false
	}

	return true
}
