// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"fmt"
	"net"
	"time"

	"github.com/onesimus-systems/dhcp4"
	"github.com/onesimus-systems/packet-guardian/src/common"
)

type subnet struct {
	allowUnknown  bool
	settings      *settings
	optionsCached bool
	net           *net.IPNet
	network       *network
	pools         []*pool
}

func newSubnet() *subnet {
	return &subnet{
		settings: newSettingsBlock(),
	}
}

// getLeaseTime returns the lease time given the requested time req and if the client is registered.
// If req is 0 then the default lease time is returned. Otherwise it will return the lower of
// req and the maximum lease time. If the subnet does not have an explicitly set duration for either,
// it will get the duration from its Network.
func (s *subnet) getLeaseTime(req time.Duration, registered bool) time.Duration {
	if req == 0 {
		if s.settings.defaultLeaseTime != 0 {
			return s.settings.defaultLeaseTime
		}
		// Save the result for later
		s.settings.defaultLeaseTime = s.network.getLeaseTime(req, registered)
		return s.settings.defaultLeaseTime
	}

	if s.settings.maxLeaseTime != 0 {
		if req < s.settings.maxLeaseTime {
			return req
		}
		return s.settings.maxLeaseTime
	}

	// Save the result for later
	s.settings.maxLeaseTime = s.network.getLeaseTime(req, registered)

	if req < s.settings.maxLeaseTime {
		return req
	}
	return s.settings.maxLeaseTime
}

func (s *subnet) getOptions(registered bool) dhcp4.Options {
	if s.optionsCached {
		return s.settings.options
	}

	higher := s.network.getOptions(registered)
	for c, v := range higher {
		if _, ok := s.settings.options[c]; !ok {
			s.settings.options[c] = v
		}
	}
	s.optionsCached = true
	return s.settings.options
}

func (s *subnet) includes(ip net.IP) bool {
	return s.net.Contains(ip)
}

func (s *subnet) getFreeLease(e *common.Environment) *Lease {
	for _, p := range s.pools {
		if l := p.getFreeLease(e); l != nil {
			return l
		}
	}
	return nil
}

func (s *subnet) print() {
	fmt.Printf("\n---Subnet - %s---\n", s.net.String())
	fmt.Printf("Registered: %t\n", !s.allowUnknown)
	fmt.Println("Subnet Settings")
	s.settings.Print()
	fmt.Println("\n--Subnet Pools--")
	for _, p := range s.pools {
		p.print()
	}
}
