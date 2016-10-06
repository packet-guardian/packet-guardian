package models

import (
	"net"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/usi-lfkeitel/packet-guardian/src/common"
	dhcp "github.com/usi-lfkeitel/pg-dhcp"
)

func TestLeaseHistorySave(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	now := time.Now()

	e := common.NewTestEnvironment()
	e.DB = &common.DatabaseAccessor{DB: db}
	e.Config.Leases.HistoryEnabled = true

	lease := dhcp.NewLease(nil)
	lease.IP = net.ParseIP("192.168.1.1")
	lease.MAC = net.HardwareAddr([]byte{0xab, 0xcd, 0xef, 0x12, 0x34, 0x56})
	lease.Network = "main"
	lease.Start = now.Add(time.Duration(-1) * time.Hour)
	lease.End = now.Add(time.Duration(1) * time.Hour)
	lease.Hostname = "something"

	mock.ExpectQuery(`SELECT "id"`).
		WithArgs("ab:cd:ef:12:34:56", "192.168.1.1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(common.LeaseHistoryTableCols))
	mock.ExpectExec(`INSERT INTO "lease_history"`).
		WithArgs("192.168.1.1", "ab:cd:ef:12:34:56", "main", lease.Start.Unix(), lease.End.Unix()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	processLease(e, lease)

	lease.Start = now.Add(time.Duration(30) * time.Minute)
	lease.End = now.Add(time.Duration(90) * time.Minute)
	lease.ID = 1
	mock.ExpectQuery(`SELECT "id"`).
		WithArgs("ab:cd:ef:12:34:56", "192.168.1.1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectExec(`UPDATE "lease_history"`).
		WithArgs(lease.End.Unix(), 1).
		WillReturnResult(sqlmock.NewResult(2, 1))

	processLease(e, lease)

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}
