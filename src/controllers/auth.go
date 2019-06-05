// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package controllers

import (
	"net/http"

	"github.com/packet-guardian/packet-guardian/src/auth"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
)

type Auth struct {
	e     *common.Environment
	users stores.UserStore
}

func NewAuthController(e *common.Environment, us stores.UserStore) *Auth {
	return &Auth{
		e:     e,
		users: us,
	}
}

func (a *Auth) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		a.showLoginPage(w, r)
	} else if r.Method == "POST" {
		a.loginUser(w, r)
	}
}

func (a *Auth) showLoginPage(w http.ResponseWriter, r *http.Request) {
	loggedin := a.e.Sessions.GetSession(r).GetBool("loggedin", false)
	if loggedin {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	a.e.Views.NewView("login", r).Render(w, nil)
}

func (a *Auth) loginUser(w http.ResponseWriter, r *http.Request) {
	// Assume invalid until convinced otherwise
	auth.LogoutUser(r, w)
	resp := common.NewAPIResponse("Invalid login", nil)
	ok := auth.LoginUser(r, w, a.users)

	// Bad login, return unauthorized
	if !ok {
		resp.WriteResponse(w, http.StatusUnauthorized)
		return
	}

	// If we're not in guest mode, we don't need to do anything else
	if !a.e.Config.Guest.GuestOnly {
		resp.Message = ""
		resp.WriteResponse(w, http.StatusNoContent)
		return
	}

	session := common.GetSessionFromContext(r)
	user, err := a.users.GetUserByUsername(session.GetString("username"))
	if err != nil {
		resp.Message = "Error getting user"
		resp.WriteResponse(w, http.StatusInternalServerError)
		return
	}

	// If the session user can is allowed to login with guest mode, allow them
	if user.Can(models.BypassGuestLogin) {
		resp.Message = ""
		resp.WriteResponse(w, http.StatusNoContent)
		return
	}

	// Default to deny
	resp.WriteResponse(w, http.StatusUnauthorized)
}

// LogoutHandler voids a user's session
func (a *Auth) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	auth.LogoutUser(r, w)
	if _, ok := r.URL.Query()["noredirect"]; ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}
