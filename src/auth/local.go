package auth

import (
	"github.com/onesimus-systems/packet-guardian/src/common"
	"golang.org/x/crypto/bcrypt"
)

func normalAuth(db *common.DatabaseAccessor, username, password string) bool {
	if password == "" || username == "" {
		return false
	}

	stmt, err := db.Prepare("SELECT \"password\" FROM \"user\" WHERE \"username\" = ?")
	if err != nil {
		return false
	}
	user := stmt.QueryRow(username)

	var storedPass string
	err = user.Scan(&storedPass)
	if err != nil {
		return false
	}

	if storedPass == "" {
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPass), []byte(password))
	return (err == nil)
}
