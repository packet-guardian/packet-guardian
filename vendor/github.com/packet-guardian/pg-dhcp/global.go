// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"fmt"
	"net"
	"time"
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
	if req <= 0 { // Default lease time
		if registered && g.registeredSettings.defaultLeaseTime > 0 {
			return g.registeredSettings.defaultLeaseTime
		}
		if !registered && g.unregisteredSettings.defaultLeaseTime > 0 {
			return g.unregisteredSettings.defaultLeaseTime
		}
		return g.settings.defaultLeaseTime
	}

	// Client requested specific lease time
	if registered {
		// Client's request is less than or equal to max
		if g.registeredSettings.maxLeaseTime > 0 {
			if req <= g.registeredSettings.maxLeaseTime {
				return req
			}
			return g.registeredSettings.maxLeaseTime
		}

		// Fallback to truly global settings
		if req <= g.settings.maxLeaseTime {
			return req
		}
		return g.settings.maxLeaseTime
	}

	// maxLeaseTime for unregistered
	// Client's request is less than or equal to max
	if g.unregisteredSettings.maxLeaseTime > 0 {
		if req <= g.unregisteredSettings.maxLeaseTime {
			return req
		}
		return g.unregisteredSettings.maxLeaseTime
	}

	// Fallback to truly global settings
	if req <= g.settings.maxLeaseTime {
		return req
	}
	return g.settings.maxLeaseTime
}

func (g *global) getSettings(registered bool) *settings {
	if registered && g.regOptionsCached {
		return g.registeredSettings
	} else if !registered && g.unregOptionsCached {
		return g.unregisteredSettings
	}

	if registered {
		// Merge "global" settings into registered settings
		mergeSettings(g.registeredSettings, g.settings)
		g.regOptionsCached = true
		return g.registeredSettings
	}

	// Merge network "global" settings into unregistered settings
	mergeSettings(g.unregisteredSettings, g.settings)
	g.unregOptionsCached = true
	return g.unregisteredSettings
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
