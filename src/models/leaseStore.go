// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"net"
	"time"

	"github.com/lfkeitel/verbose"
	"github.com/usi-lfkeitel/packet-guardian/src/common"
	"github.com/usi-lfkeitel/pg-dhcp"
)

type LeaseStore struct {
	e        *common.Environment
	histChan chan *dhcp.Lease
}

var leaseStore *LeaseStore

// NewLeaseStore will create a new LeaseStore object using the given Environment.
// Client code should use GetLeaseStore unless it's absolutly necessary to have
// a new LeaseStore object.
func NewLeaseStore(e *common.Environment) *LeaseStore {
	histChan := make(chan *dhcp.Lease, 20)
	go addToLeaseHistory(e, histChan)

	return &LeaseStore{
		e:        e,
		histChan: histChan,
	}
}

// GetLeaseStore will return an existing LeaseStore if one has been made already,
// or it will create a new one and return it. Client code should use this unless
// it's required to get a new LeaseStore object. If the environment is testing,
// it will always return a new store.
func GetLeaseStore(e *common.Environment) *LeaseStore {
	if e.IsTesting() {
		return NewLeaseStore(e)
	}
	if leaseStore != nil {
		return leaseStore
	}
	leaseStore = NewLeaseStore(e)
	return leaseStore
}

func (l *LeaseStore) GetAllLeases() ([]*dhcp.Lease, error) {
	return l.doDatabaseQuery("")
}

func (l *LeaseStore) GetLeaseByIP(ip net.IP) (*dhcp.Lease, error) {
	sql := `WHERE "ip" = ?`
	leases, err := l.doDatabaseQuery(sql, ip.String())
	if leases == nil || len(leases) == 0 {
		lease := dhcp.NewLease(l)
		lease.IP = ip
		return lease, err
	}
	return leases[0], nil
}

func (l *LeaseStore) GetRecentLeaseByMAC(mac net.HardwareAddr) (*dhcp.Lease, error) {
	sql := `WHERE "mac" = ? ORDER BY "start" DESC`
	leases, err := l.doDatabaseQuery(sql, mac.String())
	if leases == nil || len(leases) == 0 {
		lease := dhcp.NewLease(l)
		lease.MAC = mac
		return lease, err
	}
	return leases[0], nil
}

func (l *LeaseStore) GetAllLeasesByMAC(mac net.HardwareAddr) ([]*dhcp.Lease, error) {
	return l.doDatabaseQuery(`WHERE "mac" = ?`, mac.String())
}

func (l *LeaseStore) CreateLease(lease *dhcp.Lease) error {
	sql := `INSERT INTO "lease" ("ip", "mac", "network", "start", "end", "hostname", "abandoned", "registered") VALUES (?,?,?,?,?,?,?,?)`

	result, err := l.e.DB.Exec(
		sql,
		lease.IP.String(),
		lease.MAC.String(),
		lease.Network,
		lease.Start.Unix(),
		lease.End.Unix(),
		lease.Hostname,
		lease.IsAbandoned,
		lease.Registered,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	lease.ID = int(id)
	l.histChan <- lease
	return nil
}

func (l *LeaseStore) UpdateLease(lease *dhcp.Lease) error {
	sql := `UPDATE "lease" SET "mac" = ?, "start" = ?, "end" = ?, "hostname" = ?, "abandoned" = ? WHERE "id" = ?`

	_, err := l.e.DB.Exec(
		sql,
		lease.MAC.String(),
		lease.Start.Unix(),
		lease.End.Unix(),
		lease.Hostname,
		lease.IsAbandoned,
		lease.ID,
	)
	if err != nil {
		return err
	}
	l.histChan <- lease
	return nil
}

func (l *LeaseStore) DeleteLease(lease *dhcp.Lease) error {
	if lease.ID == 0 {
		return nil
	}

	sql := `DELETE FROM "lease" WHERE "id" = ?`
	_, err := l.e.DB.Exec(sql, lease.ID)
	return err
}

func (l *LeaseStore) SearchLeases(where string, vals ...interface{}) ([]*dhcp.Lease, error) {
	return l.doDatabaseQuery("WHERE "+where, vals...)
}

func (l *LeaseStore) doDatabaseQuery(where string, values ...interface{}) ([]*dhcp.Lease, error) {
	sql := `SELECT "id", "ip", "mac", "network", "start", "end", "hostname", "abandoned", "registered" FROM "lease" ` + where

	rows, err := l.e.DB.Query(sql, values...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*dhcp.Lease
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
			l.e.Log.WithFields(verbose.Fields{
				"error":   err,
				"package": "models:leasestore",
			}).Error("Failed to scan lease into struct")
			continue
		}

		mac, _ := net.ParseMAC(macStr)

		lease := dhcp.NewLease(l)
		lease.ID = id
		lease.IP = net.ParseIP(ip)
		lease.MAC = mac
		lease.Network = network
		lease.Start = time.Unix(start, 0)
		lease.End = time.Unix(end, 0)
		lease.Hostname = hostname
		lease.IsAbandoned = isAbandoned
		lease.Registered = registered
		results = append(results, lease)
	}
	return results, nil
}
