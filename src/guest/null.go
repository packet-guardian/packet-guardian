// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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