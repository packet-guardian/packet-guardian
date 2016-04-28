package dhcp

import (
	"net"
	"time"

	"github.com/onesimus-systems/packet-guardian/src/common"
)

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
	Pool        *Pool
}

func NewLease(e *common.Environment) *Lease {
	return &Lease{e: e}
}

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

func GetAllLeases(e *common.Environment) ([]*Lease, error) {
	return getLeasesFromDatabase(e, "")
}

func getLeasesFromDatabase(e *common.Environment, where string, values ...interface{}) ([]*Lease, error) {
	sql := `SELECT "id", "ip", "mac", "network", "start", "end", "hostname", "abandoned" FROM "lease" ` + where

	rows, err := e.DB.Query(sql, values...)
	if err != nil {
		return nil, err
	}

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

		err := rows.Scan(
			&id,
			&ip,
			&macStr,
			&network,
			&start,
			&end,
			&hostname,
			&isAbandoned,
		)
		if err != nil {
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
		}
		results = append(results, lease)
	}
	return results, nil
}

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
	sql := `UPDATE "lease" SET "ip" = ?, "mac" = ?, "network" = ?, "start" = ?, "end" = ?, "hostname" = ?, "abandoned" = ? WHERE "id" = ?`

	_, err := l.e.DB.Exec(
		sql,
		l.IP.String(),
		l.MAC.String(),
		l.Network,
		l.Start.Unix(),
		l.End.Unix(),
		l.Hostname,
		l.IsAbandoned,
		l.ID,
	)
	return err
}

func (l *Lease) insertLease() error {
	sql := `INSERT INTO "lease" ("ip", "mac", "network", "start", "end", "hostname", "abandoned") VALUES (?,?,?,?,?,?,?)`

	result, err := l.e.DB.Exec(
		sql,
		l.IP.String(),
		l.MAC.String(),
		l.Network,
		l.Start.Unix(),
		l.End.Unix(),
		l.Hostname,
		l.IsAbandoned,
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
