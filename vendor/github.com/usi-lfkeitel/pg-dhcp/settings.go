// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"fmt"
	"time"

	"github.com/onesimus-systems/dhcp4"
)

type settings struct {
	options          map[dhcp4.OptionCode][]byte
	defaultLeaseTime time.Duration
	maxLeaseTime     time.Duration
	freeLeaseAfter   time.Duration
}

func newSettingsBlock() *settings {
	return &settings{
		options: make(map[dhcp4.OptionCode][]byte),
	}
}

func (s *settings) Print() {
	fmt.Printf("Default Lease Time: %s\n", s.defaultLeaseTime.String())
	fmt.Printf("Max Lease Time: %s\n", s.maxLeaseTime.String())
	fmt.Printf("Free Lease After: %s \n\n", s.freeLeaseAfter.String())
	fmt.Println("-DHCP Options-")
	for c, v := range s.options {
		fmt.Printf("%s: %v\n", c.String(), v)
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
