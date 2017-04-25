// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package guest

import (
	"errors"
	"strings"

	"github.com/lfkeitel/verbose"
	"github.com/packet-guardian/packet-guardian/src/common"
)

type guestChecker interface {
	getInputLabel() string             // Show about the input textbox
	getInputText() string              // Show next to the label for explanations
	getVerificationText() string       // Show on the verification screen, again just for explanations
	normalizeCredential(string) string // Normalize a credential to be used in the rest of the application
	sendCode(*common.Environment, string, string) error
}

var checkers = make(map[string]guestChecker)

// GetInputLabel returns the text that is displayed above the textbox asking
// for a guest's credential
func GetInputLabel(e *common.Environment) string {
	f, ok := checkers[e.Config.Guest.Checker]
	if !ok {
		e.Log.WithFields(verbose.Fields{
			"package": "guest",
			"checker": e.Config.Guest.Checker,
		}).Error("Invalid guest checker")
		return ""
	}
	return f.getInputLabel()
}

func GetInputText(e *common.Environment) string {
	f, ok := checkers[e.Config.Guest.Checker]
	if !ok {
		e.Log.WithFields(verbose.Fields{
			"package": "guest",
			"checker": e.Config.Guest.Checker,
		}).Error("Invalid guest checker")
		return ""
	}
	return f.getInputText()
}

func GetVerificationText(e *common.Environment) string {
	f, ok := checkers[e.Config.Guest.Checker]
	if !ok {
		e.Log.WithFields(verbose.Fields{
			"package": "guest",
			"checker": e.Config.Guest.Checker,
		}).Error("Invalid guest checker")
		return ""
	}
	return f.getVerificationText()
}

func NormalizeCredential(e *common.Environment, c string) string {
	f, ok := checkers[e.Config.Guest.Checker]
	if !ok {
		e.Log.WithFields(verbose.Fields{
			"package": "guest",
			"checker": e.Config.Guest.Checker,
		}).Error("Invalid guest checker")
		return ""
	}
	return f.normalizeCredential(c)
}

// SendGuestCode will send the verification code using the congifured checker.
func SendGuestCode(e *common.Environment, c, code string) error {
	f, ok := checkers[e.Config.Guest.Checker]
	if !ok {
		e.Log.WithFields(verbose.Fields{
			"package": "guest",
			"checker": e.Config.Guest.Checker,
		}).Error("Invalid guest checker")
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
		// Strip country code
		return number[1:], nil
	}
	if len(number) != 10 {
		return "", errors.New("Invalid phone number")
	}
	return number, nil
}

func stripChars(str, chr string) string {
	return strings.Map(func(r rune) rune {
		if !strings.ContainsRune(chr, r) {
			return -1
		}
		return r
	}, str)
}
