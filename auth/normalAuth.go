package auth

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/onesimus-systems/net-guardian/common"
)

func normalAuth(username, password string) bool {
	stmt, err := common.DB.Prepare("SELECT \"password\" FROM \"user\" WHERE \"username\" = ?")
	if err != nil {
		return false
	}
	user := stmt.QueryRow(username)

	var storedPass string
	err = user.Scan(&storedPass)
	if err != nil {
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPass), []byte(password))
	if err == nil {
		return true
	}
	return false
}

func init() {
	authFunctions["local"] = normalAuth
}
