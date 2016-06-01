package guest

import (
	"errors"
	"strings"

	"bitbucket.org/ckvist/twilio/twirest"
	"github.com/onesimus-systems/packet-guardian/src/common"
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
		e.Log.Errorf("Error sending text message: %v", err)
		return errors.New("Error sending verification code")
	}

	e.Log.Info(resp.Message.Status)
	return nil
}

func formatPhoneNumber(number string) (string, error) {
	number = stripChars(number, "0123456789")
	if len(number) == 11 {
		if number[0] != '1' {
			return "", errors.New("Only US phone numbers are supported")
		}
		return number[1:], nil
	}
	if len(number) != 10 {
		return "", errors.New("Invalid phone number")
	}
	return number, nil
}

func stripChars(str, chr string) string {
	return strings.Map(func(r rune) rune {
		if strings.IndexRune(chr, r) < 0 {
			return -1
		}
		return r
	}, str)
}
