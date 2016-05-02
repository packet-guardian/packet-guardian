package dhcp

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"time"

	"github.com/dragonrider23/verbose"
	"github.com/onesimus-systems/dhcp4"
	"github.com/onesimus-systems/packet-guardian/src/common"
)

type Pool struct {
	RangeStart    net.IP
	RangeEnd      net.IP
	Settings      *Settings
	optionsCached bool
	Leases        map[string]*Lease // IP -> Lease
	Subnet        *Subnet
	nextStart     int
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
		if l.End.Add(weekInSecs).Before(now) { // Lease expired a week ago
			return l
		}
	}

	// No candidates, find the next available lease
	ipsInPool := dhcp4.IPRange(p.RangeStart, p.RangeEnd)
	for i := p.nextStart; i < ipsInPool; i++ {
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
		if isIPInUse(next) {
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
		p.Leases[next.String()] = l
		p.nextStart = i + 1
		return l
	}

	// No free leases, no existing leases, check Abandoned leases
	// TODO: Create abandoned recollection
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
	if runtime.GOOS == "windows" {
		return false
	}

	six := ""
	if host.To4() == nil {
		six = "6"
	}
	// -c: packet count, -w: timeout in seconds
	out, err := exec.Command("ping"+six, "-c", "1", "-w", "3", "--", host.String()).Output()
	if err != nil {
		return false
	}
	r, _ := regexp.Compile(`\d+ bytes from .*`)
	line := r.Find(out)
	return (line != nil)
}

func (p *Pool) PrintLeases() {
	for _, l := range p.Leases {
		fmt.Printf("%+v\n", l)
	}
}
