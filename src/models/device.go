package models

import (
	"errors"
	"net"
	"time"

	"github.com/onesimus-systems/packet-guardian/src/common"
)

// Device represents a device in the system
type Device struct {
	e              *common.Environment
	ID             int
	MAC            net.HardwareAddr
	Username       string
	RegisteredFrom net.IP
	Platform       string
	Expires        time.Time
	DateRegistered time.Time
	UserAgent      string
	Blacklisted    bool
}

func NewDevice(e *common.Environment) *Device {
	return &Device{e: e}
}

func GetDeviceByMAC(e *common.Environment, mac net.HardwareAddr) (*Device, error) {
	sql := "WHERE \"mac\" = ?"
	devices, err := getDevicesFromDatabase(e, sql, mac.String())
	if err != nil {
		return NewDevice(e), err
	}
	if len(devices) == 0 {
		dev := NewDevice(e)
		dev.MAC = mac
		return dev, nil
	}
	return devices[0], nil
}

func GetDeviceByID(e *common.Environment, id int) (*Device, error) {
	sql := "WHERE \"id\" = ?"
	devices, err := getDevicesFromDatabase(e, sql, id)
	if err != nil {
		return NewDevice(e), err
	}
	if len(devices) == 0 {
		return NewDevice(e), nil
	}
	return devices[0], nil
}

func GetDevicesForUser(e *common.Environment, u *User) ([]*Device, error) {
	sql := "WHERE \"username\" = ?"
	return getDevicesFromDatabase(e, sql, u.Username)
}

func GetDeviceCountForUser(e *common.Environment, u *User) (int, error) {
	sql := "SELECT count(*) as \"device_count\" FROM \"device\" WHERE \"username\" = ?"
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

func getDevicesFromDatabase(e *common.Environment, where string, values ...interface{}) ([]*Device, error) {
	sql := "SELECT \"id\", \"mac\", \"username\", \"registered_from\", \"platform\", \"expires\", \"date_registered\", \"user_agent\", \"blacklisted\" " + where

	rows, err := e.DB.Query(sql, values)
	if err != nil {
		return nil, err
	}

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
			RegisteredFrom: net.ParseIP(registeredFrom),
			Platform:       platform,
			Expires:        time.Unix(expires, 0),
			DateRegistered: time.Unix(dateRegistered, 0),
			UserAgent:      ua,
			Blacklisted:    blacklisted,
		}
		results = append(results, device)
	}
	return results, nil
}

func DeleteAllDeviceForUser(e *common.Environment, u *User) error {
	sql := "DELETE FROM \"device\" WHERE \"username\" = ?"
	_, err := e.DB.Exec(sql, u.Username)
	return err
}

func (d *Device) Save() error {
	if d.ID == 0 {
		return d.saveNew()
	}
	return d.updateExisting()
}

func (d *Device) updateExisting() error {
	sql := "UPDATE \"device\" SET \"mac\" = ?, \"username\" = ?, \"registered_from\" = ?, \"platform\" = ?, \"expires\" = ?, \"date_registered\" = ?, \"user_agent\" = ?, \"blacklisted\" = ? WHERE \"id\" = ?"

	_, err := d.e.DB.Exec(
		sql,
		d.MAC.String(),
		d.Username,
		d.RegisteredFrom.String(),
		d.Platform,
		d.Expires.Unix(),
		d.DateRegistered.Unix(),
		d.UserAgent,
		d.Blacklisted,
		d.ID,
	)
	return err
}

func (d *Device) saveNew() error {
	if d.Username == "" {
		return errors.New("Username cannot be empty")
	}

	sql := "INSERT INTO \"device\" (\"mac\", \"username\", \"registered_from\", \"platform\", \"expires\", \"date_registered\", \"user_agent\", \"blacklisted\") VALUES (?,?,?,?,?,?,?,?)"

	_, err := d.e.DB.Exec(
		sql,
		d.MAC.String(),
		d.Username,
		d.RegisteredFrom.String(),
		d.Platform,
		d.Expires.Unix(),
		d.DateRegistered.Unix(),
		d.UserAgent,
		d.Blacklisted,
	)
	return err
}

func (d *Device) Delete() error {
	sql := "DELETE FROM \"device\" WHERE \"id\" = ?"
	_, err := d.e.DB.Exec(sql, d.ID)
	return err
}
