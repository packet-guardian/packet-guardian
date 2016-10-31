// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package guest

import "github.com/usi-lfkeitel/packet-guardian/src/common"

func init() {
	checkers["null"] = null{}
}

type null struct{}

func (t null) getInputLabel() string {
	return "Anything"
}

func (t null) getInputText() string {
	return ""
}

func (t null) getVerificationText() string {
	return "Look in the log file"
}

func (t null) normalizeCredential(c string) string {
	return c
}

func (t null) sendCode(e *common.Environment, phone, code string) error {
	return nil
}
