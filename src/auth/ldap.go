package auth

import (
	"github.com/jtblin/go-ldap-client"
	"github.com/onesimus-systems/packet-guardian/src/common"
)

func init() {
	authFunctions["ldap"] = ldapAuth
	authFunctions["ad"] = ldapAuth
}

var ldapClient *ldap.LDAPClient

func ldapAuth(e *common.Environment, username, password string) bool {
	// TODO: Support full LDAP servers and not just AD
	// TODO: Support multiple LDAP servers, not just one
	if ldapClient == nil {
		ldapClient = &ldap.LDAPClient{
			Host:         e.Config.Auth.LDAP.Servers[0],
			Port:         389,
			UseSSL:       e.Config.Auth.LDAP.UseTLS,
			ADDomainName: e.Config.Auth.LDAP.DomainName,
		}
	}
	defer ldapClient.Close()

	ok, _, err := ldapClient.Authenticate(username, password)
	if err != nil {
		e.Log.Errorf("Error authenticating user %s: %s", username, err.Error())
	}
	return ok
}
