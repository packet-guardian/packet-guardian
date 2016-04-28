package auth

import (
	"net/http"

	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/models"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	authFunctions["local"] = &localAuthenticator{}
}

type localAuthenticator struct{}

func (l *localAuthenticator) loginUser(r *http.Request, w http.ResponseWriter) bool {
	e := common.GetEnvironmentFromContext(r)
	user, err := models.GetUserByUsername(e, r.FormValue("username"))
	if err != nil {
		e.Log.WithField("Err", err).Errorf("Error getting user")
		return false
	}

	testPass := user.GetPassword()
	if testPass == "" { // User doesn't have a local password
		return false
	}
	err = bcrypt.CompareHashAndPassword([]byte(testPass), []byte(r.FormValue("password")))
	if err != nil {
		e.Log.WithField("Err", err).Debug("Bcrypt failed")
		return false
	}

	// If the passwords match, check if the user is still valid
	if user.IsExpired() {
		e.Log.WithField("username", user.Username).Info("Failed login by expired user")
		return false
	}

	sess := common.GetSessionFromContext(r)
	sess.Set("loggedin", true)
	sess.Set("username", user.Username)
	sess.Save(r, w)
	return true
}

func (l *localAuthenticator) logoutUser(r *http.Request, w http.ResponseWriter) {
	sess := common.GetSessionFromContext(r)
	if sess.GetBool("loggedin") {
		sess.Set("loggedin", false)
		sess.Set("username", "")
		sess.Save(r, w)
	}
}

func (l *localAuthenticator) isLoggedIn(r *http.Request) bool {
	return common.GetSessionFromContext(r).GetBool("loggedin")
}
