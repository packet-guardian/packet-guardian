package models

import (
	"net"
	"time"

	"github.com/onesimus-systems/packet-guardian/src/common"
)

// Device represents a device in the system
type Device struct {
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
	return &Device{}
}

func GetDeviceByMAC(e *common.Environment, mac net.HardwareAddr) (*Device, error) {
	sql := "WHERE \"mac\" = ?"
	devices, err := getDevicesFromDatabase(e, sql, mac.String())
	if err != nil {
		return NewDevice(e), err
	}
	return devices[0], nil
}

func GetDeviceByID(e *common.Environment, id int) (*Device, error) {
	sql := "WHERE \"id\" = ?"
	devices, err := getDevicesFromDatabase(e, sql, id)
	if err != nil {
		return NewDevice(e), err
	}
	return devices[0], nil
}

func GetDevicesForUser(e *common.Environment, u *User) ([]*Device, error) {
	sql := "WHERE \"username\" = ?"
	return getDevicesFromDatabase(e, sql, u.Username)
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
