// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/packet-guardian/packet-guardian/src/auth"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models"
	"github.com/packet-guardian/packet-guardian/src/models/stores"

	"github.com/packet-guardian/cas-auth"
)

type CAS struct {
	e     *common.Environment
	users stores.UserStore
}

func NewCASController(e *common.Environment, us stores.UserStore) *CAS {
	return &CAS{
		e:     e,
		users: us,
	}
}

// CASHandler handles the entire CAS flow from initial redirect to session creation.
func (a *CAS) CASHandler(w http.ResponseWriter, r *http.Request) {
	if auth.IsLoggedIn(r) {
		auth.LogoutUser(w, r)
	}

	casServiceTicket := r.URL.Query().Get("ticket")

	if casServiceTicket == "" {
		// Start of authentication flow
		a.redirectCASLogin(w, r)
		return
	}

	validationResp, err := a.validateServiceTicket(casServiceTicket)
	if err != nil {
		a.e.Log.Error(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	username := validationResp.User

	if !a.e.Config.Guest.GuestOnly {
		auth.SetLoginUser(w, r, username, "cas")
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
		auth.SetLoginUser(w, r, username, "cas")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (a *CAS) redirectCASLogin(w http.ResponseWriter, r *http.Request) {
	if a.e.Config.Auth.CAS.Server == "" {
		// If CAS isn't configured, don't bother.
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	params := url.Values{
		"service": {fmt.Sprintf("%s/cas", a.e.Config.Core.SiteDomainName)},
	}

	authURL := fmt.Sprintf("%s/login?%s", a.e.Config.Auth.CAS.Server, params.Encode())
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (a *CAS) validateServiceTicket(serviceTicket string) (*cas.AuthenticationResponse, error) {
	req, err := a.buildServiceValidateRequest(serviceTicket)
	if err != nil {
		return nil, fmt.Errorf("Error building CAS serviceValidate request: %s", err.Error())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error getting CAS user info: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Non 200 response while getting CAS user info")
	}

	body, _ := ioutil.ReadAll(resp.Body)
	return cas.ParseServiceResponse(body)
}

func (a *CAS) buildServiceValidateRequest(serviceTicket string) (*http.Request, error) {
	params := url.Values{
		"service": {fmt.Sprintf("%s/cas", a.e.Config.Core.SiteDomainName)},
		"ticket":  {serviceTicket},
	}

	url := fmt.Sprintf("%s/p3/serviceValidate?%s", a.e.Config.Auth.CAS.Server, params.Encode())
	return http.NewRequest("GET", url, nil)
}
