package dhcp

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestValidMacCheck(t *testing.T) {
	testCases := map[string]bool{
		"12:34:56:78:90:12": true,
		"ab:f5:7d:3c:56:9a": true,
		"AB:F5:7d:3C:56:9a": true,
		"AB:G5:7I:3C:56:9a": false,
		"":                  false,
		"AB:F5:7d:3C:56:9a:52": false,
		"AB:F5:7d:3C:56":       false,
	}

	for mac, expected := range testCases {
		tested := isValidMac(mac)
		if tested != expected {
			t.Errorf("MAC check failed for %s: Expected %t, got %t", mac, expected, tested)
		}
	}
}

func TestBlacklist(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was encountered setting up mock database", err)
	}
	defer db.Close()

	// Test true blacklist
	mock.ExpectPrepare("SELECT \"id\" FROM \"blacklist\" WHERE 1=0 OR \"value\"=\\? OR \"value\"=\\?").
		ExpectQuery().
		WithArgs("testuser", "12:34:56:12:34:56").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	b, err := IsBlacklisted(db, "testuser", "12:34:56:12:34:56")
	if err != nil {
		t.Fatalf("Error checking blacklist: %s", err)
	}
	if !b {
		t.Error("Blacklist failed. Expected true, got false")
	}

	// Test false blacklist
	mock.ExpectPrepare("SELECT \"id\" FROM \"blacklist\" WHERE 1=0 OR \"value\"=\\?").
		ExpectQuery().
		WithArgs("testuser1")

	b, err = IsBlacklisted(db, "testuser1")
	if err != nil {
		t.Fatalf("Error checking blacklist: %s", err)
	}
	if b {
		t.Error("Blacklist failed. Expected false, got true")
	}

	// Check empty blacklist
	b, err = IsBlacklisted(db)
	if err != nil {
		t.Fatalf("Error checking empty blacklist: %s", err)
	}
	if b {
		t.Error("Blacklist failed. Expected false, got true")
	}

	// Check all SQL statements were executed
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Mock expectations not met: %s", err)
	}
}
