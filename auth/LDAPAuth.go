package auth

func ldapAuth(username, password string) bool {
	// Check username and pass against an ldap server
	return true
}

func init() {
	authFunctions["ldap"] = ldapAuth
}
