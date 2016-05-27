// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"fmt"
	"net"
	"time"

	"github.com/onesimus-systems/dhcp4"
)

type global struct {
	serverIdentifier     net.IP
	settings             *settings
	registeredSettings   *settings
	regOptionsCached     bool
	unregisteredSettings *settings
	unregOptionsCached   bool
}

func newGlobal() *global {
	return &global{
		settings:             newSettingsBlock(),
		registeredSettings:   newSettingsBlock(),
		unregisteredSettings: newSettingsBlock(),
	}
}

// GetLeaseTime returns the lease time given the requested time req and if the client is registered.
// If req is 0 then the default lease time is returned. Otherwise it will return the lower of
// req and the maximum lease time. If a duration is not set for either, they will both be 1 week.
func (g *global) getLeaseTime(req time.Duration, registered bool) time.Duration {
	// TODO: Clean this up
	if registered {
		if req == 0 {
			if g.registeredSettings.defaultLeaseTime != 0 {
				return g.registeredSettings.defaultLeaseTime
			}
			return g.settings.defaultLeaseTime
		}

		if g.registeredSettings.maxLeaseTime != 0 {
			if req < g.registeredSettings.maxLeaseTime {
				return req
			}
			return g.registeredSettings.maxLeaseTime
		}

		if req < g.settings.maxLeaseTime {
			return req
		}
		return g.settings.maxLeaseTime
	}

	if req == 0 {
		if g.unregisteredSettings.defaultLeaseTime != 0 {
			return g.unregisteredSettings.defaultLeaseTime
		}
		return g.settings.defaultLeaseTime
	}

	if g.unregisteredSettings.maxLeaseTime != 0 {
		if req < g.unregisteredSettings.maxLeaseTime {
			return req
		}
		return g.unregisteredSettings.maxLeaseTime
	}

	if req < g.settings.maxLeaseTime {
		return req
	}
	return g.settings.maxLeaseTime
}

func (g *global) getOptions(registered bool) dhcp4.Options {
	if registered && g.regOptionsCached {
		return g.registeredSettings.options
	} else if !registered && g.unregOptionsCached {
		return g.unregisteredSettings.options
	}

	if registered {
		// Merge "global" settings into registered settings
		for c, v := range g.settings.options {
			if _, ok := g.registeredSettings.options[c]; !ok {
				g.registeredSettings.options[c] = v
			}
		}
		g.regOptionsCached = true
		return g.registeredSettings.options
	}

	// Merge network "global" settings into unregistered settings
	for c, v := range g.settings.options {
		if _, ok := g.unregisteredSettings.options[c]; !ok {
			g.unregisteredSettings.options[c] = v
		}
	}
	g.unregOptionsCached = true
	return g.unregisteredSettings.options
}

func (g *global) print() {
	fmt.Println("\n---Global Configuration---")
	fmt.Printf("Server Identifier: %s\n", g.serverIdentifier.String())
	fmt.Println("\n--Global Settings--")
	g.settings.Print()
	fmt.Println("\n--Global Registered Settings--")
	g.registeredSettings.Print()
	fmt.Println("\n--Global Unregistered Settings--")
	g.unregisteredSettings.Print()
}
