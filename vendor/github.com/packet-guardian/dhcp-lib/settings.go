// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"time"

	dhcp4 "github.com/packet-guardian/pg-dhcp/dhcp"
)

type settings struct {
	options          dhcp4.Options
	vendorOptions    dhcp4.Options
	defaultLeaseTime time.Duration
	maxLeaseTime     time.Duration
	freeLeaseAfter   time.Duration
}

func newSettingsBlock() *settings {
	return &settings{
		options:       make(dhcp4.Options),
		vendorOptions: make(dhcp4.Options),
	}
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

	for c, v := range s.options {
		if _, ok := d.options[c]; !ok {
			d.options[c] = v
		}
	}
}

func (s *settings) genVendorOption() []byte {
	length := 0

	for _, vd := range s.vendorOptions {
		length += 2 + len(vd) // 2 for code and data length
	}

	// I can't find in the RFCs exactly if this option
	// can be send over multiple CLV segments.
	if length == 0 || length > 255 {
		return nil
	}

	data := make([]byte, 0, length)
	for c, vd := range s.vendorOptions {
		vdlen := byte(len(vd))
		data = append(data, byte(c))
		data = append(data, vdlen)
		data = append(data, vd...)
	}

	return data
}
