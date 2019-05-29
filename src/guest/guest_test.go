package guest

import (
	"net"
	"net/http"
	"testing"

	"github.com/packet-guardian/dhcp-lib"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
)

// TestGuestRegister tests the RegisterDevice function for guests.
// Although it tests the functionality, there's currently no way to
// inspect the device object to make sure it's being created
// and setup correctly.
func TestGuestRegister(t *testing.T) {
	testUserStore := &stores.TestUserStore{}

	testMac, _ := net.ParseMAC("ab:cd:ef:12:34:56")
	testLease := &dhcp.Lease{
		ID:         1,
		IP:         net.ParseIP("192.168.1.2"),
		MAC:        testMac,
		Registered: true,
	}

	testLeaseStore := &stores.TestLeaseStore{
		Leases: []*dhcp.Lease{testLease},
	}

	testDeviceStore := &stores.TestDeviceStore{}

	e := common.NewTestEnvironment()
	e.Config.Guest.DeviceExpirationType = "never"
	r, err := http.NewRequest("POST", "/", nil)
	r.RemoteAddr = "192.168.1.2"
	r = common.SetIPToContext(r)
	if err != nil {
		t.Fatal(err)
	}

	if err := RegisterDevice(e, "John Doe", "johndoe@example.com", r, testUserStore, testDeviceStore, testLeaseStore); err != nil {
		t.Fatal(err)
	}
}
