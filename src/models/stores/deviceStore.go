// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stores

import (
	"errors"
	"net"
	"time"

	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models"
)

var appDeviceStore DeviceStore

type DeviceStore interface {
	GetDeviceByMAC(mac net.HardwareAddr) (*models.Device, error)
	GetDeviceByID(id int) (*models.Device, error)
	GetFlaggedDevices() ([]*models.Device, error)
	GetDevicesForUser(u *models.User) ([]*models.Device, error)
	GetDevicesForUserPage(u *models.User, page int) ([]*models.Device, error)
	GetDeviceCountForUser(u *models.User) (int, error)
	GetAllDevices(e *common.Environment) ([]*models.Device, error)
	SearchDevicesByField(field, pattern string) ([]*models.Device, error)
	Save(d *models.Device) error
	Delete(d *models.Device) error
	DeleteAllDeviceForUser(u *models.User) error
}

type deviceStore struct {
	e *common.Environment
}

func newDeviceStore(e *common.Environment) *deviceStore {
	return &deviceStore{
		e: e,
	}
}

func GetDeviceStore(e *common.Environment) DeviceStore {
	if appDeviceStore == nil {
		appDeviceStore = newDeviceStore(e)
	}
	return appDeviceStore
}

func (s *deviceStore) GetDeviceByMAC(mac net.HardwareAddr) (*models.Device, error) {
	sql := `WHERE "mac" = ?`
	devices, err := s.getDevicesFromDatabase(sql, mac.String())
	if len(devices) == 0 {
		dev := models.NewDevice(s, GetLeaseStore(s.e), NewBlacklistItem(GetBlacklistStore(s.e)))
		dev.MAC = mac
		return dev, err
	}
	return devices[0], nil
}

func (s *deviceStore) GetDeviceByID(id int) (*models.Device, error) {
	sql := `WHERE "id" = ?`
	devices, err := s.getDevicesFromDatabase(sql, id)
	if len(devices) == 0 {
		return models.NewDevice(s, GetLeaseStore(s.e), NewBlacklistItem(GetBlacklistStore(s.e))), err
	}
	return devices[0], nil
}

func (s *deviceStore) GetFlaggedDevices() ([]*models.Device, error) {
	return s.getDevicesFromDatabase(`WHERE "flagged" = 1`)
}

func (s *deviceStore) GetDevicesForUser(u *models.User) ([]*models.Device, error) {
	sql := `WHERE "username" = ? ORDER BY "mac" ASC`
	return s.getDevicesFromDatabase(sql, u.Username)
}

func (s *deviceStore) GetDevicesForUserPage(u *models.User, page int) ([]*models.Device, error) {
	sql := `WHERE "username" = ? ORDER BY "mac" ASC LIMIT ?,?`
	return s.getDevicesFromDatabase(sql, u.Username, (common.PageSize*page)-common.PageSize, common.PageSize)
}

func (s *deviceStore) GetDeviceCountForUser(u *models.User) (int, error) {
	sql := `SELECT count(*) as "device_count" FROM "device" WHERE "username" = ?`
	row := s.e.DB.QueryRow(sql, u.Username)
	var deviceCount int
	err := row.Scan(&deviceCount)
	if err != nil {
		return 0, err
	}
	return deviceCount, nil
}

func (s *deviceStore) GetAllDevices(e *common.Environment) ([]*models.Device, error) {
	return s.getDevicesFromDatabase("")
}

func (s *deviceStore) SearchDevicesByField(field, pattern string) ([]*models.Device, error) {
	sql := `WHERE "` + field + `" LIKE ?`
	return s.getDevicesFromDatabase(sql, pattern)
}

func (s *deviceStore) getDevicesFromDatabase(where string, values ...interface{}) ([]*models.Device, error) {
	sql := `SELECT "id", "mac", "username", "registered_from", "platform", "expires", "date_registered", "user_agent", "description", "last_seen", "flagged" FROM "device" ` + where

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
		var flagged bool

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
			&flagged,
		)
		if err != nil {
			continue
		}

		mac, _ := net.ParseMAC(macStr)

		device := models.NewDevice(s, GetLeaseStore(s.e), NewBlacklistItem(GetBlacklistStore(s.e)))
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
		device.Flagged = flagged

		results = append(results, device)
	}
	return results, nil
}

func (s *deviceStore) Save(d *models.Device) error {
	if d.ID == 0 {
		return s.saveNew(d)
	}
	return s.updateExisting(d)
}

func (s *deviceStore) updateExisting(d *models.Device) error {
	sql := `UPDATE "device" SET "mac" = ?, "username" = ?, "registered_from" = ?, "platform" = ?, "expires" = ?, "date_registered" = ?, "user_agent" = ?, "description" = ?, "last_seen" = ?, "flagged" = ? WHERE "id" = ?`

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
		d.Flagged,
		d.ID,
	)
	if err != nil {
		return err
	}
	return d.SaveToBlacklist()
}

func (s *deviceStore) saveNew(d *models.Device) error {
	if d.Username == "" {
		return errors.New("Username cannot be empty")
	}

	sql := `INSERT INTO "device" ("mac", "username", "registered_from", "platform", "expires", "date_registered", "user_agent", "description", "last_seen", "flagged") VALUES (?,?,?,?,?,?,?,?,?,?)`

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
		d.Flagged,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	d.ID = int(id)
	return d.SaveToBlacklist()
}

func (s *deviceStore) Delete(d *models.Device) error {
	sql := `DELETE FROM "device" WHERE "id" = ?`
	_, err := s.e.DB.Exec(sql, d.ID)
	return err
}

func (s *deviceStore) DeleteAllDeviceForUser(u *models.User) error {
	sql := `DELETE FROM "device" WHERE "username" = ?`
	_, err := s.e.DB.Exec(sql, u.Username)
	return err
}
