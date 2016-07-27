// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"errors"
	"net"
	"time"

	"github.com/usi-lfkeitel/packet-guardian/src/common"
)

// Device represents a device in the system
type Device struct {
	e              *common.Environment
	ID             int
	MAC            net.HardwareAddr
	Username       string
	Description    string
	RegisteredFrom net.IP
	Platform       string
	Expires        time.Time
	DateRegistered time.Time
	UserAgent      string
	IsBlacklisted  bool
	LastSeen       time.Time
	Leases         []*Lease
}

func NewDevice(e *common.Environment) *Device {
	return &Device{e: e}
}

func GetDeviceByMAC(e *common.Environment, mac net.HardwareAddr) (*Device, error) {
	sql := "WHERE \"mac\" = ?"
	devices, err := getDevicesFromDatabase(e, sql, mac.String())
	if devices == nil || len(devices) == 0 {
		dev := NewDevice(e)
		dev.MAC = mac
		return dev, err
	}
	return devices[0], nil
}

func GetDeviceByID(e *common.Environment, id int) (*Device, error) {
	sql := "WHERE \"id\" = ?"
	devices, err := getDevicesFromDatabase(e, sql, id)
	if devices == nil || len(devices) == 0 {
		return NewDevice(e), err
	}
	return devices[0], nil
}

func GetDevicesForUser(e *common.Environment, u *User) ([]*Device, error) {
	sql := "WHERE \"username\" = ? ORDER BY \"mac\" COLLATE NOCASE ASC"
	return getDevicesFromDatabase(e, sql, u.Username)
}

func GetDeviceCountForUser(e *common.Environment, u *User) (int, error) {
	sql := `SELECT count(*) as "device_count" FROM "device" WHERE "username" = ?`
	row := e.DB.QueryRow(sql, u.Username)
	var deviceCount int
	err := row.Scan(&deviceCount)
	if err != nil {
		return 0, err
	}
	return deviceCount, nil
}

func GetAllDevices(e *common.Environment) ([]*Device, error) {
	return getDevicesFromDatabase(e, "")
}

func SearchDevicesByField(e *common.Environment, field, pattern string) ([]*Device, error) {
	sql := "WHERE \"" + field + "\" LIKE ?"
	return getDevicesFromDatabase(e, sql, pattern)
}

func getDevicesFromDatabase(e *common.Environment, where string, values ...interface{}) ([]*Device, error) {
	sql := `SELECT "id", "mac", "username", "registered_from", "platform", "expires", "date_registered", "user_agent", "blacklisted", "description", "last_seen" FROM "device" ` + where

	rows, err := e.DB.Query(sql, values...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*Device
	for rows.Next() {
		var id int
		var macStr string
		var username string
		var registeredFrom string
		var platform string
		var expires int64
		var dateRegistered int64
		var ua string
		var blacklisted bool
		var description string
		var lastSeen int64

		err := rows.Scan(
			&id,
			&macStr,
			&username,
			&registeredFrom,
			&platform,
			&expires,
			&dateRegistered,
			&ua,
			&blacklisted,
			&description,
			&lastSeen,
		)
		if err != nil {
			continue
		}

		mac, _ := net.ParseMAC(macStr)

		device := &Device{
			e:              e,
			ID:             id,
			MAC:            mac,
			Username:       username,
			Description:    description,
			RegisteredFrom: net.ParseIP(registeredFrom),
			Platform:       platform,
			Expires:        time.Unix(expires, 0),
			DateRegistered: time.Unix(dateRegistered, 0),
			UserAgent:      ua,
			IsBlacklisted:  blacklisted,
			LastSeen:       time.Unix(lastSeen, 0),
		}
		results = append(results, device)
	}
	return results, nil
}

func DeleteAllDeviceForUser(e *common.Environment, u *User) error {
	sql := `DELETE FROM "device" WHERE "username" = ?`
	_, err := e.DB.Exec(sql, u.Username)
	return err
}

func (d *Device) LoadKnownLeases() error {
	leases, err := SearchLeases(
		d.e,
		`"mac" = ? ORDER BY "start" DESC`,
		d.MAC.String(),
	)
	if err != nil {
		return err
	}
	if leases != nil {
		d.Leases = leases
	}
	return nil
}

// GetCurrentLease will return the last known lease for the device that has
// not expired. If two leases are currently active, it will return the lease
// with the newest start date. If no current lease is found, returns nil.
func (d *Device) GetCurrentLease() *Lease {
	if d.Leases == nil {
		d.LoadKnownLeases()
	}
	if d.Leases == nil || len(d.Leases) == 0 {
		return nil
	}
	if d.Leases[0].End.Before(time.Now()) {
		return nil
	}
	return d.Leases[0]
}

func (d *Device) IsExpired() bool {
	return d.Expires.Unix() > 10 && time.Now().After(d.Expires)
}

func (d *Device) Save() error {
	if d.ID == 0 {
		return d.saveNew()
	}
	return d.updateExisting()
}

func (d *Device) updateExisting() error {
	sql := `UPDATE "device" SET "mac" = ?, "username" = ?, "registered_from" = ?, "platform" = ?, "expires" = ?, "date_registered" = ?, "user_agent" = ?, "blacklisted" = ?, "description" = ?, "last_seen" = ? WHERE "id" = ?`

	_, err := d.e.DB.Exec(
		sql,
		d.MAC.String(),
		d.Username,
		d.RegisteredFrom.String(),
		d.Platform,
		d.Expires.Unix(),
		d.DateRegistered.Unix(),
		d.UserAgent,
		d.IsBlacklisted,
		d.Description,
		d.LastSeen.Unix(),
		d.ID,
	)
	return err
}

func (d *Device) saveNew() error {
	if d.Username == "" {
		return errors.New("Username cannot be empty")
	}

	sql := `INSERT INTO "device" ("mac", "username", "registered_from", "platform", "expires", "date_registered", "user_agent", "blacklisted", "description", "last_seen") VALUES (?,?,?,?,?,?,?,?,?,?)`

	result, err := d.e.DB.Exec(
		sql,
		d.MAC.String(),
		d.Username,
		d.RegisteredFrom.String(),
		d.Platform,
		d.Expires.Unix(),
		d.DateRegistered.Unix(),
		d.UserAgent,
		d.IsBlacklisted,
		d.Description,
		d.LastSeen.Unix(),
	)
	id, _ := result.LastInsertId()
	d.ID = int(id)
	return err
}

func (d *Device) Delete() error {
	sql := `DELETE FROM "device" WHERE "id" = ?`
	_, err := d.e.DB.Exec(sql, d.ID)
	return err
}
