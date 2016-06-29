// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package guest

import (
	"errors"
	"strings"

	"github.com/usi-lfkeitel/packet-guardian/src/common"
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
