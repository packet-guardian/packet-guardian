// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package guest

import (
	"errors"

	"bitbucket.org/ckvist/twilio/twirest"
	"github.com/usi-lfkeitel/packet-guardian/src/common"
)

func init() {
	checkers["twilio"] = twilio{}
}

type twilio struct{}

func (t twilio) getInputLabel() string {
	return "Phone Number"
}

func (t twilio) sendCode(e *common.Environment, phone, code string) error {
	accountSid := e.Config.Guest.Twilio.AccountSID
	authToken := e.Config.Guest.Twilio.AuthToken
	from := e.Config.Guest.Twilio.PhoneNumber
	phone, err := formatPhoneNumber(phone)
	if err != nil {
		return err
	}

	client := twirest.NewClient(accountSid, authToken)

	msg := twirest.SendMessage{
		Text: "Verification Code: " + code,
		To:   "+1" + phone,
		From: from}

	resp, err := client.Request(msg)
	if err != nil {
		e.Log.WithField("error", err).Error("Error sending text message")
		return errors.New("Error sending verification code")
	}

	e.Log.Info(resp.Message.Status)
	return nil
}
