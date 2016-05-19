package dhcp

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/onesimus-systems/dhcp4"
	"github.com/onesimus-systems/packet-guardian/src/common"
)

type Network struct {
	Global               *Global
	Name                 string
	Settings             *Settings
	RegisteredSettings   *Settings
	regOptionsCached     bool
	UnregisteredSettings *Settings
	unregOptionsCached   bool
	Subnets              []*Subnet
}

func newNetwork(name string) *Network {
	return &Network{
		Name:                 name,
		Settings:             newSettingsBlock(),
		RegisteredSettings:   newSettingsBlock(),
		UnregisteredSettings: newSettingsBlock(),
	}
}

// GetLeaseTime returns the lease time given the requested time req and if the client is registered.
// If req is 0 then the default lease time is returned. Otherwise it will return the lower of
// req and the maximum lease time. If the network does not have an explicitly set duration for either,
// it will get the duration from Global.
func (n *Network) GetLeaseTime(req time.Duration, registered bool) time.Duration {
	// TODO: Clean this up
	if registered {
		if req == 0 {
			if n.RegisteredSettings.DefaultLeaseTime != 0 {
				return n.RegisteredSettings.DefaultLeaseTime
			}
			if n.Settings.DefaultLeaseTime != 0 {
				return n.Settings.DefaultLeaseTime
			}
			// Save the result for later
			n.Settings.DefaultLeaseTime = n.Global.GetLeaseTime(req, registered)
			return n.Settings.DefaultLeaseTime
		}

		if n.RegisteredSettings.MaxLeaseTime != 0 {
			if req < n.RegisteredSettings.MaxLeaseTime {
				return req
			}
			return n.RegisteredSettings.MaxLeaseTime
		}
		if n.Settings.MaxLeaseTime != 0 {
			if req < n.Settings.MaxLeaseTime {
				return req
			}
			return n.Settings.MaxLeaseTime
		}

		// Save the result for later
		n.Settings.MaxLeaseTime = n.Global.GetLeaseTime(req, registered)

		if req < n.Settings.MaxLeaseTime {
			return req
		}
		return n.Settings.MaxLeaseTime
	}

	if req == 0 {
		if n.UnregisteredSettings.DefaultLeaseTime != 0 {
			return n.UnregisteredSettings.DefaultLeaseTime
		}
		if n.Settings.DefaultLeaseTime != 0 {
			return n.Settings.DefaultLeaseTime
		}
		// Save the result for later
		n.Settings.DefaultLeaseTime = n.Global.GetLeaseTime(req, registered)
		return n.Settings.DefaultLeaseTime
	}

	if n.UnregisteredSettings.MaxLeaseTime != 0 {
		if req < n.UnregisteredSettings.MaxLeaseTime {
			return req
		}
		return n.UnregisteredSettings.MaxLeaseTime
	}
	if n.Settings.MaxLeaseTime != 0 {
		if req < n.Settings.MaxLeaseTime {
			return req
		}
		return n.Settings.MaxLeaseTime
	}

	// Save the result for later
	n.Settings.MaxLeaseTime = n.Global.GetLeaseTime(req, registered)

	if req < n.Settings.MaxLeaseTime {
		return req
	}
	return n.Settings.MaxLeaseTime
}

func (n *Network) GetOptions(registered bool) dhcp4.Options {
	if registered && n.regOptionsCached {
		return n.RegisteredSettings.Options
	} else if !registered && n.unregOptionsCached {
		return n.UnregisteredSettings.Options
	}

	higher := n.Global.GetOptions(registered)
	if registered {
		// Merge network "global" setting into registered settings
		for c, v := range n.Settings.Options {
			if _, ok := n.RegisteredSettings.Options[c]; !ok {
				n.RegisteredSettings.Options[c] = v
			}
		}
		// Merge Global setting into registered setting
		for c, v := range higher {
			if _, ok := n.RegisteredSettings.Options[c]; !ok {
				n.RegisteredSettings.Options[c] = v
			}
		}
		n.regOptionsCached = true
		return n.RegisteredSettings.Options
	}

	// Merge network "global" setting into unregistered settings
	for c, v := range n.Settings.Options {
		if _, ok := n.UnregisteredSettings.Options[c]; !ok {
			n.UnregisteredSettings.Options[c] = v
		}
	}
	// Merge Global setting into unregistered setting
	for c, v := range higher {
		if _, ok := n.UnregisteredSettings.Options[c]; !ok {
			n.UnregisteredSettings.Options[c] = v
		}
	}
	n.unregOptionsCached = true
	return n.UnregisteredSettings.Options
}

func (n *Network) Includes(ip net.IP) bool {
	for _, s := range n.Subnets {
		if s.Includes(ip) {
			return true
		}
	}
	return false
}

func (n *Network) GetFreeLease(e *common.Environment, registered bool) *Lease {
	for _, s := range n.Subnets {
		if s.AllowUnknown == registered {
			continue
		}
		if l := s.GetFreeLease(e); l != nil {
			return l
		}
	}
	return nil
}

func (n *Network) GetLeaseByMAC(mac net.HardwareAddr, registered bool) *Lease {
	for _, s := range n.Subnets {
		if s.AllowUnknown == registered {
			continue
		}
		for _, p := range s.Pools {
			for _, l := range p.Leases {
				if bytes.Equal(l.MAC, mac) {
					return l
				}
			}
		}
	}
	return nil
}

func (n *Network) GetLeaseByIP(ip net.IP, registered bool) *Lease {
	for _, s := range n.Subnets {
		if s.AllowUnknown == registered {
			continue
		}
		for _, p := range s.Pools {
			if l, ok := p.Leases[ip.String()]; ok {
				return l
			}
		}
	}
	return nil
}

func (n *Network) Print() {
	fmt.Printf("\n---Network Configuration - %s---\n", n.Name)
	fmt.Println("\n--Network Settings--")
	n.Settings.Print()
	fmt.Println("\n--Network Registered Settings--")
	n.RegisteredSettings.Print()
	fmt.Println("\n--Network Unregistered Settings--")
	n.UnregisteredSettings.Print()
	fmt.Println("\n--Subnets in network--")
	for _, s := range n.Subnets {
		s.Print()
	}
}
