// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package middleware

import (
	"net/http"
	"strings"

	"github.com/lfkeitel/verbose/v4"
	"github.com/packet-guardian/packet-guardian/src/auth"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
)

// CheckAuth is middleware to check if a user is logged in, if not it will redirect to the login page
func CheckAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsLoggedIn(r) {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(w, r)
		return
	})
}

// CheckAuth is middleware to check if a user is logged in, if not it will redirect to the login page
func CheckAuthAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth.IsLoggedIn(r) {
			next.ServeHTTP(w, r)
			return
		}

		username, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Add("Authorization", "Basic realm=\"Packet Guardian\"")
			common.NewAPIResponse("Login required", nil).WriteResponse(w, http.StatusUnauthorized)
			return
		}

		if !auth.CheckLogin(username, password, r) {
			w.Header().Add("Authorization", "Basic realm=\"Packet Guardian\"")
			common.NewAPIResponse("Invalid username or password", nil).WriteResponse(w, http.StatusUnauthorized)
			return
		}

		// Get user model
		e := common.GetEnvironmentFromContext(r)
		sessionUser, err := stores.GetUserStore(e).GetUserByUsername(username)
		if err != nil {
			e.Log.WithFields(verbose.Fields{
				"error":    err,
				"package":  "middleware:checkauth",
				"username": username,
			}).Error("Error getting session user")
			common.NewAPIResponse("Internal Server Error", nil).WriteResponse(w, http.StatusInternalServerError)
			return
		}

		// Check for API permissions
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" { // Read requests
			if !sessionUser.Can(models.APIRead) {
				common.NewAPIResponse("Permission denied", nil).WriteResponse(w, http.StatusUnauthorized)
				return
			}
		} else if !sessionUser.Can(models.APIWrite) { // Write requests
			common.NewAPIResponse("Permission denied", nil).WriteResponse(w, http.StatusUnauthorized)
			return
		}

		// Add user to context
		r = models.SetUserToContext(r, sessionUser)

		// Continue
		next.ServeHTTP(w, r)
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

		for prefix, permission := range adminPagePermissions {
			if strings.HasPrefix(r.URL.Path, prefix) {
				if !u.Can(permission) {
					http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
					return
				}
				break
			}
		}

		next.ServeHTTP(w, r)
	})
}

var adminPagePermissions = map[string]models.Permission{
	"/admin/users": models.ViewUsers,
	"/debug":       models.ViewDebugInfo,
	"/dev":         models.ViewDebugInfo,
}
