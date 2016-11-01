// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"time"

	"github.com/onesimus-systems/dhcp4"
)

var r = regexp.MustCompile(`\d+ bytes from .*`)

type pool struct {
	rangeStart    net.IP
	rangeEnd      net.IP
	settings      *settings
	optionsCached bool
	leases        map[string]*Lease // IP -> Lease
	subnet        *subnet
	nextFreeStart int
	ipsInPool     int
}

func newPool() *pool {
	return &pool{
		settings: newSettingsBlock(),
		leases:   make(map[string]*Lease),
	}
}

func (p *pool) getCountOfIPs() int {
	if p.ipsInPool == 0 {
		p.ipsInPool = dhcp4.IPRange(p.rangeStart, p.rangeEnd)
	}
	return p.ipsInPool
}

// getLeaseTime returns the lease time given the requested time req and if the client is registered.
// If req is 0 then the default lease time is returned. Otherwise it will return the lower of
// req and the maximum lease time. If the pool does not have an explicitly set duration for either,
// it will get the duration from its subnet.
func (p *pool) getLeaseTime(req time.Duration, registered bool) time.Duration {
	if req == 0 {
		if p.settings.defaultLeaseTime > 0 {
			return p.settings.defaultLeaseTime
		}
		// Save the result for later
		p.settings.defaultLeaseTime = p.subnet.getLeaseTime(req, registered)
		return p.settings.defaultLeaseTime
	}

	if p.settings.maxLeaseTime > 0 {
		if req < p.settings.maxLeaseTime {
			return req
		}
		return p.settings.maxLeaseTime
	}

	// Save the result for later
	p.settings.maxLeaseTime = p.subnet.getLeaseTime(req, registered)

	if req <= p.settings.maxLeaseTime {
		return req
	}
	return p.settings.maxLeaseTime
}

func (p *pool) getOptions(registered bool) dhcp4.Options {
	if p.optionsCached {
		return p.settings.options
	}

	higher := p.subnet.getOptions(registered)
	for c, v := range higher {
		if _, ok := p.settings.options[c]; !ok {
			p.settings.options[c] = v
		}
	}
	p.optionsCached = true
	return p.settings.options
}

func (p *pool) getFreeLease(s *ServerConfig) *Lease {
	now := time.Now()

	regFreeTime := p.subnet.network.global.registeredSettings.freeLeaseAfter
	unRegFreeTime := p.subnet.network.global.unregisteredSettings.freeLeaseAfter
	// Find a candidate from the already used leases
	for _, l := range p.leases {
		if l.IsAbandoned { // IP in use by a device we don't know about
			continue
		}
		if l.End.After(now) { // Active lease
			continue
		}
		if l.Offered && now.After(l.End) { // Lease was offered but not taken
			l.Offered = false
			return l
		}
		if !l.Registered && l.End.Add(unRegFreeTime).Before(now) { // Unregisted lease expired
			return l
		}
		if l.Registered && l.End.Add(regFreeTime).Before(now) { // Registered lease expired
			return l
		}
	}

	// No candidates, find the next available lease
	for i := p.nextFreeStart; i < p.getCountOfIPs(); i++ {
		next := dhcp4.IPAdd(p.rangeStart, i)
		p.nextFreeStart = i + 1

		// Check if IP has a lease
		// Sanity check
		_, ok := p.leases[next.String()]
		if ok {
			continue
		}

		// IP has no lease with it
		l := NewLease(s.LeaseStore)
		// All known leases have already been checked, which means if this IP
		// is in use, we didn't do it. Mark as abandoned.
		if !s.IsTesting() && isIPInUse(next) {
			s.Log.Error("Abandoned IP %s", next.String())
			l.IsAbandoned = true
			continue
		}

		// Set IP and pool, add to leases map, return
		l.IP = next
		l.Network = p.subnet.network.name
		l.Registered = !p.subnet.allowUnknown
		p.leases[next.String()] = l
		return l
	}

	// We've exhausted all possibilities, admit defeat.
	return nil
}

func (p *pool) getFreeLeaseDesperate(s *ServerConfig) *Lease {
	now := time.Now()

	// No free leases, bring out the big guns
	// Find the oldest expired lease
	var longestExpiredLease *Lease
	for _, l := range p.leases {
		if l.End.After(now) { // Skip active leases
			continue
		}

		if longestExpiredLease == nil {
			longestExpiredLease = l
			continue
		}

		if l.End.Before(longestExpiredLease.End) {
			longestExpiredLease = l
		}
	}

	if longestExpiredLease != nil {
		return longestExpiredLease
	}

	// Now we're getting desperate
	// Check abandoned leases for availability
	for _, l := range p.leases {
		if !l.IsAbandoned { // Skip non-abandoned leases
			continue
		}
		if !isIPInUse(l.IP) {
			l.IsAbandoned = false
			return l
		}
	}
	return nil
}

func (p *pool) includes(ip net.IP) bool {
	return dhcp4.IPInRange(p.rangeStart, p.rangeEnd, ip)
}

func (p *pool) print() {
	fmt.Printf("\n---Pool %s - %s---\n", p.rangeStart.String(), p.rangeEnd.String())
	fmt.Println("Pool settings")
	p.settings.Print()
}

// isIPInUse will use the system ping utility to determine if an IP is in use.
// At the moment this is Linux specific. I need to find a more cross platform
// method to do ICMP probes. Right now abandonment checks will be disabled in
// Windows machines.
func isIPInUse(host net.IP) bool {
	count := "-c"
	wait := "1"
	if runtime.GOOS == "windows" {
		count = "-n"
		wait = "500"
	}

	// -c/-n: packet count, -w: timeout in seconds
	out, err := exec.Command("ping", count, "1", "-w", wait, host.String()).Output()
	if err != nil {
		return false
	}
	return (r.Find(out) != nil)
}

func (p *pool) printLeases() {
	for _, l := range p.leases {
		fmt.Printf("%+v\n", l)
	}
}
