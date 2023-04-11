// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
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
			redirectURL := fmt.Sprintf("/login?redirect=%s", r.URL.String())
			http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CheckAuth is middleware to check if a user is logged in, if not it will redirect to the login page
func CheckAuthAPI(next http.Handler, users stores.UserStore) http.Handler {
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

		if !auth.CheckLogin(username, password, r, users) {
			w.Header().Add("Authorization", "Basic realm=\"Packet Guardian\"")
			common.NewAPIResponse("Invalid username or password", nil).WriteResponse(w, http.StatusUnauthorized)
			return
		}

		// Get user model
		e := common.GetEnvironmentFromContext(r)
		sessionUser, err := users.GetUserByUsername(username)
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

// A PermissionChecker takes a request and a user and determines if the user
// has sufficient permissions to perform the request.
type PermissionChecker func(*http.Request, *models.User) bool

// CheckPermissions middleware ensures a session user has the required
// permissions to fulfill the request.
func CheckPermissions(next httprouter.Handle, checker PermissionChecker) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		u := models.GetUserFromContext(r)
		if !checker(r, u) {
			common.NewAPIResponse("Permission denied", nil).WriteResponse(w, http.StatusUnauthorized)
			return
		}

		next(w, r, p)
	}
}

// PermsCanAny returns a permission checker that checks if the user can
// perform any of the supplied permissions. If any permission is allowed,
// the request is allowed.
func PermsCanAny(permissions ...models.Permission) PermissionChecker {
	return func(r *http.Request, user *models.User) bool {
		for _, p := range permissions {
			if user.Can(p) {
				return true
			}
		}
		return false
	}
}
