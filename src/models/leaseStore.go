// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"net"

	"github.com/usi-lfkeitel/packet-guardian/src/common"
	"github.com/usi-lfkeitel/pg-dhcp"
)

type LeaseStore struct {
	e *common.Environment
}

func NewLeaseStore(e *common.Environment) *LeaseStore {
	return &LeaseStore{e: e}
}

func (l *LeaseStore) GetAllLeases() ([]*dhcp.Lease, error) {
	return nil, nil
}
func (l *LeaseStore) GetLeaseByIP(net.IP) (*dhcp.Lease, error) {
	return nil, nil
}

func (l *LeaseStore) GetRecentLeaseByMAC(net.HardwareAddr) (*dhcp.Lease, error) {
	return nil, nil
}
func (l *LeaseStore) GetAllLeasesByMAC(net.HardwareAddr) ([]*dhcp.Lease, error) {
	return nil, nil
}

func (l *LeaseStore) CreateLease(*dhcp.Lease) error {
	return nil
}
func (l *LeaseStore) UpdateLease(*dhcp.Lease) error {
	return nil
}
func (l *LeaseStore) DeleteLease(*dhcp.Lease) error {
	return nil
}

func (l *LeaseStore) SearchLeases(string, ...interface{}) ([]*dhcp.Lease, error) {
	return nil, nil
}

func (l *LeaseStore) doDatabaseQuery() {

}
