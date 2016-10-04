// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"net/http"

	ldapc "github.com/lfkeitel/go-ldap-client"
	"github.com/lfkeitel/verbose"
	"github.com/usi-lfkeitel/packet-guardian/src/common"
	"github.com/usi-lfkeitel/packet-guardian/src/models"
	"gopkg.in/ldap.v2"
)

func init() {
	authFunctions["ldap"] = &ldapAuthenticator{}
}

type ldapAuthenticator struct {
	client *ldapc.LDAPClient
}

func (l *ldapAuthenticator) loginUser(r *http.Request, w http.ResponseWriter) bool {
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

	ok, _, err := l.client.Authenticate(r.FormValue("username"), r.FormValue("password"))
	if err != nil && !ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
		e.Log.WithFields(verbose.Fields{
			"error":    err,
			"username": r.FormValue("username"),
			"package":  "auth:ldap",
		}).Error("Error authenticating with LDAP server")
	}

	if !ok {
		return false
	}

	user, err := models.GetUserByUsername(e, r.FormValue("username"))
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
		user.Release()
		return false
	}

	user.Release()
	return true
}
