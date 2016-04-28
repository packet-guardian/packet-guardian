package dhcp

import (
	"fmt"
	"net"
)

type Config struct {
	Global   *Global
	Networks map[string]*Network
}

func newConfig() *Config {
	return &Config{
		Networks: make(map[string]*Network),
	}
}

func (c *Config) Print() {
	fmt.Println("DHCP Configuration")
	c.Global.Print()
	for _, n := range c.Networks {
		n.Print()
	}
}

func (c *Config) SearchNetworksFor(ip net.IP) *Network {
	for _, network := range c.Networks {
		if network.Includes(ip) {
			return network
		}
	}
	return nil
}
