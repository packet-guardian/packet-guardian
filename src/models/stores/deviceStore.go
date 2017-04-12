// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stores

import (
	"errors"
	"net"
	"time"

	"github.com/usi-lfkeitel/packet-guardian/src/common"
	"github.com/usi-lfkeitel/packet-guardian/src/models"
)

var deviceStore *DeviceStore

type DeviceStore struct {
	e *common.Environment
}

func newDeviceStore(e *common.Environment) *DeviceStore {
	return &DeviceStore{
		e: e,
	}
}

func GetDeviceStore(e *common.Environment) *DeviceStore {
	if deviceStore == nil || e.IsTesting() {
		deviceStore = newDeviceStore(e)
	}
	return deviceStore
}

func (s *DeviceStore) GetDeviceByMAC(mac net.HardwareAddr) (*models.Device, error) {
	sql := `WHERE "mac" = ?`
	devices, err := s.getDevicesFromDatabase(sql, mac.String())
	if len(devices) == 0 {
		dev := models.NewDevice(s.e, s, GetLeaseStore(s.e), NewBlacklistItem(GetBlacklistStore(s.e)))
		dev.MAC = mac
		return dev, err
	}
	return devices[0], nil
}

func (s *DeviceStore) GetDeviceByID(id int) (*models.Device, error) {
	sql := `WHERE "id" = ?`
	devices, err := s.getDevicesFromDatabase(sql, id)
	if len(devices) == 0 {
		return models.NewDevice(s.e, s, GetLeaseStore(s.e), NewBlacklistItem(GetBlacklistStore(s.e))), err
	}
	return devices[0], nil
}

func (s *DeviceStore) GetDevicesForUser(u *models.User) ([]*models.Device, error) {
	sql := `WHERE "username" = ? ORDER BY "mac"`
	if s.e.DB.Driver == "sqlite" {
		sql += " COLLATE NOCASE"
	}
	sql += " ASC"
	return s.getDevicesFromDatabase(sql, u.Username)
}

func (s *DeviceStore) GetDeviceCountForUser(u *models.User) (int, error) {
	sql := `SELECT count(*) as "device_count" FROM "device" WHERE "username" = ?`
	row := s.e.DB.QueryRow(sql, u.Username)
	var deviceCount int
	err := row.Scan(&deviceCount)
	if err != nil {
		return 0, err
	}
	return deviceCount, nil
}

func (s *DeviceStore) GetAllDevices(e *common.Environment) ([]*models.Device, error) {
	return s.getDevicesFromDatabase("")
}

func (s *DeviceStore) SearchDevicesByField(field, pattern string) ([]*models.Device, error) {
	sql := `WHERE "` + field + `" LIKE ?`
	return s.getDevicesFromDatabase(sql, pattern)
}

func (s *DeviceStore) getDevicesFromDatabase(where string, values ...interface{}) ([]*models.Device, error) {
	sql := `SELECT "id", "mac", "username", "registered_from", "platform", "expires", "date_registered", "user_agent", "description", "last_seen" FROM "device" ` + where

	rows, err := s.e.DB.Query(sql, values...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.Device
	for rows.Next() {
		var id int
		var macStr string
		var username string
		var registeredFrom string
		var platform string
		var expires int64
		var dateRegistered int64
		var ua string
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
			&description,
			&lastSeen,
		)
		if err != nil {
			continue
		}

		mac, _ := net.ParseMAC(macStr)

		device := models.NewDevice(s.e, s, GetLeaseStore(s.e), NewBlacklistItem(GetBlacklistStore(s.e)))
		device.ID = id
		device.MAC = mac
		device.Username = username
		device.Description = description
		device.RegisteredFrom = net.ParseIP(registeredFrom)
		device.Platform = platform
		device.Expires = time.Unix(expires, 0)
		device.DateRegistered = time.Unix(dateRegistered, 0)
		device.UserAgent = ua
		device.LastSeen = time.Unix(lastSeen, 0)

		results = append(results, device)
	}
	return results, nil
}

func (s *DeviceStore) Save(d *models.Device) error {
	if d.ID == 0 {
		return s.saveNew(d)
	}
	return s.updateExisting(d)
}

func (s *DeviceStore) updateExisting(d *models.Device) error {
	sql := `UPDATE "device" SET "mac" = ?, "username" = ?, "registered_from" = ?, "platform" = ?, "expires" = ?, "date_registered" = ?, "user_agent" = ?, "description" = ?, "last_seen" = ? WHERE "id" = ?`

	_, err := s.e.DB.Exec(
		sql,
		d.MAC.String(),
		d.Username,
		d.RegisteredFrom.String(),
		d.Platform,
		d.Expires.Unix(),
		d.DateRegistered.Unix(),
		d.UserAgent,
		d.Description,
		d.LastSeen.Unix(),
		d.ID,
	)
	if err != nil {
		return err
	}
	return d.SaveToBlacklist()
}

func (s *DeviceStore) saveNew(d *models.Device) error {
	if d.Username == "" {
		return errors.New("Username cannot be empty")
	}

	sql := `INSERT INTO "device" ("mac", "username", "registered_from", "platform", "expires", "date_registered", "user_agent", "description", "last_seen") VALUES (?,?,?,?,?,?,?,?,?)`

	result, err := s.e.DB.Exec(
		sql,
		d.MAC.String(),
		d.Username,
		d.RegisteredFrom.String(),
		d.Platform,
		d.Expires.Unix(),
		d.DateRegistered.Unix(),
		d.UserAgent,
		d.Description,
		d.LastSeen.Unix(),
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	d.ID = int(id)
	return d.SaveToBlacklist()
}

func (s *DeviceStore) Delete(d *models.Device) error {
	sql := `DELETE FROM "device" WHERE "id" = ?`
	_, err := s.e.DB.Exec(sql, d.ID)
	return err
}

func (s *DeviceStore) DeleteAllDeviceForUser(u *models.User) error {
	sql := `DELETE FROM "device" WHERE "username" = ?`
	_, err := s.e.DB.Exec(sql, u.Username)
	return err
}
