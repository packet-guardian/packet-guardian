// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/packet-guardian/packet-guardian/src/auth"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
)

type openIDTokenResp struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	IDToken     string `json:"id_token"`
}

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

func (a *Auth) OpenIDHandler(w http.ResponseWriter, r *http.Request) {
	openIDCode := r.URL.Query().Get("code")
	openIDState := r.URL.Query().Get("state")

	if openIDCode == "" || openIDState == "" {
		// Start of authentication flow
		a.redirectOpenID(w, r)
		return
	}

	req, err := a.buildOpenIDRedirectRequest(openIDCode)
	if err != nil {
		a.e.Log.Errorf("Error building OpenID token request: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		a.e.Log.Errorf("Error getting OpenID tokens: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		a.e.Log.Error("Non 200 response while getting OpenID tokens")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var tokenResp openIDTokenResp
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&tokenResp); err != nil {
		a.e.Log.Errorf("Error decoding OpenID tokens: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Need to validate
	a.e.Log.Infof("%#v", tokenResp)
	w.WriteHeader(http.StatusNoContent)
}

func (a *Auth) buildOpenIDRedirectRequest(authCode string) (*http.Request, error) {
	formValues := url.Values{
		"grant_type":   {"authorization_code"},
		"redirect_uri": {fmt.Sprintf("%s/openid", a.e.Config.Core.SiteDomainName)},
		"code":         {authCode},
	}

	tokenURL := fmt.Sprintf("%s/oauth2/default/v1/token", a.e.Config.Auth.Openid.Server)
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(formValues.Encode()))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(a.e.Config.Auth.Openid.ClientID, a.e.Config.Auth.Openid.ClientSecret)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

func (a *Auth) redirectOpenID(w http.ResponseWriter, r *http.Request) {
	params := url.Values{
		"client_id":     {a.e.Config.Auth.Openid.ClientID},
		"response_type": {"code"},
		"scope":         {"openid"},
		"redirect_uri":  {fmt.Sprintf("%s/openid", a.e.Config.Core.SiteDomainName)},
		"state":         {"state-authtest"}, // TODO: Generate random code and put in session
	}

	authURL := fmt.Sprintf("%s/oauth2/default/v1/authorize?%s", a.e.Config.Auth.Openid.Server, params.Encode())
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}
