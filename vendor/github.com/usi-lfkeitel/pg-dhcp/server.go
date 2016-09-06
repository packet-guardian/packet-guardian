// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// TODO: Clean up the handler functions. There's a lot of duplicated code that
// could be extracted to a function.

package dhcp

import (
	"bytes"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/lfkeitel/verbose"
	"github.com/onesimus-systems/dhcp4"
)

var (
	c *Config
)

// A Handler processes all incoming DHCP packets.
type Handler struct {
	gatewayCache map[string]*network
	gatewayMutex sync.Mutex
	c            *ServerConfig
	ro           bool
}

// NewDHCPServer creates and sets up a new DHCP Handler with the give configuration.
func NewDHCPServer(conf *Config, s *ServerConfig) *Handler {
	startLogger(s)
	c = conf
	return &Handler{
		c:            s,
		gatewayCache: make(map[string]*network),
		gatewayMutex: sync.Mutex{},
	}
}

func startLogger(c *ServerConfig) {
	logger := verbose.New("DHCP")
	c.Log = logger
	if c.LogPath == "" {
		// 	if c.IsTesting() {
		// 		sh := verbose.NewStdoutHandler()
		// 		sh.SetMinLevel(verbose.LogLevelDebug)
		// 		sh.SetMaxLevel(verbose.LogLevelFatal)
		// 		logger.AddHandler("stdout", sh)
		// 	}
		return
	}

	sh := verbose.NewStdoutHandler()
	fh, _ := verbose.NewFileHandler(c.LogPath)
	logger.AddHandler("stdout", sh)
	logger.AddHandler("file", fh)

	sh.SetMinLevel(verbose.LogLevelInfo)
	fh.SetMinLevel(verbose.LogLevelInfo)
	// The verbose package sets the default max to Emergancy
	sh.SetMaxLevel(verbose.LogLevelFatal)
	fh.SetMaxLevel(verbose.LogLevelFatal)
}

// ListenAndServe starts the DHCP Handler listening on port 67 for packets.
// This is blocking like HTTP's ListenAndServe method.
func (h *Handler) ListenAndServe() error {
	h.c.Log.Info("Starting DHCP server...")
	return dhcp4.ListenAndServe(h)
}

// Readonly will force the DHCP Handler into a mode where it will process a request
// as usual but will never return anything to the client. This can be used to perhaps
// test if traffic is getting to the server without actually doing anything.
func (h *Handler) Readonly() {
	h.ro = true
}

// Respond disables readonly mode.
func (h *Handler) Respond() {
	h.ro = false
}

// LoadLeases will import any current leases saved to the database.
func (h *Handler) LoadLeases() error {
	// Get leases from storage
	leases, err := h.c.LeaseStore.GetAllLeases()
	if err != nil {
		return err
	}

	if leases == nil {
		return nil
	}

	// Find the pool each lease belongs to
	for _, lease := range leases {
		n, ok := c.networks[lease.Network]
		if !ok {
			continue
		}
	subnetLoop:
		for _, subnet := range n.subnets {
			if !subnet.includes(lease.IP) {
				continue
			}
			for _, pool := range subnet.pools {
				if !pool.includes(lease.IP) {
					continue
				}
				pool.leases[lease.IP.String()] = lease
				h.c.Log.WithField("Address", lease.IP).Debug("Loaded lease")
				break subnetLoop
			}
		}
	}
	return nil
}

// ServeDHCP processes an incoming DHCP packet and returns a response.
func (h *Handler) ServeDHCP(p dhcp4.Packet, msgType dhcp4.MessageType, options dhcp4.Options) dhcp4.Packet {
	defer func() {
		if r := recover(); r != nil {
			h.c.Log.Criticalf("Recovering from DHCP panic %s", r)
			buf := make([]byte, 2048)
			runtime.Stack(buf, false)
			h.c.Log.Criticalf("Stack Trace: %s", string(buf))
		}
	}()
	if msgType == dhcp4.Inform {
		return nil
	}

	// Log every message
	if server, ok := options[dhcp4.OptionServerIdentifier]; !ok || net.IP(server).Equal(c.global.serverIdentifier) {
		h.c.Log.WithFields(verbose.Fields{
			"Type":       msgType.String(),
			"Client IP":  p.CIAddr().String(),
			"Client MAC": p.CHAddr().String(),
			"Relay IP":   p.GIAddr().String(),
		}).Debug()
	}

	var response dhcp4.Packet
	switch msgType {
	case dhcp4.Discover:
		response = h.handleDiscover(p, options)
	case dhcp4.Request:
		response = h.handleRequest(p, options)
	case dhcp4.Release:
		response = h.handleRelease(p, options)
	case dhcp4.Decline:
		response = h.handleDecline(p, options)
	}
	return response
}

// Only allows real packet returns if not in readonly mode
func (h *Handler) readOnlyFilter(p dhcp4.Packet) dhcp4.Packet {
	if h.ro {
		return nil
	}
	return p
}

// Handle DHCP DISCOVER messages
func (h *Handler) handleDiscover(p dhcp4.Packet, options dhcp4.Options) dhcp4.Packet {
	// Don't respond to requests on the same subnet
	if p.GIAddr().Equal(net.IPv4zero) {
		return nil
	}

	// Get a device object associated with the MAC
	device, err := h.c.DeviceStore.GetDeviceByMAC(p.CHAddr())
	if err != nil {
		h.c.Log.WithFields(verbose.Fields{
			"Client MAC": p.CHAddr().String(),
			"Error":      err,
		}).Error("Error getting device")
		return nil
	}

	// Check device standing
	if device.IsBlacklisted() {
		h.c.Log.WithFields(verbose.Fields{
			"Client MAC":  p.CHAddr().String(),
			"Relay Agent": p.GIAddr().String(),
			"Username":    device.GetUsername(),
		}).Notice("Blacklisted MAC got a lease")
	}

	// Get network object that the relay IP belongs to
	h.gatewayMutex.Lock()
	network, ok := h.gatewayCache[p.GIAddr().String()]
	if !ok {
		// That gateway hasn't been seen before, find its network
		network = c.searchNetworksFor(p.GIAddr())
		if network == nil {
			h.gatewayMutex.Unlock()
			return nil
		}
		// Add to cache for later
		h.gatewayCache[p.GIAddr().String()] = network
	}
	h.gatewayMutex.Unlock()

	// Find an appropiate lease
	lease, pool := network.getLeaseByMAC(device.GetMAC(), device.IsRegistered())
	if lease == nil {
		// Device doesn't have a recent lease, get a new one
		lease, pool = network.getFreeLease(h.c, device.IsRegistered())
		if lease == nil { // No free lease was found, be more aggressive
			lease, pool = network.getFreeLeaseDesperate(h.c, device.IsRegistered())
		}
		if lease == nil { // Still no lease was found, error and go to the next request
			h.c.Log.WithFields(verbose.Fields{
				"Network":    network.name,
				"Registered": device.IsRegistered(),
			}).Alert("No free leases available in network")
			return nil
		}
	}

	h.c.Log.WithFields(verbose.Fields{
		"Lease IP":   lease.IP.String(),
		"Client MAC": p.CHAddr().String(),
		"Registered": device.IsRegistered(),
	}).Info("Offering lease to client")

	// Set temporary offered flag and end time
	lease.Offered = true
	lease.Start = time.Now()
	lease.End = time.Now().Add(time.Duration(30) * time.Second) // Set a short end time so it's not offered to other clients
	// p.CHAddr() returns a slice. A slice is basically a pointer. A pointer is NOT the value.
	// Copy the VALUE of the mac address into the lease, not the pointer. Otherwise you're gonna have a bad time.
	lease.MAC = make([]byte, len(p.CHAddr()))
	copy(lease.MAC, p.CHAddr())
	// No Save because this is a temporary "lease", if the client accepts then we commit to storage
	// Get options
	leaseOptions := pool.getOptions(device.IsRegistered())
	// Send an offer
	return h.readOnlyFilter(dhcp4.ReplyPacket(
		p,
		dhcp4.Offer,
		c.global.serverIdentifier,
		lease.IP,
		pool.getLeaseTime(0, device.IsRegistered()),
		leaseOptions.SelectOrderOrAll(options[dhcp4.OptionParameterRequestList]),
	))
}

// Handle DHCP REQUEST messages
func (h *Handler) handleRequest(p dhcp4.Packet, options dhcp4.Options) dhcp4.Packet {
	if server, ok := options[dhcp4.OptionServerIdentifier]; ok && !net.IP(server).Equal(c.global.serverIdentifier) {
		return nil // Message not for this dhcp server
	}

	reqIP := net.IP(options[dhcp4.OptionRequestedIPAddress])
	if reqIP == nil {
		reqIP = net.IP(p.CIAddr())
	}

	if len(reqIP) != 4 || reqIP.Equal(net.IPv4zero) {
		return h.readOnlyFilter(dhcp4.ReplyPacket(p, dhcp4.NAK, c.global.serverIdentifier, nil, 0, nil))
	}

	// Get a device object associated with the MAC
	device, err := h.c.DeviceStore.GetDeviceByMAC(p.CHAddr())
	if err != nil {
		h.c.Log.WithFields(verbose.Fields{
			"Client MAC": p.CHAddr().String(),
			"Error":      err,
		}).Error("Error getting device")
		return h.readOnlyFilter(dhcp4.ReplyPacket(p, dhcp4.NAK, c.global.serverIdentifier, nil, 0, nil))
	}

	// Check device standing
	if device.IsBlacklisted() {
		h.c.Log.WithFields(verbose.Fields{
			"Client MAC":  p.CHAddr().String(),
			"Relay Agent": p.GIAddr().String(),
			"Username":    device.GetUsername(),
		}).Notice("Blacklisted MAC renewed lease")
	}

	var network *network
	// Get network object that the relay or client IP belongs to
	if p.GIAddr().Equal(net.IPv4zero) {
		// Coming directly from the client
		network = c.searchNetworksFor(reqIP)
	} else {
		// Coming from a relay
		h.gatewayMutex.Lock()
		var ok bool
		network, ok = h.gatewayCache[p.GIAddr().String()]
		h.gatewayMutex.Unlock()
		if !ok {
			// That gateway hasn't been seen before, it needs to go through DISCOVER
			return h.readOnlyFilter(dhcp4.ReplyPacket(p, dhcp4.NAK, c.global.serverIdentifier, nil, 0, nil))
		}
	}

	if network == nil {
		h.c.Log.WithFields(verbose.Fields{
			"Requested IP": reqIP.String(),
			"Registered":   device.IsRegistered(),
		}).Info("Got a REQUEST for IP not in a scope")
		return h.readOnlyFilter(dhcp4.ReplyPacket(p, dhcp4.NAK, c.global.serverIdentifier, nil, 0, nil))
	}

	lease, pool := network.getLeaseByIP(reqIP, device.IsRegistered())
	if lease == nil || lease.MAC == nil { // If it returns a new lease, the MAC is nil
		h.c.Log.WithFields(verbose.Fields{
			"Requested IP": reqIP.String(),
			"Client MAC":   p.CHAddr().String(),
			"Network":      network.name,
			"Registered":   device.IsRegistered(),
		}).Info("Client tried to request a lease that doesn't exist")
		return h.readOnlyFilter(dhcp4.ReplyPacket(p, dhcp4.NAK, c.global.serverIdentifier, nil, 0, nil))
	}

	if !bytes.Equal(lease.MAC, p.CHAddr()) {
		h.c.Log.WithFields(verbose.Fields{
			"Requested IP": reqIP.String(),
			"Client MAC":   p.CHAddr().String(),
			"Lease MAC":    lease.MAC.String(),
			"Network":      network.name,
			"Registered":   device.IsRegistered(),
		}).Info("Client tried to request lease not belonging to them")
		return h.readOnlyFilter(dhcp4.ReplyPacket(p, dhcp4.NAK, c.global.serverIdentifier, nil, 0, nil))
	}

	leaseDur := pool.getLeaseTime(0, device.IsRegistered())
	lease.Start = time.Now()
	lease.End = time.Now().Add(leaseDur + (time.Duration(10) * time.Second)) // Add 10 seconds to account for slight clock drift
	lease.Offered = false
	if ci, ok := options[dhcp4.OptionHostName]; ok {
		lease.Hostname = string(ci)
	}
	if err := lease.Save(); err != nil {
		h.c.Log.WithFields(verbose.Fields{
			"Client MAC": p.CHAddr().String(),
			"Error":      err,
		}).Error("Error saving lease")
		return h.readOnlyFilter(dhcp4.ReplyPacket(p, dhcp4.NAK, c.global.serverIdentifier, nil, 0, nil))
	}
	leaseOptions := pool.getOptions(device.IsRegistered())

	h.c.Log.WithFields(verbose.Fields{
		"Requested IP": lease.IP.String(),
		"Client MAC":   lease.MAC.String(),
		"Duration":     leaseDur.String(),
		"Network":      network.name,
		"Registered":   device.IsRegistered(),
		"Hostname":     lease.Hostname,
	}).Info("Acknowledging request")

	if device.IsRegistered() {
		device.SetLastSeen(time.Now())
		if err := device.Save(); err != nil {
			// We won't consider this a critical error, still give out the lease
			h.c.Log.WithField("Err", err).Error("Failed updating device last_seen attribute")
		}
	}

	return h.readOnlyFilter(dhcp4.ReplyPacket(
		p,
		dhcp4.ACK,
		c.global.serverIdentifier,
		lease.IP,
		leaseDur,
		leaseOptions.SelectOrderOrAll(options[dhcp4.OptionParameterRequestList]),
	))
}

// Handle DHCP RELEASE messages
func (h *Handler) handleRelease(p dhcp4.Packet, options dhcp4.Options) dhcp4.Packet {
	reqIP := p.CIAddr()
	if reqIP == nil || reqIP.Equal(net.IPv4zero) {
		return nil
	}

	// Get a device object associated with the MAC
	device, err := h.c.DeviceStore.GetDeviceByMAC(p.CHAddr())
	if err != nil {
		h.c.Log.WithFields(verbose.Fields{
			"Client MAC": p.CHAddr().String(),
			"Error":      err,
		}).Error("Error getting device")
		return nil
	}

	network := c.searchNetworksFor(reqIP)
	if network == nil {
		h.c.Log.WithFields(verbose.Fields{
			"Releasing IP": reqIP.String(),
			"Registered":   device.IsRegistered(),
		}).Notice("Got a RELEASE for IP not in a scope")
		return nil
	}

	lease, _ := network.getLeaseByIP(reqIP, device.IsRegistered())
	if lease == nil || !bytes.Equal(lease.MAC, p.CHAddr()) {
		h.c.Log.WithFields(verbose.Fields{
			"Releasing IP": reqIP.String(),
			"Client MAC":   p.CHAddr().String(),
			"Lease MAC":    lease.MAC.String(),
			"Network":      network.name,
			"Registered":   device.IsRegistered(),
		}).Notice("Client tried to release lease not belonging to them")
		return nil
	}
	h.c.Log.WithFields(verbose.Fields{
		"IP":         lease.IP.String(),
		"Client MAC": lease.MAC.String(),
		"Network":    network.name,
		"Registered": device.IsRegistered(),
	}).Info("Releasing lease")
	lease.Start = time.Unix(1, 0)
	lease.End = time.Unix(1, 0)
	if err := lease.Save(); err != nil {
		h.c.Log.WithFields(verbose.Fields{
			"Client MAC": p.CHAddr().String(),
			"Error":      err,
		}).Error("Error saving lease")
	}
	return nil
}

// Handle DHCP DECLINE messages
func (h *Handler) handleDecline(p dhcp4.Packet, options dhcp4.Options) dhcp4.Packet {
	reqIP := p.CIAddr()
	if reqIP == nil || reqIP.Equal(net.IPv4zero) {
		return nil
	}

	// Get a device object associated with the MAC
	device, err := h.c.DeviceStore.GetDeviceByMAC(p.CHAddr())
	if err != nil {
		h.c.Log.WithFields(verbose.Fields{
			"Client MAC": p.CHAddr().String(),
			"Error":      err,
		}).Error("Error getting device")
		return nil
	}

	network := c.searchNetworksFor(reqIP)
	if network == nil {
		h.c.Log.WithFields(verbose.Fields{
			"Releasing IP": reqIP.String(),
			"Registered":   device.IsRegistered(),
		}).Notice("Got a DECLINE for IP not in a scope")
		return nil
	}

	lease, _ := network.getLeaseByIP(reqIP, device.IsRegistered())
	if lease == nil || !bytes.Equal(lease.MAC, p.CHAddr()) {
		h.c.Log.WithFields(verbose.Fields{
			"Declined IP": reqIP.String(),
			"Client MAC":  p.CHAddr().String(),
			"Lease MAC":   lease.MAC.String(),
			"Network":     network.name,
			"Registered":  device.IsRegistered(),
		}).Notice("Client tried to decline lease not belonging to them")
		return nil
	}
	h.c.Log.WithFields(verbose.Fields{
		"IP":         lease.IP.String(),
		"Client MAC": lease.MAC.String(),
		"Network":    network.name,
		"Registered": device.IsRegistered(),
	}).Notice("Abandoned lease")
	lease.IsAbandoned = true
	lease.Start = time.Unix(1, 0)
	lease.End = time.Unix(1, 0)
	if err := lease.Save(); err != nil {
		h.c.Log.WithFields(verbose.Fields{
			"Client MAC": p.CHAddr().String(),
			"Error":      err,
		}).Error("Error saving lease")
	}
	return nil
}
