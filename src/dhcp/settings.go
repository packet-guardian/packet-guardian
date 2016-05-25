// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"fmt"
	"time"

	"github.com/onesimus-systems/dhcp4"
)

type Settings struct {
	Options          map[dhcp4.OptionCode][]byte
	DefaultLeaseTime time.Duration
	MaxLeaseTime     time.Duration
}

func newSettingsBlock() *Settings {
	return &Settings{
		Options: make(map[dhcp4.OptionCode][]byte),
	}
}

func (s *Settings) Print() {
	fmt.Printf("Default Lease Time: %s\n", s.DefaultLeaseTime.String())
	fmt.Printf("Max Lease Time: %s\n", s.MaxLeaseTime.String())
	fmt.Println("-DHCP Options-")
	for c, v := range s.Options {
		fmt.Printf("%s: %v\n", c.String(), v)
	}
}
