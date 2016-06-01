package guest

import "github.com/onesimus-systems/packet-guardian/src/common"

func init() {
	checkers["null"] = null{}
}

type null struct{}

func (t null) getInputLabel() string {
	return "Anything"
}

func (t null) sendCode(e *common.Environment, phone, code string) error {
	return nil
}
