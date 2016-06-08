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

// A Lease represents a single DHCP lease in a pool. It is bound to a particular
// pool and network.
type Lease struct {
	e           *common.Environment
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

func NewLease(e *common.Environment) *Lease {
	return &Lease{e: e}
}

// IsRegisteredByIP checks if an IP is leased to a registered MAC address.
// It will return false if an error occurs as well as the error itself.
func IsRegisteredByIP(e *common.Environment, ip net.IP) (bool, error) {
	lease, err := GetLeaseByIP(e, ip)
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
// it will have an ID = 0.
func GetLeaseByMAC(e *common.Environment, mac net.HardwareAddr) (*Lease, error) {
	sql := "WHERE \"mac\" = ?"
	leases, err := getLeasesFromDatabase(e, sql, mac.String())
	if leases == nil || len(leases) == 0 {
		lease := NewLease(e)
		lease.MAC = mac
		return lease, err
	}
	return leases[0], nil
}

// GetAllLeasesByMAC returns a slice of Lease given the mac address. If no leases
// exist, the slice will be nil.
func GetAllLeasesByMAC(e *common.Environment, mac net.HardwareAddr) ([]*Lease, error) {
	sql := "WHERE \"mac\" = ?"
	return getLeasesFromDatabase(e, sql, mac.String())
}

// GetLeaseByIP returns a Lease given the IP address. This method will always return
// a Lease. Make sure to check if error is nil. If a new lease object was created
// it will have an ID = 0.
func GetLeaseByIP(e *common.Environment, ip net.IP) (*Lease, error) {
	sql := "WHERE \"ip\" = ?"
	leases, err := getLeasesFromDatabase(e, sql, ip.String())
	if leases == nil || len(leases) == 0 {
		lease := NewLease(e)
		lease.IP = ip
		return lease, err
	}
	return leases[0], nil
}

// GetAllLeases will return a slice of all leases in the database.
func GetAllLeases(e *common.Environment) ([]*Lease, error) {
	return getLeasesFromDatabase(e, "")
}

func SearchLeases(e *common.Environment, where string, vals ...interface{}) ([]*Lease, error) {
	return getLeasesFromDatabase(e, where, vals...)
}

func getLeasesFromDatabase(e *common.Environment, where string, values ...interface{}) ([]*Lease, error) {
	sql := `SELECT "id", "ip", "mac", "network", "start", "end", "hostname", "abandoned", "registered" FROM "lease" ` + where

	rows, err := e.DB.Query(sql, values...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*Lease
	for rows.Next() {
		var id int
		var ip string
		var macStr string
		var network string
		var start int64
		var end int64
		var hostname string
		var isAbandoned bool
		var registered bool

		err := rows.Scan(
			&id,
			&ip,
			&macStr,
			&network,
			&start,
			&end,
			&hostname,
			&isAbandoned,
			&registered,
		)
		if err != nil {
			e.Log.WithField("Err", err).Error("Failed to scan lease into struct")
			continue
		}

		mac, _ := net.ParseMAC(macStr)

		lease := &Lease{
			e:           e,
			ID:          id,
			IP:          net.ParseIP(ip),
			MAC:         mac,
			Network:     network,
			Start:       time.Unix(start, 0),
			End:         time.Unix(end, 0),
			Hostname:    hostname,
			IsAbandoned: isAbandoned,
			Registered:  registered,
		}
		results = append(results, lease)
	}
	return results, nil
}

// IsFree determines if the lease is expired and available for use
func (l *Lease) IsFree() bool {
	return (l.ID == 0 || time.Now().After(l.End))
}

func (l *Lease) Save() error {
	if l.ID == 0 {
		return l.insertLease()
	}
	return l.updateLease()
}

func (l *Lease) updateLease() error {
	sql := `UPDATE "lease" SET "mac" = ?, "start" = ?, "end" = ?, "hostname" = ?, "abandoned" = ? WHERE "id" = ?`

	_, err := l.e.DB.Exec(
		sql,
		l.MAC.String(),
		l.Start.Unix(),
		l.End.Unix(),
		l.Hostname,
		l.IsAbandoned,
		l.ID,
	)
	return err
}

func (l *Lease) insertLease() error {
	sql := `INSERT INTO "lease" ("ip", "mac", "network", "start", "end", "hostname", "abandoned", "registered") VALUES (?,?,?,?,?,?,?,?)`

	result, err := l.e.DB.Exec(
		sql,
		l.IP.String(),
		l.MAC.String(),
		l.Network,
		l.Start.Unix(),
		l.End.Unix(),
		l.Hostname,
		l.IsAbandoned,
		l.Registered,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	l.ID = int(id)
	return nil
}

func (l *Lease) Delete() error {
	if l.ID == 0 {
		return nil
	}

	sql := `DELETE FROM "lease" WHERE "id" = ?`
	_, err := l.e.DB.Exec(sql, l.ID)
	return err
}
