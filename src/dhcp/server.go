package dhcp

import (
	"bytes"
	"net"
	"sync"
	"time"

	"github.com/dragonrider23/verbose"
	"github.com/krolaw/dhcp4"
	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/models"
)

var (
	gatewayCache = map[string]*Network{}
	gatewayMutex = sync.Mutex{}
)

type DHCPHandler struct {
	c  *Config
	e  *common.Environment
	ro bool
}

func NewDHCPServer(c *Config, e *common.Environment) *DHCPHandler {
	return &DHCPHandler{
		c: c,
		e: e,
	}
}

func (h *DHCPHandler) ListenAndServe() error {
	h.e.Log.Info("Starting DHCP server")
	return dhcp4.ListenAndServe(h)
}

func (h *DHCPHandler) Readonly() {
	h.ro = true
}

func (h *DHCPHandler) Respond() {
	h.ro = false
}

func (h *DHCPHandler) LoadLeases() error {
	// Get leases from storage
	leases, err := GetAllLeases(h.e)
	if err != nil {
		return err
	}
	if leases == nil {
		return nil
	}

	// Find the pool each lease belongs to
	for _, lease := range leases {
		n, ok := h.c.Networks[lease.Network]
		if !ok {
			continue
		}
	subnetLoop:
		for _, subnet := range n.Subnets {
			if !subnet.Includes(lease.IP) {
				continue
			}
			for _, pool := range subnet.Pools {
				if !pool.Includes(lease.IP) {
					continue
				}
				lease.Pool = pool
				break subnetLoop
			}
		}
	}
	return nil
}

func (h *DHCPHandler) ServeDHCP(p dhcp4.Packet, msgType dhcp4.MessageType, options dhcp4.Options) dhcp4.Packet {
	if msgType == dhcp4.Inform {
		return nil
	}

	// Log every message
	if server, ok := options[dhcp4.OptionServerIdentifier]; !ok || net.IP(server).Equal(h.c.Global.ServerIdentifier) {
		h.e.Log.WithFields(verbose.Fields{
			"Type":       msgType.String(),
			"Client IP":  p.CIAddr().String(),
			"Client MAC": p.CHAddr().String(),
			"Relay IP":   p.GIAddr().String(),
		}).Info()
	}

	var response dhcp4.Packet
	switch msgType {
	case dhcp4.Discover:
		response = h.handleDiscover(p, msgType, options)
	case dhcp4.Request:
		response = h.handleRequest(p, msgType, options)
	case dhcp4.Release:
		response = h.handleRelease(p, msgType, options)
	case dhcp4.Decline:
		//return h.handleDecline(p, msgType, options)
		// Mark address as abandoned
	}
	return response
}

// Handle DHCP DISCOVER messages
func (h *DHCPHandler) handleDiscover(p dhcp4.Packet, msgType dhcp4.MessageType, options dhcp4.Options) dhcp4.Packet {
	// Get a device object associated with the MAC
	device, err := models.GetDeviceByMAC(h.e, p.CHAddr())
	if err != nil {
		h.e.Log.WithFields(verbose.Fields{
			"MAC":   p.CHAddr().String(),
			"Error": err,
		}).Error("Error getting device")
		return nil
	}

	// Check device standing
	if device.IsBlacklisted {
		h.e.Log.WithFields(verbose.Fields{
			"MAC":      p.CHAddr().String(),
			"Gateway":  p.GIAddr().String(),
			"Username": device.Username,
		}).Warning("Blacklisted MAC tried to get a lease")
		return nil
	}
	registered := (device.ID != 0)

	// Get network object that the relay IP belongs to
	gatewayMutex.Lock()
	network, ok := gatewayCache[p.GIAddr().String()]
	if !ok {
		// That gateway hasn't been seen before, find its network
		network = h.c.SearchNetworksFor(p.GIAddr())
		if network == nil {
			gatewayMutex.Unlock()
			return nil
		}
		// Add to cache for later
		gatewayCache[p.GIAddr().String()] = network
	}
	gatewayMutex.Unlock()

	// Find an appropiate lease
	h.e.Log.WithFields(verbose.Fields{
		"MAC": device.MAC.String(),
	}).Debug()
	network.Subnets[0].Pools[0].PrintLeases()
	lease := network.GetLeaseByMAC(device.MAC, registered)
	if lease == nil {
		// Device doesn't have a recent lease, get a new one
		lease = network.GetFreeLease(h.e, registered)
		if lease == nil {
			h.e.Log.WithFields(verbose.Fields{
				"Network":    network.Name,
				"Registered": registered,
			}).Alert("No free leases available in network")
			return nil
		}
		h.e.Log.Debug("Got new, free lease")
	} else {
		h.e.Log.Debug("Got existing lease")
	}

	h.e.Log.WithFields(verbose.Fields{
		"IP":  lease.IP.String(),
		"MAC": p.CHAddr().String(),
	}).Info("Offering lease to client")

	// Set temporary offered flag and end time
	lease.Offered = true
	lease.Start = time.Now()
	lease.End = time.Now().Add(time.Duration(10) * time.Second) // Set a short end time so it's not offered to other clients
	// p.CHAddr() returns a slice. A slice is basically a pointer. A pointer is NOT the value.
	// Copy the VALUE of the mac address into the lease, not the pointer. Otherwise you're gonna have a bad time.
	lease.MAC = make([]byte, len(p.CHAddr()))
	copy(lease.MAC, p.CHAddr())
	// No Save because this is a temporary "lease", if the client accepts then we commit to storage
	// Get options
	leaseOptions := lease.Pool.GetOptions(registered)
	lease.Pool.PrintLeases()
	// Send an offer
	return h.readOnlyFilter(dhcp4.ReplyPacket(
		p,
		dhcp4.Offer,
		h.c.Global.ServerIdentifier,
		lease.IP,
		lease.Pool.GetLeaseTime(0, registered),
		leaseOptions.SelectOrderOrAll(options[dhcp4.OptionParameterRequestList]),
	))
}

// Only allows real packet returns if not in readonly mode
func (h *DHCPHandler) readOnlyFilter(p dhcp4.Packet) dhcp4.Packet {
	if h.ro {
		return nil
	}
	return p
}

// Handle DHCP REQUEST messages
func (h *DHCPHandler) handleRequest(p dhcp4.Packet, msgType dhcp4.MessageType, options dhcp4.Options) dhcp4.Packet {
	if server, ok := options[dhcp4.OptionServerIdentifier]; ok && !net.IP(server).Equal(h.c.Global.ServerIdentifier) {
		return nil // Message not for this dhcp server
	}

	reqIP := net.IP(options[dhcp4.OptionRequestedIPAddress])
	if reqIP == nil {
		reqIP = net.IP(p.CIAddr())
	}

	if len(reqIP) != 4 || reqIP.Equal(net.IPv4zero) {
		return h.readOnlyFilter(dhcp4.ReplyPacket(p, dhcp4.NAK, h.c.Global.ServerIdentifier, nil, 0, nil))
	}

	// Get a device object associated with the MAC
	device, err := models.GetDeviceByMAC(h.e, p.CHAddr())
	if err != nil {
		h.e.Log.WithFields(verbose.Fields{
			"MAC":   p.CHAddr().String(),
			"Error": err,
		}).Error("Error getting device")
		return h.readOnlyFilter(dhcp4.ReplyPacket(p, dhcp4.NAK, h.c.Global.ServerIdentifier, nil, 0, nil))
	}

	// Check device standing
	if device.IsBlacklisted {
		h.e.Log.WithFields(verbose.Fields{
			"MAC":      p.CHAddr().String(),
			"Gateway":  p.GIAddr().String(),
			"Username": device.Username,
		}).Notice("Blacklisted MAC tried to renew lease")
		return h.readOnlyFilter(dhcp4.ReplyPacket(p, dhcp4.NAK, h.c.Global.ServerIdentifier, nil, 0, nil))
	}
	registered := (device.ID != 0)

	network := h.c.SearchNetworksFor(reqIP)
	if network == nil {
		h.e.Log.WithFields(verbose.Fields{
			"IP": reqIP.String(),
		}).Notice("Got a REQUEST for IP not in a scope")
		return h.readOnlyFilter(dhcp4.ReplyPacket(p, dhcp4.NAK, h.c.Global.ServerIdentifier, nil, 0, nil))
	}

	lease := network.GetLeaseByIP(reqIP, registered)
	if lease == nil {
		h.e.Log.WithFields(verbose.Fields{
			"IP":  reqIP.String(),
			"MAC": p.CHAddr().String(),
		}).Notice("Client tried to request a lease that doesn't exist")
		return h.readOnlyFilter(dhcp4.ReplyPacket(p, dhcp4.NAK, h.c.Global.ServerIdentifier, nil, 0, nil))
	}

	if !bytes.Equal(lease.MAC, p.CHAddr()) {
		h.e.Log.WithFields(verbose.Fields{
			"IP":        reqIP.String(),
			"MAC":       p.CHAddr().String(),
			"Lease MAC": lease.MAC.String(),
		}).Notice("Client tried to request lease not belonging to them")
		return h.readOnlyFilter(dhcp4.ReplyPacket(p, dhcp4.NAK, h.c.Global.ServerIdentifier, nil, 0, nil))
	}

	leaseDur := lease.Pool.GetLeaseTime(0, registered)
	lease.Start = time.Now()
	lease.End = time.Now().Add(leaseDur + (time.Duration(10) * time.Second)) // Add 10 seconds to account for slight clock drift
	lease.Offered = false
	if ci, ok := options[dhcp4.OptionClientIdentifier]; ok {
		lease.Hostname = string(ci)
	}
	lease.Save()
	leaseOptions := lease.Pool.GetOptions(registered)
	lease.Pool.PrintLeases()

	h.e.Log.WithFields(verbose.Fields{
		"IP":       lease.IP.String(),
		"MAC":      lease.MAC.String(),
		"Duration": leaseDur.String(),
	}).Info("Acknowledging request")

	return h.readOnlyFilter(dhcp4.ReplyPacket(
		p,
		dhcp4.ACK,
		h.c.Global.ServerIdentifier,
		net.IP(options[dhcp4.OptionRequestedIPAddress]),
		leaseDur,
		leaseOptions.SelectOrderOrAll(options[dhcp4.OptionParameterRequestList]),
	))
}

// Handle DHCP RELEASE messages
func (h *DHCPHandler) handleRelease(p dhcp4.Packet, msgType dhcp4.MessageType, options dhcp4.Options) dhcp4.Packet {
	reqIP := p.CIAddr()
	if reqIP == nil || reqIP.Equal(net.IPv4zero) {
		return nil
	}

	// Get a device object associated with the MAC
	device, err := models.GetDeviceByMAC(h.e, p.CHAddr())
	if err != nil {
		h.e.Log.WithFields(verbose.Fields{
			"MAC":   p.CHAddr().String(),
			"Error": err,
		}).Error("Error getting device")
		return nil
	}
	registered := (device.ID != 0)

	network := h.c.SearchNetworksFor(reqIP)
	if network == nil {
		return nil
	}

	lease := network.GetLeaseByIP(reqIP, registered)
	if lease == nil || !bytes.Equal(lease.MAC, p.CHAddr()) {
		return nil
	}
	h.e.Log.WithFields(verbose.Fields{
		"IP":  lease.IP.String(),
		"MAC": lease.MAC.String(),
	}).Info("Releasing lease")
	lease.End = time.Unix(1, 0)
	lease.Save()
	return nil
}
