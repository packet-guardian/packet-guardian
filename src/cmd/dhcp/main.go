package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dragonrider23/verbose"
	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/dhcp"
)

var (
	configFile string
	dhcpConfig string
)

func init() {
	flag.StringVar(&configFile, "config", "", "Configuration file path")
	flag.StringVar(&dhcpConfig, "dhcp", "", "DHCP configuration file path")
}

func main() {
	flag.Parse()

	var err error
	config, err := dhcp.ParseFile(dhcpConfig)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	//config.Print()

	e := common.NewEnvironment(true)

	e.Log = &common.Logger{Logger: verbose.New("dhcp")}
	e.Log.AddHandler("stdout", verbose.NewStdoutHandler())

	e.Config, err = common.NewConfig(configFile)
	if err != nil {
		e.Log.Fatalf("Error loading configuration: %s", err.Error())
	}
	e.Log.Infof("Configuration loaded from %s", configFile)

	e.DB, err = common.NewDatabaseAccessor(e.Config)
	if err != nil {
		e.Log.Fatalf("Error loading database: %s", err.Error())
	}
	e.Log.Infof("Using %s database at %s", e.Config.Database.Type, e.Config.Database.Address)

	handler := dhcp.NewDHCPServer(config, e)
	//handler.Readonly()
	handler.LoadLeases()
	e.Log.Critical(handler.ListenAndServe())
}
