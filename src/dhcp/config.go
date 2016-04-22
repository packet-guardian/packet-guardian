package dhcp

import (
	"fmt"
	"net"
	"time"

	"github.com/krolaw/dhcp4"
)

type Config struct {
	Global   *Global
	Networks []*Network
}

func (c *Config) Print() {
	fmt.Println("DHCP Configuration")
	c.Global.print()
	for _, n := range c.Networks {
		n.print()
	}
}

type Global struct {
	ServerIdentifier     net.IP
	Settings             *Settings
	RegisteredSettings   *Settings
	UnregisteredSettings *Settings
}

func newGlobal() *Global {
	return &Global{
		Settings:             newSettingsBlock(),
		RegisteredSettings:   newSettingsBlock(),
		UnregisteredSettings: newSettingsBlock(),
	}
}

func (g *Global) print() {
	fmt.Println("\n---Global Configuration---")
	fmt.Printf("Server Identifier: %s\n", g.ServerIdentifier.String())
	fmt.Println("\n--Global Settings--")
	g.Settings.print()
	fmt.Println("\n--Global Registered Settings--")
	g.RegisteredSettings.print()
	fmt.Println("\n--Global Unregistered Settings--")
	g.UnregisteredSettings.print()
}

type Network struct {
	Name                 string
	Settings             *Settings
	RegisteredSettings   *Settings
	UnregisteredSettings *Settings
	Subnets              []*Subnet
}

func newNetwork(name string) *Network {
	return &Network{
		Name:                 name,
		Settings:             newSettingsBlock(),
		RegisteredSettings:   newSettingsBlock(),
		UnregisteredSettings: newSettingsBlock(),
	}
}

func (n *Network) print() {
	fmt.Printf("\n---Network Configuration - %s---\n", n.Name)
	fmt.Println("\n--Network Settings--")
	n.Settings.print()
	fmt.Println("\n--Network Registered Settings--")
	n.RegisteredSettings.print()
	fmt.Println("\n--Network Unregistered Settings--")
	n.UnregisteredSettings.print()
	fmt.Println("\n--Subnets in network--")
	for _, s := range n.Subnets {
		s.print()
	}
}

type Settings struct {
	Options          map[dhcp4.OptionCode][]byte
	DefaultLeaseTime time.Duration
	MaxLeaseTime     time.Duration
}

func newSettingsBlock() *Settings {
	return &Settings{
		Options: make(map[dhcp4.OptionCode][]byte),
	}
}

func (s *Settings) print() {
	fmt.Printf("Default Lease Time: %s\n", s.DefaultLeaseTime.String())
	fmt.Printf("Max Lease Time: %s\n", s.MaxLeaseTime.String())
	fmt.Println("-DHCP Options-")
	for c, v := range s.Options {
		fmt.Printf("%s: %s\n", c.String(), string(v))
	}
}

type Subnet struct {
	AllowUnknown bool
	Settings     *Settings
	Network      *net.IPNet
	Pools        []*Pool
}

func newSubnet() *Subnet {
	return &Subnet{
		Settings: newSettingsBlock(),
	}
}

func (s *Subnet) print() {
	fmt.Printf("\n---Subnet - %s---\n", s.Network.String())
	fmt.Printf("Registered: %t\n", !s.AllowUnknown)
	fmt.Println("Subnet Settings")
	s.Settings.print()
	fmt.Println("Subnet Pools")
	for _, p := range s.Pools {
		p.print()
	}
}

type Pool struct {
	RangeStart net.IP
	RangeEnd   net.IP
	Settings   *Settings
}

func newPool() *Pool {
	return &Pool{
		Settings: newSettingsBlock(),
	}
}

func (p *Pool) print() {
	fmt.Printf("\n---Pool %s - %s---\n", p.RangeStart.String(), p.RangeEnd.String())
	fmt.Println("Pool Settings")
	p.Settings.print()
}
