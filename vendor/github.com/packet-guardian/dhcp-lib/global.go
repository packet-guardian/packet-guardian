// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"net"
)

type global struct {
	serverIdentifier     net.IP
	settings             *settings
	registeredSettings   *settings
	unregisteredSettings *settings
}

func newGlobal() *global {
	return &global{
		settings:             newSettingsBlock(),
		registeredSettings:   newSettingsBlock(),
		unregisteredSettings: newSettingsBlock(),
	}
}
