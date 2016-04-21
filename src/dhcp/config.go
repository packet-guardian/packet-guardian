package dhcp

import (
	"net"
	"time"

	"github.com/krolaw/dhcp4"
)

type Config struct {
	Global   *Global
	Networks []*Network
}

type Global struct {
	ServerIdentifier     net.IP
	Settings             *Settings
	RegisteredSettings   *Settings
	UnregisteredSettings *Settings
}

type Network struct {
	Settings             *Settings
	RegisteredSettings   *Settings
	UnregisteredSettings *Settings
	Subnets              []*Subnet
}

type Settings struct {
	Options          map[dhcp4.OptionCode][]byte
	DefaultLeaseTime time.Duration
	MaxLeaseTime     time.Duration
}

type Subnet struct {
	AllowUnknown bool
	Settings     *Settings
	Network      *net.IPNet
	Pools        []*Pool
}

type Pool struct {
	RangeStart net.IP
	RangeEnd   net.IP
	Settings   *Settings
}
