// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"net"
	"time"

	"github.com/lfkeitel/verbose"
)

type Environment string

const (
	EnvTesting Environment = "testing"
	EnvDev     Environment = "dev"
	EnvProd    Environment = "prod"
)

type ServerConfig struct {
	LeaseStore  LeaseStore
	DeviceStore DeviceStore
	Log         *verbose.Logger
	LogPath     string
	Env         Environment
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

type DeviceStore interface {
	GetDeviceByMAC(net.HardwareAddr) (Device, error)
}

type Device interface {
	SetLastSeen(time.Time)
	GetID() int
	GetMAC() net.HardwareAddr
	GetUsername() string
	IsBlacklisted() bool
	IsExpired() bool
	IsRegistered() bool
	Save() error
}

func (s *ServerConfig) IsTesting() bool {
	return (s.Env == EnvTesting)
}

func (s *ServerConfig) IsProd() bool {
	return (s.Env == EnvProd)
}

func (s *ServerConfig) IsDev() bool {
	return (s.Env == EnvDev)
}
