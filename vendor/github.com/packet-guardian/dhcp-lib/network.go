// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"bytes"
	"net"
	"strings"
)

type network struct {
	global               *global
	name                 string
	settings             *settings
	registeredSettings   *settings
	unregisteredSettings *settings
	subnets              []*subnet
}

func newNetwork(name string) *network {
	return &network{
		name:                 strings.ToLower(name),
		settings:             newSettingsBlock(),
		registeredSettings:   newSettingsBlock(),
		unregisteredSettings: newSettingsBlock(),
	}
}

func (n *network) includes(ip net.IP) bool {
	for _, s := range n.subnets {
		if s.includes(ip) {
			return true
		}
	}
	return false
}

func (n *network) getPoolOfIP(ip net.IP) *pool {
	for _, s := range n.subnets {
		for _, p := range s.pools {
			if p.includes(ip) {
				return p
			}
		}
	}
	return nil
}

func (n *network) getLeaseByMAC(mac net.HardwareAddr, registered bool) (*Lease, *pool) {
	for _, s := range n.subnets {
		if s.allowUnknown == registered {
			continue
		}
		for _, p := range s.pools {
			for _, l := range p.leases {
				if bytes.Equal(l.MAC, mac) {
					return l, p
				}
			}
		}
	}
	return nil, nil
}

func (n *network) getLeaseByIP(ip net.IP, registered bool) (*Lease, *pool) {
	for _, s := range n.subnets {
		if s.allowUnknown == registered {
			continue
		}
		for _, p := range s.pools {
			if l, ok := p.leases[ip.String()]; ok {
				return l, p
			}
		}
	}
	return nil, nil
}
