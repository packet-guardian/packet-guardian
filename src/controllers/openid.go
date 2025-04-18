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
	"time"

	"github.com/packet-guardian/packet-guardian/src/auth"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models"
	"github.com/packet-guardian/packet-guardian/src/models/stores"

	"github.com/gofrs/uuid"
)

const (
	openIDStateCookie = "PG_OPENID_STATE"
)

type openIDTokenResp struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	IDToken     string `json:"id_token"`
}

type openIDUserInfoResp struct {
	Sub               string `json:"sub"`
	Name              string `json:"name"`
	Username          string `json:"username"`
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
}

type OpenID struct {
	e     *common.Environment
	users stores.UserStore
}

func NewOpenIDController(e *common.Environment, us stores.UserStore) *OpenID {
	return &OpenID{
		e:     e,
		users: us,
	}
}

// OpenIDHandler handles the entire OpenID flow from initial redirect
// to final token retrieval and session creation.
func (a *OpenID) OpenIDHandler(w http.ResponseWriter, r *http.Request) {
	if auth.IsLoggedIn(r) {
		auth.LogoutUser(w, r)
	}

	openIDCode := r.URL.Query().Get("code")
	openIDState := r.URL.Query().Get("state")

	if openIDCode == "" || openIDState == "" {
		// Start of authentication flow
		a.redirectOpenID(w, r)
		return
	}

	clientStateCookie, _ := r.Cookie(openIDStateCookie)
	if clientStateCookie == nil || clientStateCookie.Value != openIDState {
		a.e.Log.Error("OpenID state mismatch")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	tokenResp, err := a.getOpenIDTokens(openIDCode)
	if err != nil {
		a.e.Log.Error(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	userInfoResp, err := a.getOpenIDUserInfo(tokenResp.AccessToken)
	if err != nil {
		a.e.Log.Error(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	username := userInfoResp.PreferredUsername
	if username == "" { // No preferred username
		if userInfoResp.Username != "" {
			username = userInfoResp.Username
		} else if userInfoResp.Email != "" {
			a.e.Log.Info("No username returned from OpenID server, using email")
			username = userInfoResp.Email
		} else {
			a.e.Log.Info("No email returned from OpenID server, failing login")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
	}

	if a.e.Config.Auth.Openid.StripDomain {
		username = strings.SplitN(username, "@", 2)[0]
	}

	if !a.e.Config.Guest.GuestOnly {
		auth.SetLoginUser(w, r, username, "openid")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	user, err := a.users.GetUserByUsername(username)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// If the session user is allowed to login with guest mode, allow them
	if user.Can(models.BypassGuestLogin) {
		auth.SetLoginUser(w, r, username, "openid")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (a *OpenID) redirectOpenID(w http.ResponseWriter, r *http.Request) {
	if a.e.Config.Auth.Openid.Server == "" {
		// If OpenID isn't configured, don't bother.
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	stateID, _ := uuid.NewV4()

	params := url.Values{
		"client_id":     {a.e.Config.Auth.Openid.ClientID},
		"response_type": {"code"},
		"scope":         {"openid profile email"},
		"redirect_uri":  {fmt.Sprintf("%s/openid", a.e.Config.Core.SiteDomainName)},
		"state":         {stateID.String()},
	}

	// Save state value to compare with when the authorization code comes back
	http.SetCookie(w, &http.Cookie{
		Name:     openIDStateCookie,
		Value:    stateID.String(),
		Path:     "/",
		Expires:  time.Now().Add(5 * time.Minute),
		HttpOnly: true,
	})

	authURL := fmt.Sprintf("%s?%s", a.e.Config.Auth.Openid.AuthorizeEndoint, params.Encode())
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (a *OpenID) getOpenIDTokens(authCode string) (*openIDTokenResp, error) {
	req, err := a.buildOpenIDTokenRequest(authCode)
	if err != nil {
		return nil, fmt.Errorf("Error building OpenID token request: %s", err.Error())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error getting OpenID tokens: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Non 200 response while getting OpenID tokens")
	}

	var tokenResp openIDTokenResp
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("Error decoding OpenID tokens: %s", err.Error())
	}

	return &tokenResp, nil
}

func (a *OpenID) buildOpenIDTokenRequest(authCode string) (*http.Request, error) {
	formValues := url.Values{
		"grant_type":   {"authorization_code"},
		"redirect_uri": {fmt.Sprintf("%s/openid", a.e.Config.Core.SiteDomainName)},
		"code":         {authCode},
	}

	tokenURL := a.e.Config.Auth.Openid.TokenEndoint
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(formValues.Encode()))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(a.e.Config.Auth.Openid.ClientID, a.e.Config.Auth.Openid.ClientSecret)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

func (a *OpenID) getOpenIDUserInfo(accessToken string) (*openIDUserInfoResp, error) {
	req, err := a.buildOpenIDUserinfoRequest(accessToken)
	if err != nil {
		return nil, fmt.Errorf("Error building OpenID introspect request: %s", err.Error())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error getting OpenID introspection: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Non 200 response while getting OpenID introspection")
	}

	var userinfoResp openIDUserInfoResp
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&userinfoResp); err != nil {
		return nil, fmt.Errorf("Error decoding OpenID introspection: %s", err.Error())
	}

	return &userinfoResp, nil
}

func (a *OpenID) buildOpenIDUserinfoRequest(accessToken string) (*http.Request, error) {
	userInfoURL := a.e.Config.Auth.Openid.UserinfoEndpoint
	req, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	return req, nil
}
