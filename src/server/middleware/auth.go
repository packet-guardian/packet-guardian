package middleware

import (
	"net/http"
	"strings"

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

func CheckAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsLoggedIn(r) {
			common.NewAPIResponse("Login required", nil).WriteResponse(w, http.StatusUnauthorized)
			return
		}

		// The device handler checks for appropiate pernissions
		if strings.HasPrefix(r.URL.Path, "/api/device") {
			next.ServeHTTP(w, r)
			return
		}

		u := models.GetUserFromContext(r)
		// Only admin and helpdesk users may proceed
		if !u.IsAdmin() && !u.IsHelpDesk() {
			common.NewAPIResponse("Insufficient privilages", nil).WriteResponse(w, http.StatusForbidden)
			return
		}

		if r.Method != "GET" && !u.IsAdmin() {
			common.NewAPIResponse("Admins only", nil).WriteResponse(w, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func CheckAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := models.GetUserFromContext(r)
		// Only admin and helpdesk users may proceed
		if !u.IsAdmin() && !u.IsHelpDesk() {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		// /admin/users is for full admins only
		if strings.HasPrefix(r.URL.Path, "/admin/users") && !u.IsAdmin() {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
}
