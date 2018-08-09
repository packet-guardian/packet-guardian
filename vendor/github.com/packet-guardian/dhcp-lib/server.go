// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// TODO: Clean up the handler functions. There's a lot of duplicated code that
// could be extracted to a function.

package dhcp

var (
	c *Config
)

// A Handler processes all incoming DHCP packets.
type Handler struct {
	c *ServerConfig
}

// NewDHCPServer creates and sets up a new DHCP Handler with the give configuration.
func NewDHCPServer(conf *Config, s *ServerConfig) *Handler {
	c = conf
	return &Handler{
		c: s,
	}
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
				break subnetLoop
			}
		}
	}
	return nil
}
