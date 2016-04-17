package middleware

import (
	"net/http"

	"github.com/onesimus-systems/packet-guardian/src/auth"
	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/models"
)

// CheckAuth is middleware to check if a user is logged in, if not it will redirect to the login page
func CheckAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsLoggedIn(r) {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CheckAdmin is middleware that checks if a user is an administrator, it calls
// the CheckAuth middleware before checking itself
func CheckAdmin(next http.Handler) http.Handler {
	admin := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !models.GetUserFromContext(r).IsAdmin() {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
	return CheckAuth(admin)
}

// CheckRead is middleware that checks if a user is at least a help desk user, it calls
// the CheckAuth middleware before checking itself
func CheckRead(next http.Handler) http.Handler {
	admin := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := models.GetUserFromContext(r)
		if !u.IsHelpDesk() && !u.IsAdmin() {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
	return CheckAuth(admin)
}

// CheckAuthAPI is middleware to check if a user is logged in, if not it will return an AuthNeeded api status
func CheckAuthAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsLoggedIn(r) {
			common.NewAPIResponse(common.APIStatusAuthNeeded, "Not logged in", nil).WriteTo(w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CheckAdminAPI is middleware that checks if a user is an administrator, it calls
// the CheckAuthAPI middleware before checking itself
func CheckAdminAPI(next http.Handler) http.Handler {
	admin := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !models.GetUserFromContext(r).IsAdmin() {
			common.NewAPIResponse(common.APIStatusInsufficientPrivilages, "Not an administrator", nil).WriteTo(w)
			return
		}
		next.ServeHTTP(w, r)
	})
	return CheckAuthAPI(admin)
}

// CheckRead is middleware that checks if a user is at least a help desk user, it calls
// the CheckAuth middleware before checking itself
func CheckReadAPI(next http.Handler) http.Handler {
	admin := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := models.GetUserFromContext(r)
		if !u.IsHelpDesk() && !u.IsAdmin() {
			common.NewAPIResponse(common.APIStatusInsufficientPrivilages, "Not a help desk user", nil).WriteTo(w)
			return
		}
		next.ServeHTTP(w, r)
	})
	return CheckAuthAPI(admin)
}
