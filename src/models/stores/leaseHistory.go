package stores

import (
	"database/sql"
	"net"
	"time"

	"github.com/lfkeitel/verbose"
	"github.com/packet-guardian/packet-guardian/src/common"
	dhcp "github.com/packet-guardian/pg-dhcp"
)

type LeaseHistory struct {
	ID      int
	IP      net.IP
	MAC     net.HardwareAddr
	Network string
	Start   time.Time
	End     time.Time
}

func (l *LeaseHistory) GetID() int {
	return l.ID
}
func (l *LeaseHistory) GetIP() net.IP {
	return l.IP
}
func (l *LeaseHistory) GetMAC() net.HardwareAddr {
	return l.MAC
}
func (l *LeaseHistory) GetNetworkName() string {
	return l.Network
}
func (l *LeaseHistory) GetStartTime() time.Time {
	return l.Start
}
func (l *LeaseHistory) GetEndTime() time.Time {
	return l.End
}

type leaseHistoryStore struct {
	e *common.Environment
}

func newLeaseHistoryStore(e *common.Environment) *leaseHistoryStore {
	return &leaseHistoryStore{
		e: e,
	}
}

func (l *leaseHistoryStore) getActiveLeaseHistory(mac net.HardwareAddr, ip net.IP) (*LeaseHistory, error) {
	stmt := `SELECT "id" FROM "lease_history" WHERE "mac" = ? AND ip = ? AND "start" <= ? AND "end" >= ? ORDER BY "start" DESC LIMIT 1`

	now := time.Now().Unix()
	row := l.e.DB.QueryRow(stmt, mac.String(), ip.String(), now, now)
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
func (l *leaseHistoryStore) addToLeaseHistory(leaseChan <-chan *dhcp.Lease) {
	for lease := range leaseChan {
		if err := l.processLease(lease); err != nil {
			l.e.Log.WithFields(verbose.Fields{
				"error":   err,
				"package": "models:leaseHistory",
			}).Error("Error saving lease history")
		}
	}
}

func (l *leaseHistoryStore) processLease(lease *dhcp.Lease) error {
	leaseHist, err := l.getActiveLeaseHistory(lease.MAC, lease.IP)
	if err != nil {
		return err
	}
	if leaseHist == nil {
		return l.insertLeaseHistory(lease)
	}
	return l.updateLeaseHistory(leaseHist.ID, lease.End)
}

func (l *leaseHistoryStore) updateLeaseHistory(id int, end time.Time) error {
	stmt := `UPDATE "lease_history" SET "end" = ? WHERE "id" = ?`
	_, err := l.e.DB.Exec(stmt, end.Unix(), id)
	return err
}

func (l *leaseHistoryStore) insertLeaseHistory(lease *dhcp.Lease) error {
	stmt := `INSERT INTO "lease_history" ("ip", "mac", "network", "start", "end") VALUES (?,?,?,?,?)`

	_, err := l.e.DB.Exec(
		stmt,
		lease.IP.String(),
		lease.MAC.String(),
		lease.Network,
		lease.Start.Unix(),
		lease.End.Unix(),
	)
	return err
}
