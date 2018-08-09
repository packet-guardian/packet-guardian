// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"time"
)

type settings struct {
	defaultLeaseTime time.Duration
	maxLeaseTime     time.Duration
	freeLeaseAfter   time.Duration
}

func newSettingsBlock() *settings {
	return &settings{}
}

// mergeSettings will merge s into d.
func mergeSettings(d, s *settings) {
	if d.defaultLeaseTime == 0 {
		d.defaultLeaseTime = s.defaultLeaseTime
	}
	if d.maxLeaseTime == 0 {
		d.maxLeaseTime = s.maxLeaseTime
	}
	if d.freeLeaseAfter == 0 {
		d.freeLeaseAfter = s.freeLeaseAfter
	}
}
