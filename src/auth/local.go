package auth

import (
	"github.com/onesimus-systems/packet-guardian/src/common"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	authFunctions["local"] = normalAuth
}

func normalAuth(e *common.Environment, username, password string) bool {
	result := e.DB.QueryRow(`SELECT "password" FROM "user" WHERE "username" = ?`, username)

	var storedPass string
	err := result.Scan(&storedPass)
	if err != nil {
		return false
	}

	if storedPass == "" {
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPass), []byte(password))
	return (err == nil)
}
