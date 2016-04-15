package auth

import (
	"github.com/onesimus-systems/packet-guardian/src/common"
)

func ldapAuth(db *common.DatabaseAccessor, username, password string) bool {
	// Check username and pass against an ldap server
	return false
}
