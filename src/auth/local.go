package auth

import (
	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/models"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	authFunctions["local"] = normalAuth
}

func normalAuth(e *common.Environment, username, password string) bool {
	user, err := models.GetUserByUsername(e, username)
	if err != nil {
		e.Log.Errorf("Error authenticating user: %s", err.Error())
		return false
	}

	testPass := user.GetPassword()

	// Always hash to avoid timing attacks
	err = bcrypt.CompareHashAndPassword([]byte(testPass), []byte(password))
	if err != nil {
		return false
	}

	// If the passwords match, check if the user is still valid
	return user.IsExpired()
}
