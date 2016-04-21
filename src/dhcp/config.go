package dhcp

import (
	"net"
	"time"
)

type Option int

type Config struct {
	Global   *GlobalConfig
	Networks []*NetworkConfig
}

type GlobalConfig struct {
	Settings             *SettingsConfig
	RegisteredSettings   *SettingsConfig
	UnregisteredSettings *SettingsConfig
}

type NetworkConfig struct {
	Settings             *SettingsConfig
	RegisteredSettings   *SettingsConfig
	UnregisteredSettings *SettingsConfig
	Subnets              []*SubnetConfig
}

type SettingsConfig struct {
	// Replace with a map of the dhcp server package's options
	Options          map[Option][]byte
	DefaultLeaseTime time.Duration
	MaxLeaseTime     time.Duration
}

type SubnetConfig struct {
	AllowUnknown bool
	Settings     *SettingsConfig
	Network      *net.IPNet
	Pools        []*Pool
}

type Pool struct {
	RangeStart net.IP
	RangeEnd   net.IP
	Settings   *SettingsConfig
}
