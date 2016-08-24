// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"errors"
	"net"
	"time"
)

// A Lease represents a single DHCP lease in a pool. It is bound to a particular
// pool and network.
type Lease struct {
	store       LeaseStore
	ID          int
	IP          net.IP
	MAC         net.HardwareAddr
	Network     string
	Start       time.Time
	End         time.Time
	Hostname    string
	IsAbandoned bool
	Offered     bool
	Registered  bool
}

func NewLease(s LeaseStore) *Lease {
	return &Lease{store: s}
}

// IsRegisteredByIP checks if an IP is leased to a registered MAC address.
// It will return false if an error occurs as well as the error itself.
func IsRegisteredByIP(s LeaseStore, ip net.IP) (bool, error) {
	lease, err := s.GetLeaseByIP(ip)
	if err != nil {
		return false, err
	}
	if lease.ID == 0 {
		return false, errors.New("No lease given for IP " + ip.String())
	}
	return lease.Registered, nil
}

// GetLeaseByMAC returns a Lease given the mac address. This method will always return
// a Lease. Make sure to check if error is nil. If a new lease object was created
// it will have an ID = 0. The lease returned will be the most recent least given
// to the provided MAC address.
func GetLeaseByMAC(s LeaseStore, mac net.HardwareAddr) (*Lease, error) {
	lease, err := s.GetRecentLeaseByMAC(mac)
	if lease == nil {
		lease = NewLease(s)
		lease.MAC = mac
	}
	return lease, err
}

// GetAllLeasesByMAC returns a slice of Lease given the mac address. If no leases
// exist, the slice will be nil.
func GetAllLeasesByMAC(s LeaseStore, mac net.HardwareAddr) ([]*Lease, error) {
	return s.GetAllLeasesByMAC(mac)
}

// GetLeaseByIP returns a Lease given the IP address. This method will always return
// a Lease. Make sure to check if error is nil. If a new lease object was created
// it will have an ID = 0.
func GetLeaseByIP(s LeaseStore, ip net.IP) (*Lease, error) {
	lease, err := s.GetLeaseByIP(ip)
	if lease == nil {
		lease = NewLease(s)
		lease.IP = ip
	}
	return lease, err
}

// GetAllLeases will return a slice of all leases in the database.
func GetAllLeases(s LeaseStore) ([]*Lease, error) {
	return s.GetAllLeases()
}

func SearchLeases(s LeaseStore, where string, vals ...interface{}) ([]*Lease, error) {
	return s.SearchLeases("WHERE "+where, vals...)
}

// IsFree determines if the lease is expired and available for use
func (l *Lease) IsFree() bool {
	return (l.ID == 0 || time.Now().After(l.End))
}

func (l *Lease) IsExpired() bool {
	return l.End.Before(time.Now())
}

func (l *Lease) Save() error {
	if l.ID == 0 {
		return l.insertLease()
	}
	return l.updateLease()
}

func (l *Lease) updateLease() error {
	return l.store.UpdateLease(l)
}

func (l *Lease) insertLease() error {
	return l.store.CreateLease(l)
}

func (l *Lease) Delete() error {
	return l.store.DeleteLease(l)
}
