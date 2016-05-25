// This source file is part of the Packet Guardian project.
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

	"github.com/lfkeitel/verbose"
	"github.com/onesimus-systems/dhcp4"
	"github.com/onesimus-systems/packet-guardian/src/common"
)

var r = regexp.MustCompile(`\d+ bytes from .*`)

type Pool struct {
	RangeStart    net.IP
	RangeEnd      net.IP
	Settings      *Settings
	optionsCached bool
	Leases        map[string]*Lease // IP -> Lease
	Subnet        *Subnet
	nextFreeStart int
	ipsInPool     int
}

func newPool() *Pool {
	return &Pool{
		Settings: newSettingsBlock(),
		Leases:   make(map[string]*Lease),
	}
}

func (p *Pool) GetCountOfIPs() int {
	if p.ipsInPool == 0 {
		p.ipsInPool = dhcp4.IPRange(p.RangeStart, p.RangeEnd)
	}
	return p.ipsInPool
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
	p.optionsCached = true
	return p.Settings.Options
}

func (p *Pool) GetFreeLease(e *common.Environment) *Lease {
	now := time.Now()

	// Find an expired current lease over a week old
	weekInSecs := time.Duration(604800) * time.Second
	// Find an unregistered lease over two hours old
	twoHours := time.Duration(2) * time.Hour
	for _, l := range p.Leases {
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
		if !l.Registered && l.End.Add(twoHours).Before(now) { // Unregisted lease expired two hours ago
			return l
		}
		if l.Registered && l.End.Add(weekInSecs).Before(now) { // Lease expired a week ago
			return l
		}
	}

	// No candidates, find the next available lease
	for i := p.nextFreeStart; i < p.GetCountOfIPs(); i++ {
		next := dhcp4.IPAdd(p.RangeStart, i)
		// Check if IP has a lease
		_, ok := p.Leases[next.String()]
		if ok {
			continue
		}

		// IP has no lease with it
		l := NewLease(e)
		// All known leases have already been checked, which means if this IP
		// is in use, we didn't do it. Mark as abandoned.
		if !e.IsTesting() && isIPInUse(next) {
			e.Log.WithFields(verbose.Fields{
				"IP": next.String(),
			}).Notice("Abandoned IP")
			l.IsAbandoned = true
			continue
		}

		// Set IP and pool, add to leases map, return
		l.IP = next
		l.Network = p.Subnet.Network.Name
		l.Pool = p
		l.Registered = !p.Subnet.AllowUnknown
		p.Leases[next.String()] = l
		p.nextFreeStart = i + 1
		return l
	}

	// No free leases, bring out the big guns
	// Find the oldest expired lease
	var longestExpiredLease *Lease
	for _, l := range p.Leases {
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
	for _, l := range p.Leases {
		if !l.IsAbandoned { // Skip non-abandoned leases
			continue
		}
		if !isIPInUse(l.IP) {
			return l
		}
	}

	// We've exhausted all possibilities, admit defeat.
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

// isIPInUse will use the system ping utility to determine if an IP is in use.
// At the moment this is Linux specific. I need to find a more cross platform
// method to do ICMP probes. Right now abandonment checks will be disabled in
// Windows machines.
func isIPInUse(host net.IP) bool {
	count := "-c"
	wait := "2"
	if runtime.GOOS == "windows" {
		count = "-n"
		wait = "2000"
	}

	// -c/-n: packet count, -w: timeout in seconds
	out, err := exec.Command("ping", count, "1", "-w", wait, host.String()).Output()
	if err != nil {
		return false
	}
	return (r.Find(out) != nil)
}

func (p *Pool) PrintLeases() {
	for _, l := range p.Leases {
		fmt.Printf("%+v\n", l)
	}
}
