package dhcp

import (
	"fmt"
	"net"
	"time"

	"github.com/krolaw/dhcp4"
	"github.com/onesimus-systems/packet-guardian/src/common"
)

type Pool struct {
	RangeStart    net.IP
	RangeEnd      net.IP
	Settings      *Settings
	optionsCached bool
	Leases        map[string]*Lease // IP -> Lease
	Subnet        *Subnet
}

func newPool() *Pool {
	return &Pool{
		Settings: newSettingsBlock(),
		Leases:   make(map[string]*Lease),
	}
}

// GetLeaseTime returns the lease time given the requested time req and if the client is registered.
// If req is 0 then the default lease time is returned. Otherwise it will return the lower of
// req and the maximum lease time. If the pool does not have an explicitly set duration for either,
// it will get the duration from its Subnet.
func (p *Pool) GetLeaseTime(req time.Duration, registered bool) time.Duration {
	if req == 0 {
		if p.Settings.DefaultLeaseTime != 0 {
			return p.Settings.DefaultLeaseTime
		}
		// Save the result for later
		p.Settings.DefaultLeaseTime = p.Subnet.GetLeaseTime(req, registered)
		return p.Settings.DefaultLeaseTime
	}

	if p.Settings.MaxLeaseTime != 0 {
		if req < p.Settings.MaxLeaseTime {
			return req
		}
		return p.Settings.MaxLeaseTime
	}

	// Save the result for later
	p.Settings.MaxLeaseTime = p.Subnet.GetLeaseTime(req, registered)

	if req < p.Settings.MaxLeaseTime {
		return req
	}
	return p.Settings.MaxLeaseTime
}

func (p *Pool) GetOptions(registered bool) dhcp4.Options {
	if p.optionsCached {
		return p.Settings.Options
	}

	higher := p.Subnet.GetOptions(registered)
	for c, v := range higher {
		if _, ok := p.Settings.Options[c]; !ok {
			p.Settings.Options[c] = v
		}
	}
	return p.Settings.Options
}

func (p *Pool) GetFreeLease(e *common.Environment) *Lease {
	now := time.Now()

	// Find an expired current lease over a week old
	weekInSecs := time.Duration(604800) * time.Second
	for _, l := range p.Leases {
		if l.End.After(now) { // Active lease
			continue
		}
		if l.Offered && now.After(l.End) {
			l.Offered = false
			return l
		}
		if l.End.Add(weekInSecs).Before(now) { // Lease expired a week ago
			return l
		}
	}

	// No candidates, find the next available lease
	ipsInPool := dhcp4.IPRange(p.RangeStart, p.RangeEnd)
	for i := 0; i < ipsInPool; i++ {
		next := dhcp4.IPAdd(p.RangeStart, i)
		_, ok := p.Leases[next.String()]
		if ok {
			continue
		}

		l := NewLease(e)
		l.IP = next
		l.Pool = p
		p.Leases[next.String()] = l
		return l
	}
	return nil
}

func (p *Pool) Includes(ip net.IP) bool {
	return dhcp4.IPInRange(p.RangeStart, p.RangeEnd, ip)
}

func (p *Pool) Print() {
	fmt.Printf("\n---Pool %s - %s---\n", p.RangeStart.String(), p.RangeEnd.String())
	fmt.Println("Pool Settings")
	p.Settings.Print()
}
