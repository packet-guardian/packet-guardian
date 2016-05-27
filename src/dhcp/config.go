// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"fmt"
	"net"
)

// A Config is the parsed object generated from a PG-DHCP configuration file.
type Config struct {
	global   *global
	networks map[string]*network
}

func newConfig() *Config {
	return &Config{
		networks: make(map[string]*network),
	}
}

func (c *Config) print() {
	fmt.Println("DHCP Configuration")
	c.global.print()
	for _, n := range c.networks {
		n.print()
	}
}

func (c *Config) searchNetworksFor(ip net.IP) *network {
	for _, network := range c.networks {
		if network.includes(ip) {
			return network
		}
	}
	return nil
}
