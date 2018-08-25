// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"net"
)

type subnet struct {
	allowUnknown bool
	settings     *settings
	net          *net.IPNet
	network      *network
	pools        []*pool
}

func newSubnet() *subnet {
	return &subnet{
		settings: newSettingsBlock(),
	}
}

func (s *subnet) includes(ip net.IP) bool {
	return s.net.Contains(ip)
}
