package dhcp

import (
	"bufio"
	"bytes"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lfkeitel/verbose"
	d4 "github.com/onesimus-systems/dhcp4"
	"github.com/onesimus-systems/packet-guardian/src/common"
)

var testConfig = `
global
    option domain-name example.com

    server-identifier 10.0.0.1

    registered
        default-lease-time 86400
        max-lease-time 86400
        option domain-name-server 10.1.0.1, 10.1.0.2
    end

    unregistered
        default-lease-time 360
        max-lease-time 360
        option domain-name-server 10.0.0.1
    end
end

network Network1
    unregistered
        subnet 10.0.1.0/24
            range 10.0.1.10 10.0.1.200
            option router 10.0.1.1
        end
    end
    registered
        subnet 10.0.2.0/24
            range 10.0.2.10 10.0.2.200
            option router 10.0.2.1
        end
    end
end

network Network2
    unregistered
        subnet 10.0.3.0/24
            option router 10.0.3.1
            range 10.0.3.20 10.0.3.200
        end
        subnet 10.0.4.0/24
            range 10.0.4.10 10.0.4.200
            option router 10.0.4.1
        end
    end
    registered
        subnet 10.0.5.0/24
            range 10.0.5.10 10.0.5.200
            option router 10.0.5.1
        end
    end
end
`

var deviceTableRows = []string{
	"id",
	"mac",
	"username",
	"registered_from",
	"platform",
	"expires",
	"date_registered",
	"user_agent",
	"blacklisted",
	"description",
}

func TestDiscover(t *testing.T) {
	// Set up mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows(deviceTableRows).AddRow(
		1, "12:34:56:12:34:56", "", "", "", time.Now().Add(time.Duration(15)*time.Second).Unix(), 0, "", false, "",
	)

	rows2 := sqlmock.NewRows(deviceTableRows).AddRow(
		1, "12:34:56:12:34:56", "", "", "", time.Now().Add(time.Duration(15)*time.Second).Unix(), 0, "", false, "",
	)

	rows3 := sqlmock.NewRows(deviceTableRows)

	rows4 := sqlmock.NewRows(deviceTableRows)

	mock.ExpectQuery("SELECT .*? FROM \"device\"").
		WithArgs("12:34:56:12:34:56").
		WillReturnRows(rows)

	mock.ExpectQuery("SELECT .*? FROM \"device\"").
		WithArgs("12:34:56:12:34:56").
		WillReturnRows(rows2)

	mock.ExpectQuery("SELECT .*? FROM \"device\"").
		WithArgs("12:34:56:12:34:56").
		WillReturnRows(rows3)

	mock.ExpectQuery("SELECT .*? FROM \"device\"").
		WithArgs("12:34:56:12:34:56").
		WillReturnRows(rows4)

	mock.ExpectExec("INSERT INTO \"lease\"").
		WillReturnResult(sqlmock.NewResult(1, 1))

		// Setup environment
	e := common.NewTestEnvironment()
	e.DB = &common.DatabaseAccessor{DB: db}
	stdout := verbose.NewStdoutHandler()
	stdout.SetMinLevel(verbose.LogLevelDebug)
	e.Log.AddHandler("stdout", stdout)

	// Setup Confuration
	reader := strings.NewReader(testConfig)
	c, err := newParser(bufio.NewScanner(reader)).parse()
	if err != nil {
		t.Fatalf("Test config failed parsing: %v", err)
		return
	}

	server := NewDHCPServer(c, e)

	// Create test request packet
	mac, _ := net.ParseMAC("12:34:56:12:34:56")
	opts := []d4.Option{
		d4.Option{
			Code:  d4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23},
		},
	}
	p := d4.RequestPacket(d4.Discover, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.1.5"))

	// Process a DISCOVER request
	start := time.Now()
	dp := server.ServeDHCP(p, d4.Discover, p.ParseOptions())
	t.Logf("Discover took: %v", time.Since(start))

	if dp == nil {
		t.Fatal("Processed packet is nil")
	}

	if !bytes.Equal(dp.YIAddr().To4(), []byte{0xa, 0x0, 0x2, 0xa}) {
		t.Errorf("Incorrect offered IP. Expected %v, got %v", []byte{0xa, 0x0, 0x2, 0xa}, dp.YIAddr())
	}

	options := dp.ParseOptions()
	if v, ok := options[d4.OptionSubnetMask]; !ok { // 0x1
		t.Error("Subnet mask option not received")
	} else if !bytes.Equal(v, []byte{0xff, 0xff, 0xff, 0x0}) {
		t.Errorf("Incorrect subnet mask option. Expected %v, got %v", []byte{0xff, 0xff, 0xff, 0x0}, v)
	}

	if v, ok := options[d4.OptionRouter]; !ok { // 0x3
		t.Error("Router option not received")
	} else if !bytes.Equal(v, []byte{0xa, 0x0, 0x2, 0x1}) {
		t.Errorf("Incorrect router option. Expected %v, got %v", []byte{0xa, 0x0, 0x2, 0x1}, v)
	}

	if v, ok := options[d4.OptionDomainNameServer]; !ok { // 0x6
		t.Error("Domain name server option not received")
	} else if !bytes.Equal(v, []byte{0xa, 0x1, 0x0, 0x1, 0xa, 0x1, 0x0, 0x2}) {
		t.Errorf("Incorrect domain name server option. Expected %v, got %v", []byte{0xa, 0x1, 0x0, 0x1, 0xa, 0x1, 0x0, 0x2}, v)
	}

	if v, ok := options[d4.OptionDomainName]; !ok { // 0xf (15)
		t.Error("Domain name option not received")
	} else if string(v) != "example.com" {
		t.Errorf("Incorrect domain name option. Expected %s, got %s", "example.com", string(v))
	}

	if v, ok := options[d4.OptionIPAddressLeaseTime]; !ok { // 0x23 (51)
		t.Error("Lease time option not received")
	} else if !bytes.Equal(v, []byte{0x0, 0x1, 0x51, 0x80}) {
		t.Errorf("Incorrect lease time option. Expected %v, got %v", []byte{0x0, 0x1, 0x51, 0x80}, v)
	}

	opts = []d4.Option{
		d4.Option{
			Code:  d4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23},
		},
		d4.Option{
			Code:  d4.OptionServerIdentifier,
			Value: []byte(options[d4.OptionServerIdentifier]),
		},
		d4.Option{
			Code:  d4.OptionRequestedIPAddress,
			Value: []byte(dp.YIAddr().To4()),
		},
	}
	p = d4.RequestPacket(d4.Request, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.1.5"))

	// Process a REQUEST request
	start = time.Now()
	rp := server.ServeDHCP(p, d4.Request, p.ParseOptions())
	t.Logf("Request took: %v", time.Since(start))

	if rp == nil {
		t.Fatal("Processed packet is nil")
	}

	if !bytes.Equal(rp.YIAddr().To4(), dp.YIAddr().To4()) {
		t.Errorf("Incorrect offered IP. Expected %v, got %v", dp.YIAddr(), rp.YIAddr())
	}

	options = rp.ParseOptions()
	if v, ok := options[d4.OptionDHCPMessageType]; !ok { // 0x53
		t.Error("DHCP message type option not received")
	} else if !bytes.Equal(v, []byte{0x5}) {
		t.Errorf("Incorrect DHCP message option option. Expected %v, got %v", []byte{0x5}, v)
	}

	if v, ok := options[d4.OptionSubnetMask]; !ok { // 0x1
		t.Error("Subnet mask option not received")
	} else if !bytes.Equal(v, []byte{0xff, 0xff, 0xff, 0x0}) {
		t.Errorf("Incorrect subnet mask option. Expected %v, got %v", []byte{0xff, 0xff, 0xff, 0x0}, v)
	}

	if v, ok := options[d4.OptionRouter]; !ok { // 0x3
		t.Error("Router option not received")
	} else if !bytes.Equal(v, []byte{0xa, 0x0, 0x2, 0x1}) {
		t.Errorf("Incorrect router option. Expected %v, got %v", []byte{0xa, 0x0, 0x2, 0x1}, v)
	}

	if v, ok := options[d4.OptionDomainNameServer]; !ok { // 0x6
		t.Error("Domain name server option not received")
	} else if !bytes.Equal(v, []byte{0xa, 0x1, 0x0, 0x1, 0xa, 0x1, 0x0, 0x2}) {
		t.Errorf("Incorrect domain name server option. Expected %v, got %v", []byte{0xa, 0x1, 0x0, 0x1, 0xa, 0x1, 0x0, 0x2}, v)
	}

	if v, ok := options[d4.OptionDomainName]; !ok { // 0xf (15)
		t.Error("Domain name option not received")
	} else if string(v) != "example.com" {
		t.Errorf("Incorrect domain name option. Expected %s, got %s", "example.com", string(v))
	}

	if v, ok := options[d4.OptionIPAddressLeaseTime]; !ok { // 0x23 (51)
		t.Error("Lease time option not received")
	} else if !bytes.Equal(v, []byte{0x0, 0x1, 0x51, 0x80}) {
		t.Errorf("Incorrect lease time option. Expected %v, got %v", []byte{0x0, 0x1, 0x51, 0x80}, v)
	}

	// ROUND 2 - Fight!
	opts = []d4.Option{
		d4.Option{
			Code:  d4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23},
		},
	}
	p = d4.RequestPacket(d4.Discover, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.1.5"))

	// Process a DISCOVER request
	start = time.Now()
	dp = server.ServeDHCP(p, d4.Discover, p.ParseOptions())
	t.Logf("Discover took: %v", time.Since(start))

	if dp == nil {
		t.Fatal("Processed packet is nil")
	}

	if !bytes.Equal(dp.YIAddr().To4(), []byte{0xa, 0x0, 0x1, 0xa}) {
		t.Errorf("Incorrect offered IP. Expected %v, got %v", []byte{0xa, 0x0, 0x1, 0xa}, dp.YIAddr())
	}

	options = dp.ParseOptions()
	if v, ok := options[d4.OptionSubnetMask]; !ok { // 0x1
		t.Error("Subnet mask option not received")
	} else if !bytes.Equal(v, []byte{0xff, 0xff, 0xff, 0x0}) {
		t.Errorf("Incorrect subnet mask option. Expected %v, got %v", []byte{0xff, 0xff, 0xff, 0x0}, v)
	}

	if v, ok := options[d4.OptionRouter]; !ok { // 0x3
		t.Error("Router option not received")
	} else if !bytes.Equal(v, []byte{0xa, 0x0, 0x1, 0x1}) {
		t.Errorf("Incorrect router option. Expected %v, got %v", []byte{0xa, 0x0, 0x1, 0x1}, v)
	}

	if v, ok := options[d4.OptionDomainNameServer]; !ok { // 0x6
		t.Error("Domain name server option not received")
	} else if !bytes.Equal(v, []byte{0xa, 0x0, 0x0, 0x1}) {
		t.Errorf("Incorrect domain name server option. Expected %v, got %v", []byte{0xa, 0x0, 0x0, 0x1}, v)
	}

	if v, ok := options[d4.OptionDomainName]; !ok { // 0xf (15)
		t.Error("Domain name option not received")
	} else if string(v) != "example.com" {
		t.Errorf("Incorrect domain name option. Expected %s, got %s", "example.com", string(v))
	}

	if v, ok := options[d4.OptionIPAddressLeaseTime]; !ok { // 0x23 (51)
		t.Error("Lease time option not received")
	} else if !bytes.Equal(v, []byte{0x0, 0x0, 0x1, 0x68}) {
		t.Errorf("Incorrect lease time option. Expected %v, got %v", []byte{0x0, 0x0, 0x1, 0x68}, v)
	}

	opts = []d4.Option{
		d4.Option{
			Code:  d4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23},
		},
		d4.Option{
			Code:  d4.OptionServerIdentifier,
			Value: []byte(options[d4.OptionServerIdentifier]),
		},
		d4.Option{
			Code:  d4.OptionRequestedIPAddress,
			Value: []byte(dp.YIAddr().To4()),
		},
	}
	p = d4.RequestPacket(d4.Request, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.1.5"))

	// Process a REQUEST request
	start = time.Now()
	rp = server.ServeDHCP(p, d4.Request, p.ParseOptions())
	t.Logf("Request took: %v", time.Since(start))

	if rp == nil {
		t.Fatal("Processed packet is nil")
	}

	if !bytes.Equal(rp.YIAddr().To4(), dp.YIAddr().To4()) {
		t.Errorf("Incorrect offered IP. Expected %v, got %v", dp.YIAddr(), rp.YIAddr())
	}

	options = rp.ParseOptions()
	if v, ok := options[d4.OptionDHCPMessageType]; !ok { // 0x53
		t.Error("DHCP message type option not received")
	} else if !bytes.Equal(v, []byte{0x5}) {
		t.Errorf("Incorrect DHCP message option option. Expected %v, got %v", []byte{0x5}, v)
	}

	if v, ok := options[d4.OptionSubnetMask]; !ok { // 0x1
		t.Error("Subnet mask option not received")
	} else if !bytes.Equal(v, []byte{0xff, 0xff, 0xff, 0x0}) {
		t.Errorf("Incorrect subnet mask option. Expected %v, got %v", []byte{0xff, 0xff, 0xff, 0x0}, v)
	}

	if v, ok := options[d4.OptionRouter]; !ok { // 0x3
		t.Error("Router option not received")
	} else if !bytes.Equal(v, []byte{0xa, 0x0, 0x1, 0x1}) {
		t.Errorf("Incorrect router option. Expected %v, got %v", []byte{0xa, 0x0, 0x1, 0x1}, v)
	}

	if v, ok := options[d4.OptionDomainNameServer]; !ok { // 0x6
		t.Error("Domain name server option not received")
	} else if !bytes.Equal(v, []byte{0xa, 0x0, 0x0, 0x1}) {
		t.Errorf("Incorrect domain name server option. Expected %v, got %v", []byte{0xa, 0x0, 0x0, 0x1}, v)
	}

	if v, ok := options[d4.OptionDomainName]; !ok { // 0xf (15)
		t.Error("Domain name option not received")
	} else if string(v) != "example.com" {
		t.Errorf("Incorrect domain name option. Expected %s, got %s", "example.com", string(v))
	}

	if v, ok := options[d4.OptionIPAddressLeaseTime]; !ok { // 0x23 (51)
		t.Error("Lease time option not received")
	} else if !bytes.Equal(v, []byte{0x0, 0x0, 0x1, 0x68}) {
		t.Errorf("Incorrect lease time option. Expected %v, got %v", []byte{0x0, 0x0, 0x1, 0x68}, v)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expections: %s", err)
	}
}
