// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/onesimus-systems/dhcp4"
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
	// TODO: Clean this up
	if registered {
		if req == 0 {
			if n.registeredSettings.defaultLeaseTime != 0 {
				return n.registeredSettings.defaultLeaseTime
			}
			if n.settings.defaultLeaseTime != 0 {
				return n.settings.defaultLeaseTime
			}
			// Save the result for later
			n.registeredSettings.defaultLeaseTime = n.global.getLeaseTime(req, registered)
			return n.registeredSettings.defaultLeaseTime
		}

		if n.registeredSettings.maxLeaseTime != 0 {
			if req < n.registeredSettings.maxLeaseTime {
				return req
			}
			return n.registeredSettings.maxLeaseTime
		}
		if n.settings.maxLeaseTime != 0 {
			if req < n.settings.maxLeaseTime {
				return req
			}
			return n.settings.maxLeaseTime
		}

		// Save the result for later
		n.registeredSettings.maxLeaseTime = n.global.getLeaseTime(req, registered)

		if req < n.registeredSettings.maxLeaseTime {
			return req
		}
		return n.registeredSettings.maxLeaseTime
	}

	if req == 0 {
		if n.unregisteredSettings.defaultLeaseTime != 0 {
			return n.unregisteredSettings.defaultLeaseTime
		}
		if n.settings.defaultLeaseTime != 0 {
			return n.settings.defaultLeaseTime
		}
		// Save the result for later
		n.unregisteredSettings.defaultLeaseTime = n.global.getLeaseTime(req, registered)
		return n.unregisteredSettings.defaultLeaseTime
	}

	if n.unregisteredSettings.maxLeaseTime != 0 {
		if req < n.unregisteredSettings.maxLeaseTime {
			return req
		}
		return n.unregisteredSettings.maxLeaseTime
	}
	if n.settings.maxLeaseTime != 0 {
		if req < n.settings.maxLeaseTime {
			return req
		}
		return n.settings.maxLeaseTime
	}

	// Save the result for later
	n.unregisteredSettings.maxLeaseTime = n.global.getLeaseTime(req, registered)

	if req < n.unregisteredSettings.maxLeaseTime {
		return req
	}
	return n.unregisteredSettings.maxLeaseTime
}

func (n *network) getOptions(registered bool) dhcp4.Options {
	if registered && n.regOptionsCached {
		return n.registeredSettings.options
	} else if !registered && n.unregOptionsCached {
		return n.unregisteredSettings.options
	}

	higher := n.global.getOptions(registered)
	if registered {
		// Merge network "global" setting into registered settings
		for c, v := range n.settings.options {
			if _, ok := n.registeredSettings.options[c]; !ok {
				n.registeredSettings.options[c] = v
			}
		}
		// Merge Global setting into registered setting
		for c, v := range higher {
			if _, ok := n.registeredSettings.options[c]; !ok {
				n.registeredSettings.options[c] = v
			}
		}
		n.regOptionsCached = true
		return n.registeredSettings.options
	}

	// Merge network "global" setting into unregistered settings
	for c, v := range n.settings.options {
		if _, ok := n.unregisteredSettings.options[c]; !ok {
			n.unregisteredSettings.options[c] = v
		}
	}
	// Merge Global setting into unregistered setting
	for c, v := range higher {
		if _, ok := n.unregisteredSettings.options[c]; !ok {
			n.unregisteredSettings.options[c] = v
		}
	}
	n.unregOptionsCached = true
	return n.unregisteredSettings.options
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
