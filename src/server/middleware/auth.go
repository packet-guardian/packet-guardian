package middleware

import (
	"net/http"

	"github.com/onesimus-systems/packet-guardian/src/auth"
	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/models"
)

// CheckAuthMid is middleware to check if a user is logged in, if not it will redirect to the login page
func CheckAuth(e *common.Environment, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsLoggedIn(r) {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CheckAdminMid is middleware that checks if a user is an administrator, it calls
// the CheckAuthMid middleware before checking itself
func CheckAdmin(e *common.Environment, next http.Handler) http.Handler {
	admin := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !models.GetUserFromContext(r).IsAdmin() {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
	return CheckAuth(e, admin)
}

// CheckAuthAPIMid is middleware to check if a user is logged in, if not it will return an AuthNeeded api status
func CheckAuthAPI(e *common.Environment, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsLoggedIn(r) {
			common.NewAPIResponse(common.APIStatusAuthNeeded, "Not logged in", nil).WriteTo(w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CheckAdminAPIMid is middleware that checks if a user is an administrator, it calls
// the CheckAuthAPIMid middleware before checking itself
func CheckAdminAPI(e *common.Environment, next http.Handler) http.Handler {
	admin := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !models.GetUserFromContext(r).IsAdmin() {
			common.NewAPIResponse(common.APIStatusNotAdmin, "Not an administrator", nil).WriteTo(w)
			return
		}
		next.ServeHTTP(w, r)
	})
	return CheckAuthAPI(e, admin)
}
