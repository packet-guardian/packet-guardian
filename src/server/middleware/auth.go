// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package middleware

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/lfkeitel/verbose"
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

		// Not logged in via cookies, check for HTTP header
		// Check for presents of authorization header
		if r.Header.Get("Authorization") == "" {
			w.Header().Add("Authorization", "Basic realm=\"Packet Guardian\"")
			common.NewAPIResponse("Login required", nil).WriteResponse(w, http.StatusUnauthorized)
			return
		}

		// Check formatting of header
		authHeader := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(authHeader) != 2 || authHeader[0] != "Basic" {
			w.Header().Add("Authorization", "Basic realm=\"Packet Guardian\"")
			common.NewAPIResponse("Invalid Authorization header", nil).WriteResponse(w, http.StatusUnauthorized)
			return
		}

		// Decode username:password part
		authHeaderDecoded, err := base64.StdEncoding.DecodeString(authHeader[1])
		if err != nil {
			w.Header().Add("Authorization", "Basic realm=\"Packet Guardian\"")
			common.NewAPIResponse("Invalid Authorization header", nil).WriteResponse(w, http.StatusUnauthorized)
			return
		}

		// Split username and password
		userPass := bytes.SplitN(authHeaderDecoded, []byte{':'}, 2)
		if len(userPass) != 2 {
			w.Header().Add("Authorization", "Basic realm=\"Packet Guardian\"")
			common.NewAPIResponse("Invalid Authorization header", nil).WriteResponse(w, http.StatusUnauthorized)
			return
		}

		// Check credentials
		username := string(userPass[0])
		if !auth.CheckLogin(username, string(userPass[1]), r) {
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

		if strings.HasPrefix(r.URL.Path, "/admin/users") && !u.Can(models.ViewUsers) {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		if (strings.HasPrefix(r.URL.Path, "/debug") || strings.HasPrefix(r.URL.Path, "/dev")) && !u.Can(models.ViewDebugInfo) {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
}
