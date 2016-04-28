package dhcp

import (
	"fmt"
	"net"
	"time"

	"github.com/onesimus-systems/dhcp4"
)

type Global struct {
	ServerIdentifier     net.IP
	Settings             *Settings
	RegisteredSettings   *Settings
	regOptionsCached     bool
	UnregisteredSettings *Settings
	unregOptionsCached   bool
}

func newGlobal() *Global {
	return &Global{
		Settings:             newSettingsBlock(),
		RegisteredSettings:   newSettingsBlock(),
		UnregisteredSettings: newSettingsBlock(),
	}
}

// GetLeaseTime returns the lease time given the requested time req and if the client is registered.
// If req is 0 then the default lease time is returned. Otherwise it will return the lower of
// req and the maximum lease time. If a duration is not set for either, they will both be 1 week.
func (g *Global) GetLeaseTime(req time.Duration, registered bool) time.Duration {
	// TODO: Clean this up
	if registered {
		if req == 0 {
			if g.RegisteredSettings.DefaultLeaseTime != 0 {
				return g.RegisteredSettings.DefaultLeaseTime
			}
			return g.Settings.DefaultLeaseTime
		}

		if g.RegisteredSettings.MaxLeaseTime != 0 {
			if req < g.RegisteredSettings.MaxLeaseTime {
				return req
			}
			return g.RegisteredSettings.MaxLeaseTime
		}

		if req < g.Settings.MaxLeaseTime {
			return req
		}
		return g.Settings.MaxLeaseTime
	} else {
		if req == 0 {
			if g.UnregisteredSettings.DefaultLeaseTime != 0 {
				return g.UnregisteredSettings.DefaultLeaseTime
			}
			return g.Settings.DefaultLeaseTime
		}

		if g.UnregisteredSettings.MaxLeaseTime != 0 {
			if req < g.UnregisteredSettings.MaxLeaseTime {
				return req
			}
			return g.UnregisteredSettings.MaxLeaseTime
		}

		if req < g.Settings.MaxLeaseTime {
			return req
		}
		return g.Settings.MaxLeaseTime
	}
}

func (g *Global) GetOptions(registered bool) dhcp4.Options {
	if registered && g.regOptionsCached {
		return g.RegisteredSettings.Options
	} else if !registered && g.unregOptionsCached {
		return g.UnregisteredSettings.Options
	}

	if registered {
		// Merge "global" settings into registered settings
		for c, v := range g.Settings.Options {
			if _, ok := g.RegisteredSettings.Options[c]; !ok {
				g.RegisteredSettings.Options[c] = v
			}
		}
		g.regOptionsCached = true
		return g.RegisteredSettings.Options
	}

	// Merge network "global" settings into unregistered settings
	for c, v := range g.Settings.Options {
		if _, ok := g.UnregisteredSettings.Options[c]; !ok {
			g.UnregisteredSettings.Options[c] = v
		}
	}
	g.unregOptionsCached = true
	return g.UnregisteredSettings.Options
}

func (g *Global) Print() {
	fmt.Println("\n---Global Configuration---")
	fmt.Printf("Server Identifier: %s\n", g.ServerIdentifier.String())
	fmt.Println("\n--Global Settings--")
	g.Settings.Print()
	fmt.Println("\n--Global Registered Settings--")
	g.RegisteredSettings.Print()
	fmt.Println("\n--Global Unregistered Settings--")
	g.UnregisteredSettings.Print()
}
