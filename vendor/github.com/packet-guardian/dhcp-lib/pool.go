// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"net"
	"regexp"
)

var r = regexp.MustCompile(`\d+ bytes from .*`)

type pool struct {
	rangeStart net.IP
	rangeEnd   net.IP
	settings   *settings
	leases     map[string]*Lease // IP -> Lease
	subnet     *subnet
	ipsInPool  int
}

func newPool() *pool {
	return &pool{
		settings: newSettingsBlock(),
		leases:   make(map[string]*Lease),
	}
}

func (p *pool) getCountOfIPs() int {
	if p.ipsInPool == 0 {
		p.ipsInPool = IPRange(p.rangeStart, p.rangeEnd)
	}
	return p.ipsInPool
}

func (p *pool) includes(ip net.IP) bool {
	return IPInRange(p.rangeStart, p.rangeEnd, ip)
}
