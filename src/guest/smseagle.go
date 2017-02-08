// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package guest

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/usi-lfkeitel/packet-guardian/src/common"
)

func init() {
	checkers["smseagle"] = smseagle{}
}

type smseagle struct{}

func (s smseagle) getInputLabel() string {
	return "Phone Number (with area code)"
}

func (s smseagle) getInputText() string {
	return "You will receive a text message. Data rates may apply."
}

func (s smseagle) getVerificationText() string {
	return "Please enter the verification code that was texted to you."
}

func (s smseagle) normalizeCredential(c string) string {
	c, _ = formatPhoneNumber(c)
	return c
}

func (s smseagle) sendCode(e *common.Environment, phone, code string) error {
	address := e.Config.Guest.Smseagle.Address
	username := e.Config.Guest.Smseagle.Username
	password := e.Config.Guest.Smseagle.Password
	highPri := e.Config.Guest.Smseagle.HighPriority
	flashMsg := e.Config.Guest.Smseagle.FlashMsg
	phone, err := formatPhoneNumber(phone)
	if err != nil {
		return err
	}

	if username == "" || password == "" {
		e.Log.Error("Username and password not given for SMSEagle")
		return errors.New("Error sending verification code")
	}

	if address == "" {
		e.Log.Error("Address not given for SMSEagle")
		return errors.New("Error sending verification code")
	}

	p := make(url.Values, 0)
	p.Add("login", username)
	p.Add("pass", password)
	p.Add("to", phone)
	p.Add("message", "Verification Code: "+code)
	p.Add("highpriority", strconv.Itoa(highPri))
	p.Add("flash", strconv.Itoa(flashMsg))

	resp, err := http.Get(address + "/index.php/http_api/send_sms?" + p.Encode())
	if err != nil {
		e.Log.WithField("error", err).Error("Error sending verification code")
		return errors.New("Error sending verification code")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		e.Log.WithField("error", err).Error("Error sending verification code")
		return errors.New("Error sending verification code")
	}

	// For some reason the SMSEagle API doesn't use proper HTTP codes, so this
	// is probably unnecassary but it's a good catch just in case.
	if resp.StatusCode != 200 {
		e.Log.WithField("error", body).Error("Error sending verification code")
		return errors.New("Error sending verification code")
	}

	// A good response looks like "OK; ID=[ID of message in outbox]"
	// I don't care about the ID
	if !bytes.HasPrefix(body, []byte("OK")) {
		e.Log.WithField("error", body).Error("Error sending verification code")
		return errors.New("Error sending verification code")
	}

	return nil
}
