// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package middleware

import (
	"net/http"
	"strings"

	"github.com/usi-lfkeitel/packet-guardian/src/auth"
	"github.com/usi-lfkeitel/packet-guardian/src/common"
	"github.com/usi-lfkeitel/packet-guardian/src/models"
)

// CheckAuth is middleware to check if a user is logged in, if not it will redirect to the login page
func CheckAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth.IsLoggedIn(r) {
			next.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api") {
			common.NewAPIResponse("Login required", nil).WriteResponse(w, http.StatusUnauthorized)
			return
		}
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	})
}

func CheckAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := models.GetUserFromContext(r)
		// Only admin and helpdesk users may proceed
		if !u.Can(models.ViewAdminPage) {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		// /admin/users is for full admins only
		if strings.HasPrefix(r.URL.Path, "/admin/users") && !u.Can(models.ViewUsers) {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
}
