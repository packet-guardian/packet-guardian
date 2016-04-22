package dhcp

import (
	"time"

	"github.com/krolaw/dhcp4"
	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/models"
)

type DHCPHandler struct {
	c *Config
	e *common.Environment
}

func NewDHCPServer(c *Config, e *common.Environment) *DHCPHandler {
	return &DHCPHandler{
		c: c,
		e: e,
	}
}

func (h *DHCPHandler) ListenAndServe() {
	dhcp4.ListenAndServe(h)
}

func (h *DHCPHandler) ServeDHCP(p dhcp4.Packet, msgType dhcp4.MessageType, options dhcp4.Options) (d dhcp4.Packet) {
	switch msgType {
	case dhcp4.Discover:
		return h.handleDiscover(p, msgType, options)
	case dhcp4.Request:
		return h.handleRequest(p, msgType, options)
	case dhcp4.Release:
		return h.handleRelease(p, msgType, options)
	case dhcp4.Decline:
		return h.handleDecline(p, msgType, options)
	case dhcp4.Inform:
		return h.handleInform(p, msgType, options)
	}
	return nil
}

func (h *DHCPHandler) handleDiscover(p dhcp4.Packet, msgType dhcp4.MessageType, options dhcp4.Options) (d dhcp4.Packet) {
	logger := h.e.Log.GetLogger("dhcp")
	// Get a device object associated with the MAC
	device, err := models.GetDeviceByMAC(h.e, p.CHAddr())
	if err != nil {
		logger.Errorf("Error getting device: %s", err.Error())
		return nil
	}

	// Check device standing
	if device.IsBlacklisted {
		logger.Errorf("Blacklisted MAC %s tried to get a lease from relay %s", p.CHAddr().String(), p.GIAddr().String())
		return nil
	}
	registered := (device.ID != 0)

	// Get network object that the relay IP belongs to
	network := h.c.SearchNetworksFor(p.GIAddr())
	if network == nil {
		return nil
	}

	// Find an appropiate lease
	lease := network.GetLeaseByMAC(device.MAC, registered)
	if lease == nil {
		// Device doesn't have a recent lease, get a new one
		lease = network.GetFreeLease(h.e, registered)
		if lease == nil {
			logger.Errorf("No free leases available in network %s where registered is %t", network.Name, registered)
			return nil
		}
	}

	// Set temporary offered flag and end time
	lease.Offered = true
	lease.End = time.Now().Add(time.Duration(10) * time.Second) // Set a short end time so it's not offered to other clients
	// Get options
	leaseOptions := lease.Pool.GetOptions(registered)
	// Sent an offer
	return dhcp4.ReplyPacket(
		p,
		dhcp4.Offer,
		h.c.Global.ServerIdentifier,
		lease.IP,
		lease.Pool.GetLeaseTime(0, registered),
		leaseOptions.SelectOrderOrAll(options[dhcp4.OptionParameterRequestList]),
	)
}

func (h *DHCPHandler) handleRequest(p dhcp4.Packet, msgType dhcp4.MessageType, options dhcp4.Options) (d dhcp4.Packet) {
	// Check if for this server
	//
	// Get requested IP, if one not send use CIAddr
	//
	// Validate IP
	// - Correct length
	// - In the network (registration)
	//
	// Validate lease
	// - Lease doesn't exist OR
	// - Lease MAC == client MAC
	//
	// Save lease to database
	return nil
}

func (h *DHCPHandler) handleRelease(p dhcp4.Packet, msgType dhcp4.MessageType, options dhcp4.Options) (d dhcp4.Packet) {
	// Release: Remove lease from database
	return nil
}

func (h *DHCPHandler) handleDecline(p dhcp4.Packet, msgType dhcp4.MessageType, options dhcp4.Options) (d dhcp4.Packet) {
	// Decline: Don't do anything?
	return nil
}

func (h *DHCPHandler) handleInform(p dhcp4.Packet, msgType dhcp4.MessageType, options dhcp4.Options) (d dhcp4.Packet) {
	// If IP is in a network, mark as abandoned
	// Maybe log until sure what this does
	return nil
}

// Example from author
//
// func (h *DHCPHandler) ServeDHCP(p dhcp.Packet, msgType dhcp.MessageType, options dhcp.Options) (d dhcp.Packet) {
// 	switch msgType {
//
// 	case dhcp.Discover:
// 		free, nic := -1, p.CHAddr().String()
// 		for i, v := range h.leases { // Find previous lease
// 			if v.nic == nic {
// 				free = i
// 				goto reply
// 			}
// 		}
// 		if free = h.freeLease(); free == -1 {
// 			return
// 		}
// 	reply:
// 		return dhcp.ReplyPacket(p, dhcp.Offer, h.ip, dhcp.IPAdd(h.start, free), h.leaseDuration,
// 			h.options.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
//
// 	case dhcp.Request:
// 		if server, ok := options[dhcp.OptionServerIdentifier]; ok && !net.IP(server).Equal(h.ip) {
// 			return nil // Message not for this dhcp server
// 		}
// 		reqIP := net.IP(options[dhcp.OptionRequestedIPAddress])
// 		if reqIP == nil {
// 			reqIP = net.IP(p.CIAddr())
// 		}
//
// 		if len(reqIP) == 4 && !reqIP.Equal(net.IPv4zero) {
// 			if leaseNum := dhcp.IPRange(h.start, reqIP) - 1; leaseNum >= 0 && leaseNum < h.leaseRange {
// 				if l, exists := h.leases[leaseNum]; !exists || l.nic == p.CHAddr().String() {
// 					h.leases[leaseNum] = lease{nic: p.CHAddr().String(), expiry: time.Now().Add(h.leaseDuration)}
// 					return dhcp.ReplyPacket(p, dhcp.ACK, h.ip, net.IP(options[dhcp.OptionRequestedIPAddress]), h.leaseDuration,
// 						h.options.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
// 				}
// 			}
// 		}
// 		return dhcp.ReplyPacket(p, dhcp.NAK, h.ip, nil, 0, nil)
//
// 	case dhcp.Release, dhcp.Decline:
// 		nic := p.CHAddr().String()
// 		for i, v := range h.leases {
// 			if v.nic == nic {
// 				delete(h.leases, i)
// 				break
// 			}
// 		}
// 	}
// 	return nil
// }
