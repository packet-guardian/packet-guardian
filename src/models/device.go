// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"encoding/json"
	"net"
	"time"

	"github.com/packet-guardian/packet-guardian/src/common"
)

type DeviceStore interface {
	Save(*Device) error
	Delete(*Device) error
	DeleteAllDeviceForUser(u *User) error
}

type LeaseStore interface {
	GetLeaseHistory(net.HardwareAddr) ([]LeaseHistory, error)
	GetLatestLease(net.HardwareAddr) LeaseHistory
}

type LeaseHistory interface {
	GetIP() net.IP
	GetMAC() net.HardwareAddr
	GetNetworkName() string
	GetStartTime() time.Time
	GetEndTime() time.Time
}

type BlacklistItem interface {
	Blacklist()
	Unblacklist()
	IsBlacklisted(string) bool
	Save(string) error
}

// Device represents a device in the system
type Device struct {
	e              *common.Environment
	deviceStore    DeviceStore
	leaseStore     LeaseStore
	ID             int              `json:"id"`
	MAC            net.HardwareAddr `json:"mac"`
	Username       string           `json:"username"`
	Description    string           `json:"description"`
	RegisteredFrom net.IP           `json:"registered_from"`
	Platform       string           `json:"platform"`
	Expires        time.Time        `json:"-"`
	DateRegistered time.Time        `json:"-"`
	UserAgent      string           `json:"-"`
	blacklist      BlacklistItem
	LastSeen       time.Time      `json:"-"`
	Leases         []LeaseHistory `json:"-"`
}

func NewDevice(e *common.Environment, s DeviceStore, l LeaseStore, b BlacklistItem) *Device {
	return &Device{
		e:           e,
		deviceStore: s,
		leaseStore:  l,
		blacklist:   b,
	}
}

func (d *Device) MarshalJSON() ([]byte, error) {
	type Alias Device
	return json.Marshal(&struct {
		*Alias
		Expires        time.Time `json:"expires"`
		DateRegistered time.Time `json:"registered"`
		LastSeen       time.Time `json:"last_seen"`
		Blacklisted    bool      `json:"blacklisted"`
		MAC            string    `json:"mac"`
	}{
		Alias:          (*Alias)(d),
		Expires:        d.Expires.UTC(),
		DateRegistered: d.DateRegistered.UTC(),
		LastSeen:       d.LastSeen.UTC(),
		Blacklisted:    d.IsBlacklisted(),
		MAC:            d.MAC.String(),
	})
}

func (d *Device) GetID() int {
	return d.ID
}

func (d *Device) GetMAC() net.HardwareAddr {
	return d.MAC
}

func (d *Device) GetUsername() string {
	return d.Username
}

func (d *Device) IsBlacklisted() bool {
	return d.blacklist.IsBlacklisted(d.MAC.String())
}

func (d *Device) SetBlacklist(b bool) {
	if b {
		d.blacklist.Blacklist()
		return
	}
	d.blacklist.Unblacklist()
}

func (d *Device) IsRegistered() bool {
	return (d.ID != 0 && !d.IsBlacklisted() && !d.IsExpired())
}

func (d *Device) SetLastSeen(t time.Time) {
	d.LastSeen = t
}

func (d *Device) LoadLeaseHistory() error {
	if d.Leases != nil {
		return nil
	}

	leases, err := d.leaseStore.GetLeaseHistory(d.MAC)
	if err != nil {
		return err
	}
	d.Leases = leases
	return nil
}

// GetCurrentLease will return the last known lease for the device that has
// not expired. If two leases are currently active, it will return the lease
// with the newest start date. If no current lease is found, returns nil.
func (d *Device) GetCurrentLease() LeaseHistory {
	return d.leaseStore.GetLatestLease(d.MAC)
}

func (d *Device) IsExpired() bool {
	return d.Expires.Unix() > 10 && time.Now().After(d.Expires)
}

func (d *Device) SaveToBlacklist() error {
	return d.blacklist.Save(d.MAC.String())
}

func (d *Device) Save() error {
	if err := d.deviceStore.Save(d); err != nil {
		return err
	}
	return d.SaveToBlacklist()
}

func (d *Device) Delete() error {
	if err := d.deviceStore.Delete(d); err != nil {
		return err
	}
	return nil
}
