// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package controllers

import (
	"net/http"
	"strings"
	"time"

	"github.com/dchest/captcha"
	"github.com/lfkeitel/verbose/v4"
	"github.com/packet-guardian/dhcp-lib"
	"github.com/packet-guardian/packet-guardian/src/auth"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/guest"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
)

type Guest struct {
	e       *common.Environment
	users   stores.UserStore
	devices stores.DeviceStore
	leases  stores.LeaseStore
}

func NewGuestController(e *common.Environment, us stores.UserStore, ds stores.DeviceStore, ls stores.LeaseStore) *Guest {
	return &Guest{
		e:       e,
		users:   us,
		devices: ds,
		leases:  ls,
	}
}

func (g *Guest) RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	loggedIn := auth.IsLoggedIn(r) // Only non-guests will be logged in.
	if loggedIn {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if !g.e.Config.Guest.Enabled {
		g.e.Views.NewView("user-register-guest", r).Render(w, nil)
		return
	}

	ip := common.GetIPFromContext(r)
	reg, _ := dhcp.IsRegisteredByIP(g.leases, ip)
	if reg {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == "GET" {
		g.showGuestRegPage(w, r)
		return
	}
	g.checkGuestInfo(w, r)
}

func (g *Guest) showGuestRegPage(w http.ResponseWriter, r *http.Request) {
	label := guest.GetInputLabel(g.e)
	if label == "" {
		g.renderErrorMessage("Guest registrations are currently unavailable. Please notify the IT help desk.", w, r)
		return
	}

	data := map[string]interface{}{
		"policy":         common.LoadPolicyText(g.e.Config.Registration.RegistrationPolicyFile),
		"guestCredLabel": label,
		"guestCredText":  guest.GetInputText(g.e),
		"captchaID":      captcha.New(),
	}

	g.e.Views.NewView("user-register-guest", r).Render(w, data)
}

func (g *Guest) checkGuestInfo(w http.ResponseWriter, r *http.Request) {
	if !g.e.Config.Guest.Enabled {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	session := common.GetSessionFromContext(r)

	// Check if a verification code has already been issued and not expired
	if session.GetString("_verify-code", "") != "" && session.GetInt64("_expires", 0) > time.Now().Unix()+5 {
		http.Redirect(w, r, "/register/guest/verify", http.StatusSeeOther)
	}

	if !g.e.Config.Guest.DisableCaptcha &&
		!captcha.VerifyString(r.FormValue("captchaId"), r.FormValue("captchaSolution")) {
		session.AddFlash(common.FlashMessage{Message: "Incorrect Captcha answer", Type: common.FlashMessageError})
		g.showGuestRegPage(w, r)
		return
	}

	guestCred := guest.NormalizeCredential(g.e, r.FormValue("guest-cred"))
	guestName := r.FormValue("guest-name")

	if guestCred == "" || guestName == "" {
		session.AddFlash(common.FlashMessage{Message: "Please fill in all required fields", Type: common.FlashMessageError})
		g.showGuestRegPage(w, r)
		return
	}

	guestUser, err := g.users.GetUserByUsername(guestCred)
	if err != nil {
		g.e.Log.WithFields(verbose.Fields{
			"error":    err,
			"package":  "controllers:guest",
			"username": guestCred,
		}).Error("Error getting user")
		g.e.Views.RenderError(w, r, nil)
		return
	}

	if guestUser.IsBlacklisted() {
		session.AddFlash(common.FlashMessage{Message: "Permission Denied", Type: common.FlashMessageError})
		g.showGuestRegPage(w, r)
		return
	}

	verifyCode := guest.GenerateGuestCode()
	session.Set("_verify-code", verifyCode)
	session.Set("_expires", time.Now().Add(time.Duration(g.e.Config.Guest.VerifyCodeExpiration)*time.Minute).Unix())
	session.Set("_guest-credential", guestCred)
	session.Set("_guest-name", guestName)
	session.Save(r, w)
	if err := guest.SendGuestCode(g.e, guestCred, verifyCode); err != nil {
		session.AddFlash(common.FlashMessage{Message: err.Error(), Type: common.FlashMessageError})
		g.showGuestRegPage(w, r)
		return
	}
	g.e.Log.WithField("verify-code", verifyCode).Debug("Guest code")
	http.Redirect(w, r, "/register/guest/verify", http.StatusSeeOther)
}

func (g *Guest) VerificationHandler(w http.ResponseWriter, r *http.Request) {
	if !g.e.Config.Guest.Enabled {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	loggedIn := auth.IsLoggedIn(r)
	if loggedIn {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	ip := common.GetIPFromContext(r)
	reg, _ := dhcp.IsRegisteredByIP(g.leases, ip)
	if reg {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	session := common.GetSessionFromContext(r)
	if session.GetString("_verify-code") == "" {
		http.Redirect(w, r, "/register/guest", http.StatusSeeOther)
		return
	}

	if r.Method == "GET" {
		g.showGuestVerifyPage(w, r)
		return
	}
	g.verifyGuestRegistration(w, r)
}

func (g *Guest) showGuestVerifyPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"guestVerText": guest.GetVerificationText(g.e),
	}
	g.e.Views.NewView("user-register-guest-verify", r).Render(w, data)
}

func (g *Guest) verifyGuestRegistration(w http.ResponseWriter, r *http.Request) {
	session := common.GetSessionFromContext(r)
	if session.GetInt64("_expires") < time.Now().Unix() {
		session.AddFlash(common.FlashMessage{Message: "Verification code has expired", Type: common.FlashMessageError})
		session.Save(r, w)
		http.Redirect(w, r, "/register/guest", http.StatusSeeOther)
		return
	}

	if session.GetString("_verify-code") != strings.ToUpper(r.FormValue("verify-code")) {
		session.AddFlash(common.FlashMessage{Message: "Incorrect verification code", Type: common.FlashMessageError})
		g.showGuestVerifyPage(w, r)
		return
	}

	session.Delete(r, w)
	if err := guest.RegisterDevice(
		g.e,
		session.GetString("_guest-name"),
		session.GetString("_guest-credential"),
		r,
		g.users,
		g.devices,
		g.leases,
	); err != nil {
		g.renderErrorMessage(err.Error(), w, r)
		return
	}
	g.renderMessage("Please disconnect your computer and reconnect to the network", w, r)
}

func (g *Guest) renderErrorMessage(message string, w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"msg":   message,
		"error": true,
	}
	g.e.Views.NewView("user-register-guest-msg", r).Render(w, data)
}

func (g *Guest) renderMessage(message string, w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"msg":   message,
		"error": false,
	}
	g.e.Views.NewView("user-register-guest-msg", r).Render(w, data)
}
