// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/**
 * This application runs the Packet Guardian DHCP server as a separate process.
 * By default, the main PG binary will not run a DHCP server and it may be better
 * in some circumstances to not allow the main binary to run with root privilages
 * as they are needed to bind to DHCP port 69.
 */

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/dhcp"
)

var (
	configFile string
	dhcpConfig string
	dev        bool
)

func init() {
	flag.StringVar(&configFile, "pc", "", "Packet Guardian configuration file")
	flag.StringVar(&dhcpConfig, "dc", "", "DHCP configuration file")
	flag.BoolVar(&dev, "d", false, "Run in development mode")
}

func main() {
	flag.Parse()

	var err error
	e := common.NewEnvironment(common.EnvProd)
	if dev {
		e.Env = common.EnvDev
	}

	e.Config, err = common.NewConfig(configFile)
	if err != nil {
		fmt.Printf("Error loading configuration: %s\n", err.Error())
		os.Exit(1)
	}

	e.Log = common.NewLogger(e.Config, "dhcp")
	e.Log.Debugf("Configuration loaded from %s", configFile)

	e.DB, err = common.NewDatabaseAccessor(e.Config)
	if err != nil {
		e.Log.Fatalf("Error loading database: %s", err.Error())
	}
	e.Log.Debugf("Using %s database at %s", e.Config.Database.Type, e.Config.Database.Address)

	dhcpConfig, err := dhcp.ParseFile(dhcpConfig)
	if err != nil {
		e.Log.WithField("ErrMsg", err).Fatal("Error loading DHCP configuration")
	}

	handler := dhcp.NewDHCPServer(dhcpConfig, e)
	if err := handler.LoadLeases(); err != nil {
		e.Log.WithField("ErrMsg", err).Fatal("Couldn't load leases")
	}
	e.Log.Fatal(handler.ListenAndServe())
}
