package guest

import "github.com/onesimus-systems/packet-guardian/src/common"

func init() {
	checkers["email"] = email{}
}

type email struct{}

func (t email) getInputLabel() string {
	return "Email Address"
}

func (t email) sendCode(e *common.Environment, phone, code string) error {
	return nil
}
