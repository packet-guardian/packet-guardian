package dhcp

import (
	"fmt"
	"time"

	"github.com/krolaw/dhcp4"
)

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

func (s *Settings) Print() {
	fmt.Printf("Default Lease Time: %s\n", s.DefaultLeaseTime.String())
	fmt.Printf("Max Lease Time: %s\n", s.MaxLeaseTime.String())
	fmt.Println("-DHCP Options-")
	for c, v := range s.Options {
		fmt.Printf("%s: %v\n", c.String(), v)
	}
}
