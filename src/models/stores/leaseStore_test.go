// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stores

import (
	"net"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/usi-lfkeitel/packet-guardian/src/common"
	dhcp "github.com/usi-lfkeitel/pg-dhcp"
)

func TestLeaseSaveNoHistory(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	now := time.Now()

	e := common.NewTestEnvironment()
	e.DB = &common.DatabaseAccessor{DB: db}

	lease := dhcp.NewLease(NewLeaseStore(e))
	lease.IP = net.ParseIP("192.168.1.1")
	lease.MAC = net.HardwareAddr([]byte{0xab, 0xcd, 0xef, 0x12, 0x34, 0x56})
	lease.Network = "main"
	lease.Start = now.Add(time.Duration(-1) * time.Hour)
	lease.End = now.Add(time.Duration(1) * time.Hour)
	lease.Hostname = "something"

	mock.ExpectExec(`INSERT INTO "lease"`).
		WithArgs("192.168.1.1", "ab:cd:ef:12:34:56", "main", lease.Start.Unix(), lease.End.Unix(), "something", false, false).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := lease.Save(); err != nil {
		t.Fatalf("Failed to save lease: %s", err)
	}

	lease.Start = now.Add(time.Duration(30) * time.Minute)
	lease.End = now.Add(time.Duration(90) * time.Minute)
	mock.ExpectExec(`UPDATE "lease"`).
		WithArgs("ab:cd:ef:12:34:56", lease.Start.Unix(), lease.End.Unix(), "something", false, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := lease.Save(); err != nil {
		t.Fatalf("Failed to save lease: %s", err)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
