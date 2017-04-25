// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/**
 * This application runs the Packet Guardian DHCP server as a separate process.
 * By default, the main PG binary will not run a DHCP server and it may be better
 * in some circumstances to not allow the main binary to run with root privileges
 * as they are needed to bind to DHCP port 69.
 */

package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/lfkeitel/verbose"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/db"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
	"github.com/packet-guardian/pg-dhcp"
)

var (
	configFile         string
	dev                bool
	testMainConfigFlag bool
	testDHCPConfigFlag bool
	verFlag            bool

	version   = ""
	buildTime = ""
	builder   = ""
	goversion = ""
)

func init() {
	flag.StringVar(&configFile, "c", "", "Configuration file path")
	flag.BoolVar(&dev, "d", false, "Run in development mode")
	flag.BoolVar(&testMainConfigFlag, "t", false, "Test main configuration file")
	flag.BoolVar(&testDHCPConfigFlag, "td", false, "Test DHCP server configuration file")
	flag.BoolVar(&verFlag, "version", false, "Display version information")
	flag.BoolVar(&verFlag, "v", verFlag, "Display version information")
}

func main() {
	flag.Parse()

	if verFlag {
		displayVersionInfo()
		return
	}

	if testMainConfigFlag {
		testMainConfig()
		return
	}

	if testDHCPConfigFlag {
		testDHCPConfig()
		return
	}

	var err error
	e := common.NewEnvironment(common.EnvProd)
	if dev {
		e.Env = common.EnvDev
	}

	if configFile == "" || !common.FileExists(configFile) {
		configFile = common.FindConfigFile()
	}
	if configFile == "" {
		fmt.Println("No configuration file found")
		os.Exit(1)
	}

	e.Config, err = common.NewConfig(configFile)
	if err != nil {
		fmt.Printf("Error loading configuration: %s\n", err.Error())
		os.Exit(1)
	}

	e.Log = common.NewLogger(e.Config, "dhcp")
	common.SystemLogger = e.Log
	e.Log.Debugf("Configuration loaded from %s", configFile)

	if !common.FileExists(e.Config.DHCP.ConfigFile) {
		e.Log.Fatalf("DHCP configuration file not found: %s", e.Config.DHCP.ConfigFile)
	}

	c := e.SubscribeShutdown()
	go func(e *common.Environment) {
		<-c
		e.Log.Notice("Shutting down...")
		time.Sleep(2)
	}(e)

	e.DB, err = db.NewDatabaseAccessor(e)
	if err != nil {
		e.Log.WithField("error", err).Fatal("Error loading database")
	}
	e.Log.WithFields(verbose.Fields{
		"type":    e.Config.Database.Type,
		"address": e.Config.Database.Address,
	}).Debug("Loaded database")

	dhcpConfig, err := dhcp.ParseFile(e.Config.DHCP.ConfigFile)
	if err != nil {
		e.Log.WithField("error", err).Fatal("Error loading DHCP configuration")
	}

	dhcpPkgConfig := &dhcp.ServerConfig{
		LeaseStore:  stores.NewLeaseStore(e),
		DeviceStore: &dhcpDeviceStore{e: e},
		Env:         dhcp.EnvDev,
		Log:         common.NewLogger(e.Config, "dhcp").Logger,
	}

	handler := dhcp.NewDHCPServer(dhcpConfig, dhcpPkgConfig)
	if err := handler.LoadLeases(); err != nil {
		e.Log.WithField("error", err).Fatal("Couldn't load leases")
	}
	e.Log.Fatal(handler.ListenAndServe())
}

type dhcpDeviceStore struct {
	e *common.Environment
}

func (d *dhcpDeviceStore) GetDeviceByMAC(mac net.HardwareAddr) (dhcp.Device, error) {
	return stores.GetDeviceStore(d.e).GetDeviceByMAC(mac)
}

func displayVersionInfo() {
	fmt.Printf(`Packet Guardian - (C) 2016 The Packet Guardian Authors

Component:   DHCP Server
Version:     %s
Built:       %s
Compiled by: %s
Go version:  %s
`, version, buildTime, builder, goversion)
}

func testMainConfig() {
	_, err := common.NewConfig(configFile)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Configuration looks good")
}

func testDHCPConfig() {
	_, err := dhcp.ParseFile(configFile)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	fmt.Println("Configuration looks good")
}
