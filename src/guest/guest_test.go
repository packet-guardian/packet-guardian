package guest

import (
	"net/http"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/usi-lfkeitel/packet-guardian/src/common"
)

// TestGuestRegister tests the RegisterDevice function for guests.
// Although it tests the functionality, there's currently no way to
// inspect the device object to make sure it's being created
// and setup correctly.
func TestGuestRegister(t *testing.T) {
	// Setup Mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Get user model
	mock.ExpectQuery("SELECT .*? FROM \"user\" .*").
		WithArgs("johndoe@example.com").
		WillReturnRows(sqlmock.NewRows(common.UserTableCols))

	// Check device count
	mock.ExpectQuery("SELECT .*? FROM \"device\" .*").
		WithArgs("johndoe@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(0))

	// Get lease
	mock.ExpectQuery("SELECT .*? FROM \"lease\" .*").
		WithArgs("192.168.1.2").
		WillReturnRows(sqlmock.NewRows(common.LeaseTableCols).AddRow(
			1, "192.168.1.2", "ab:cd:ef:12:34:56", "", 0, 0, "", 0, 1,
		))

	// Get device
	mock.ExpectQuery("SELECT .*? FROM \"device\"").
		WithArgs("ab:cd:ef:12:34:56").
		WillReturnRows(sqlmock.NewRows(common.DeviceTableRows))

	// Save device
	mock.ExpectExec("INSERT INTO \"device\"").WillReturnResult(sqlmock.NewResult(1, 1))

	e := common.NewTestEnvironment()
	e.DB = &common.DatabaseAccessor{DB: db}
	e.Config.Guest.DeviceExpirationType = "never"
	r, err := http.NewRequest("POST", "/", nil)
	r.RemoteAddr = "192.168.1.2"
	r = common.SetIPToContext(r)
	if err != nil {
		t.Fatal(err)
	}

	if err := RegisterDevice(e, "John Doe", "johndoe@example.com", r); err != nil {
		t.Fatal(err)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}
