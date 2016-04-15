package middleware

import (
	"net/http"

	"github.com/onesimus-systems/packet-guardian/src/auth"
	"github.com/onesimus-systems/packet-guardian/src/common"
)

// CheckAuthMid is middleware to check if a user is logged in, if not it will redirect to the login page
func CheckAuth(e *common.Environment, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsLoggedIn(e, r) {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		next(w, r)
	}
}

// CheckAdminMid is middleware that checks if a user is an administrator, it calls
// the CheckAuthMid middleware before checking itself
func CheckAdmin(e *common.Environment, next http.HandlerFunc) http.HandlerFunc {
	admin := func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsAdminUser(e, r) {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		next(w, r)
	}
	return CheckAuth(e, admin)
}

// CheckAuthAPIMid is middleware to check if a user is logged in, if not it will return an AuthNeeded api status
func CheckAuthAPI(e *common.Environment, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsLoggedIn(e, r) {
			common.NewAPIResponse(common.APIStatusAuthNeeded, "Not logged in", nil).WriteTo(w)
			return
		}
		next(w, r)
	}
}

// CheckAdminAPIMid is middleware that checks if a user is an administrator, it calls
// the CheckAuthAPIMid middleware before checking itself
func CheckAdminAPI(e *common.Environment, next http.HandlerFunc) http.HandlerFunc {
	admin := func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsAdminUser(e, r) {
			common.NewAPIResponse(common.APIStatusNotAdmin, "Not an administrator", nil).WriteTo(w)
			return
		}
		next(w, r)
	}
	return CheckAuthAPI(e, admin)
}
