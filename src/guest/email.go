// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
