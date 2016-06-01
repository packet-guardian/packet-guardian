package guest

import "github.com/onesimus-systems/packet-guardian/src/common"

type guestChecker func(*common.Environment, string, string) error

var checkers = make(map[string]guestChecker)

func SendGuestCode(e *common.Environment, c, code string) error {
	f, ok := checkers[e.Config.Guest.Checker]
	if !ok {
		e.Log.Errorf("%s is not a guest checker type", e.Config.Guest.Checker)
	}
	return f(e, c, code)
}
