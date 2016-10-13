package models

import (
	"database/sql"
	"net"
	"time"

	"github.com/lfkeitel/verbose"
	"github.com/usi-lfkeitel/packet-guardian/src/common"
	dhcp "github.com/usi-lfkeitel/pg-dhcp"
)

type LeaseHistory struct {
	ID      int
	IP      net.IP
	MAC     net.HardwareAddr
	Network string
	Start   time.Time
	End     time.Time
}

func GetLeaseHistory(e *common.Environment, mac net.HardwareAddr) ([]*LeaseHistory, error) {
	if !e.Config.Leases.HistoryEnabled {
		return make([]*LeaseHistory, 0), nil
	}
	stmt := `SELECT "id", "ip", "network", "start", "end" FROM "lease_history" WHERE "mac" = ?`

	rows, err := e.DB.Query(stmt, mac.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*LeaseHistory
	for rows.Next() {
		var id int
		var ip string
		var network string
		var start int64
		var end int64

		err := rows.Scan(
			&id,
			&ip,
			&network,
			&start,
			&end,
		)
		if err != nil {
			e.Log.WithFields(verbose.Fields{
				"error":   err,
				"package": "models:leasehistory",
			}).Error("Failed to scan lease into struct")
			continue
		}

		lease := &LeaseHistory{}
		lease.ID = id
		lease.IP = net.ParseIP(ip)
		lease.MAC = mac
		lease.Network = network
		lease.Start = time.Unix(start, 0)
		lease.End = time.Unix(end, 0)
		results = append(results, lease)
	}
	return results, nil
}

func getActiveLeaseHistory(e *common.Environment, mac net.HardwareAddr, ip net.IP) (*LeaseHistory, error) {
	stmt := `SELECT "id" FROM "lease_history" WHERE "mac" = ? AND ip = ? AND "start" <= ? AND "end" >= ? ORDER BY "start" DESC LIMIT 1`

	now := time.Now().Unix()
	row := e.DB.QueryRow(stmt, mac.String(), ip.String(), now, now)
	var id int

	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	lease := &LeaseHistory{}
	lease.ID = id
	return lease, nil
}

// Start this a goroutine and send leases on the channel to add them to history
func addToLeaseHistory(e *common.Environment, leaseChan <-chan *dhcp.Lease) {
	for lease := range leaseChan {
		if !e.Config.Leases.HistoryEnabled {
			continue
		}
		if err := processLease(e, lease); err != nil {
			e.Log.WithFields(verbose.Fields{
				"error":   err,
				"package": "models:leaseHistory",
			}).Error("Error saving lease history")
		}
	}
}

func processLease(e *common.Environment, lease *dhcp.Lease) error {
	leaseHist, err := getActiveLeaseHistory(e, lease.MAC, lease.IP)
	if err != nil {
		return err
	}
	if leaseHist == nil {
		return insertLeaseHistory(e, lease)
	}
	return updateLeaseHistory(e, leaseHist.ID, lease.End)
}

func updateLeaseHistory(e *common.Environment, id int, end time.Time) error {
	stmt := `UPDATE "lease_history" SET "end" = ? WHERE "id" = ?`
	_, err := e.DB.Exec(stmt, end.Unix(), id)
	return err
}

func insertLeaseHistory(e *common.Environment, lease *dhcp.Lease) error {
	stmt := `INSERT INTO "lease_history" ("ip", "mac", "network", "start", "end") VALUES (?,?,?,?,?)`

	_, err := e.DB.Exec(
		stmt,
		lease.IP.String(),
		lease.MAC.String(),
		lease.Network,
		lease.Start.Unix(),
		lease.End.Unix(),
	)
	return err
}

func ClearLeaseHistory(e *common.Environment, mac net.HardwareAddr) error {
	stmt := `DELETE FROM "lease_history" WHERE "mac" = ?`
	_, err := e.DB.Exec(stmt, mac.String())
	return err
}
