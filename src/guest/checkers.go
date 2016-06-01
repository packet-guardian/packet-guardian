package guest

import (
	"errors"

	"github.com/onesimus-systems/packet-guardian/src/common"
)

type guestChecker interface {
	getInputLabel() string
	sendCode(*common.Environment, string, string) error
}

var checkers = make(map[string]guestChecker)

// GetInputLabel returns the text that is displayed above the textbox asking
// for a guest's credential
func GetInputLabel(e *common.Environment) string {
	f, ok := checkers[e.Config.Guest.Checker]
	if !ok {
		e.Log.Errorf("%s is not a guest checker type", e.Config.Guest.Checker)
		return ""
	}
	return f.getInputLabel()
}

// SendGuestCode will send the verification code using the congifured checker.
func SendGuestCode(e *common.Environment, c, code string) error {
	f, ok := checkers[e.Config.Guest.Checker]
	if !ok {
		e.Log.Errorf("%s is not a guest checker type", e.Config.Guest.Checker)
		return errors.New("Internal server error")
	}
	return f.sendCode(e, c, code)
}
