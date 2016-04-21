package auth

import (
	ldapc "github.com/dragonrider23/go-ldap-client"
	"github.com/onesimus-systems/packet-guardian/src/common"
	"gopkg.in/ldap.v2"
)

func init() {
	authFunctions["ldap"] = ldapAuth
}

var ldapClient *ldapc.LDAPClient

func ldapAuth(e *common.Environment, username, password string) bool {
	// TODO: Support full LDAP servers and not just AD
	// TODO: Support multiple LDAP servers, not just one
	if ldapClient == nil {
		ldapClient = &ldapc.LDAPClient{
			Host:         e.Config.Auth.LDAP.Servers[0],
			Port:         389,
			UseSSL:       false, //e.Config.Auth.LDAP.UseTLS,
			ADDomainName: e.Config.Auth.LDAP.DomainName,
		}
	}
	defer ldapClient.Close()

	ok, _, err := ldapClient.Authenticate(username, password)
	if err != nil && !ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
		e.Log.Infof("%#v", err)
		e.Log.Errorf("Error authenticating user %s: %s", username, err.Error())
	}
	return ok
}
