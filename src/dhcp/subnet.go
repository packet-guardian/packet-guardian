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

type Subnet struct {
	AllowUnknown  bool
	Settings      *Settings
	optionsCached bool
	Net           *net.IPNet
	Network       *Network
	Pools         []*Pool
}

func newSubnet() *Subnet {
	return &Subnet{
		Settings: newSettingsBlock(),
	}
}

// GetLeaseTime returns the lease time given the requested time req and if the client is registered.
// If req is 0 then the default lease time is returned. Otherwise it will return the lower of
// req and the maximum lease time. If the subnet does not have an explicitly set duration for either,
// it will get the duration from its Network.
func (s *Subnet) GetLeaseTime(req time.Duration, registered bool) time.Duration {
	if req == 0 {
		if s.Settings.DefaultLeaseTime != 0 {
			return s.Settings.DefaultLeaseTime
		}
		// Save the result for later
		s.Settings.DefaultLeaseTime = s.Network.GetLeaseTime(req, registered)
		return s.Settings.DefaultLeaseTime
	}

	if s.Settings.MaxLeaseTime != 0 {
		if req < s.Settings.MaxLeaseTime {
			return req
		}
		return s.Settings.MaxLeaseTime
	}

	// Save the result for later
	s.Settings.MaxLeaseTime = s.Network.GetLeaseTime(req, registered)

	if req < s.Settings.MaxLeaseTime {
		return req
	}
	return s.Settings.MaxLeaseTime
}

func (s *Subnet) GetOptions(registered bool) dhcp4.Options {
	if s.optionsCached {
		return s.Settings.Options
	}

	higher := s.Network.GetOptions(registered)
	for c, v := range higher {
		if _, ok := s.Settings.Options[c]; !ok {
			s.Settings.Options[c] = v
		}
	}
	s.optionsCached = true
	return s.Settings.Options
}

func (s *Subnet) Includes(ip net.IP) bool {
	return s.Net.Contains(ip)
}

func (s *Subnet) GetFreeLease(e *common.Environment) *Lease {
	for _, p := range s.Pools {
		if l := p.GetFreeLease(e); l != nil {
			return l
		}
	}
	return nil
}

func (s *Subnet) Print() {
	fmt.Printf("\n---Subnet - %s---\n", s.Net.String())
	fmt.Printf("Registered: %t\n", !s.AllowUnknown)
	fmt.Println("Subnet Settings")
	s.Settings.Print()
	fmt.Println("\n--Subnet Pools--")
	for _, p := range s.Pools {
		p.Print()
	}
}
