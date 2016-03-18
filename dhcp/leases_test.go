package dhcp

import (
	"net"
	"strings"
	"testing"
	"time"
)

type testCase struct {
	pass     bool
	ip       net.IP
	text     string
	expected *lease
}

// Various test cases for a single lease block
var leaseTests = []testCase{
	// Passing block
	testCase{pass: true, ip: net.ParseIP("192.168.4.2"), text: `
        lease 192.168.4.2 {
          starts 2 2016/01/12 19:45:47;
          ends 2 2016/01/12 19:51:47;
          binding state free;
          hardware ethernet 4c:aa:16:87:b5:88;
        }`,
		expected: &lease{
			mac:   "4c:aa:16:87:b5:88",
			start: time.Date(2016, 1, 12, 19, 45, 47, 0, time.UTC),
			end:   time.Date(2016, 1, 12, 19, 51, 47, 0, time.UTC),
		}},
	// Missing starts date
	testCase{pass: false, ip: net.ParseIP("192.168.4.2"), text: `
        lease 192.168.4.2 {
          ends 2 2016/01/12 19:51:47;
          hardware ethernet 4c:aa:16:87:b5:88;
        }`},
	// Missing ends date
	testCase{pass: false, ip: net.ParseIP("192.168.4.2"), text: `
        lease 192.168.4.2 {
          starts 2 2016/01/12 19:45:47;
          hardware ethernet 4c:aa:16:87:b5:88;
        }`},
	// Missing MAC address
	testCase{pass: false, ip: net.ParseIP("192.168.4.2"), text: `
        lease 192.168.4.2 {
          starts 2 2016/01/12 19:45:47;
          ends 2 2016/01/12 19:51:47;
        }`},
	// Missing everything
	testCase{pass: false, ip: net.ParseIP("192.168.4.2"), text: `
        lease 192.168.4.2 {
        }`},
	// Non-matching block
	testCase{pass: false, ip: net.ParseIP("192.168.4.2"), text: `
        lease 192.168.5.9 {
          starts 2 2016/01/12 19:45:47;
          ends 2 2016/01/12 19:51:47;
          binding state free;
          hardware ethernet 4c:aa:16:87:b5:88;
        }`},
}

func TestLeaseParse(t *testing.T) {
	for _, test := range leaseTests {
		r := strings.NewReader(test.text)
		l, err := getLeaseReader(test.ip, r)
		if test.pass {
			if !sameLease(l, test.expected) {
				t.Errorf("Test Failed. Expected %#v, got %#v", test.expected, l)
			}
			if err != nil {
				t.Errorf("Error encountered: %s", err.Error())
			}
		} else {
			if l != nil {
				t.Errorf("Test failed. Expected nil, got %#v", l)
			}
			if err == nil {
				t.Error("Test failed. Expected error, got nil")
			}
		}
	}
}

// Test that the returned lease is actually the current lease.
// DHCPd appends new leases to the leases file and many times will
// have two lease block for the same IP address. The last block is the current.
func TestMultipleLeases(t *testing.T) {
	text := `
        lease 192.168.4.2 {
          starts 2 2016/01/12 19:45:47;
          ends 2 2016/01/12 19:51:47;
          binding state free;
          hardware ethernet 5b:5f:cd:35:a6:8b;
        }
        lease 192.168.4.2 {
          starts 2 2016/03/18 13:28:47;
          ends 2 2016/03/20 13:28:47;
          binding state free;
          hardware ethernet 4c:aa:16:87:b5:88;
        }`
	expected := &lease{
		start: time.Date(2016, 3, 18, 13, 28, 47, 0, time.UTC),
		end:   time.Date(2016, 3, 20, 13, 28, 47, 0, time.UTC),
		mac:   "4c:aa:16:87:b5:88",
	}

	l, err := getLeaseReader(net.ParseIP("192.168.4.2"), strings.NewReader(text))
	if !sameLease(l, expected) {
		t.Errorf("Lease Failed. Expected %#v, got %#v", expected, l)
	}
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
}

func sameLease(l1, l2 *lease) bool {
	if l1 == nil || l2 == nil {
		return false
	}
	if !l1.start.Equal(l2.start) {
		return false
	}
	if !l1.end.Equal(l2.end) {
		return false
	}
	if l1.mac != l2.mac {
		return false
	}
	return true
}
