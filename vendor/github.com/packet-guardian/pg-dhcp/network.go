// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"
)

type network struct {
	global               *global
	name                 string
	settings             *settings
	registeredSettings   *settings
	regOptionsCached     bool
	unregisteredSettings *settings
	unregOptionsCached   bool
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

// GetLeaseTime returns the lease time given the requested time req and if the client is registered.
// If req is 0 then the default lease time is returned. Otherwise it will return the lower of
// req and the maximum lease time. If the network does not have an explicitly set duration for either,
// it will get the duration from Global.
func (n *network) getLeaseTime(req time.Duration, registered bool) time.Duration {
	if req == 0 {
		return n.getDefaultLeaseTime(registered)
	}
	return n.getMaxLeaseTime(req, registered)
}

func (n *network) getDefaultLeaseTime(registered bool) time.Duration {
	if registered {
		if n.registeredSettings.defaultLeaseTime > 0 {
			return n.registeredSettings.defaultLeaseTime
		}
		if n.settings.defaultLeaseTime > 0 {
			return n.settings.defaultLeaseTime
		}
		// Save to return early next time
		n.registeredSettings.defaultLeaseTime = n.global.getLeaseTime(0, registered)
		return n.registeredSettings.defaultLeaseTime
	}

	if n.unregisteredSettings.defaultLeaseTime > 0 {
		return n.unregisteredSettings.defaultLeaseTime
	}
	if n.settings.defaultLeaseTime > 0 {
		return n.settings.defaultLeaseTime
	}
	// Save to return early next time
	n.unregisteredSettings.defaultLeaseTime = n.global.getLeaseTime(0, registered)
	return n.unregisteredSettings.defaultLeaseTime
}

func (n *network) getMaxLeaseTime(req time.Duration, registered bool) time.Duration {
	// Registered devices
	if registered {
		if n.registeredSettings.maxLeaseTime > 0 {
			if req <= n.registeredSettings.maxLeaseTime {
				return req
			}
			return n.registeredSettings.maxLeaseTime
		}
		if n.settings.maxLeaseTime > 0 {
			if req <= n.settings.maxLeaseTime {
				return req
			}
			return n.settings.maxLeaseTime
		}
		return n.global.getLeaseTime(req, registered)
	}

	// Unregistered devices
	if n.unregisteredSettings.maxLeaseTime > 0 {
		if req <= n.unregisteredSettings.maxLeaseTime {
			return req
		}
		return n.unregisteredSettings.maxLeaseTime
	}
	if n.settings.maxLeaseTime > 0 {
		if req <= n.settings.maxLeaseTime {
			return req
		}
		return n.settings.maxLeaseTime
	}
	return n.global.getLeaseTime(req, registered)
}

func (n *network) getSettings(registered bool) *settings {
	if registered && n.regOptionsCached {
		return n.registeredSettings
	} else if !registered && n.unregOptionsCached {
		return n.unregisteredSettings
	}

	gSet := n.global.getSettings(registered)
	if registered {
		mergeSettings(n.registeredSettings, gSet)
		n.regOptionsCached = true
		return n.registeredSettings
	}

	mergeSettings(n.unregisteredSettings, gSet)
	n.unregOptionsCached = true
	return n.unregisteredSettings
}

func (n *network) includes(ip net.IP) bool {
	for _, s := range n.subnets {
		if s.includes(ip) {
			return true
		}
	}
	return false
}

func (n *network) getFreeLease(e *ServerConfig, registered bool) (*Lease, *pool) {
	for _, s := range n.subnets {
		if s.allowUnknown == registered {
			continue
		}
		for _, p := range s.pools {
			if l := p.getFreeLease(e); l != nil {
				return l, p
			}
		}
	}
	return nil, nil
}

func (n *network) getFreeLeaseDesperate(e *ServerConfig, registered bool) (*Lease, *pool) {
	for _, s := range n.subnets {
		if s.allowUnknown == registered {
			continue
		}
		for _, p := range s.pools {
			if l := p.getFreeLeaseDesperate(e); l != nil {
				return l, p
			}
		}
	}
	return nil, nil
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

func (n *network) print() {
	fmt.Printf("\n---Network Configuration - %s---\n", n.name)
	fmt.Println("\n--Network Settings--")
	n.settings.Print()
	fmt.Println("\n--Network Registered Settings--")
	n.registeredSettings.Print()
	fmt.Println("\n--Network Unregistered Settings--")
	n.unregisteredSettings.Print()
	fmt.Println("\n--Subnets in network--")
	for _, s := range n.subnets {
		s.print()
	}
}
