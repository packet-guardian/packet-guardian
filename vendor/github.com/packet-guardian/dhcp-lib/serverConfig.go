// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"net"
)

type ServerConfig struct {
	LeaseStore LeaseStore
}

type LeaseStore interface {
	GetAllLeases() ([]*Lease, error)
	GetLeaseByIP(net.IP) (*Lease, error)

	GetRecentLeaseByMAC(net.HardwareAddr) (*Lease, error)
	GetAllLeasesByMAC(net.HardwareAddr) ([]*Lease, error)

	CreateLease(*Lease) error
	UpdateLease(*Lease) error
	DeleteLease(*Lease) error

	SearchLeases(string, ...interface{}) ([]*Lease, error)
}
