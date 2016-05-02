package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/server"
)

var (
	configFile string
	dev        bool
)

func init() {
	flag.StringVar(&configFile, "config", "", "Configuration file path")
	flag.BoolVar(&dev, "dev", false, "Run in development mode")
}

func main() {
	// Parse CLI flags
	flag.Parse()

	var err error
	e := common.NewEnvironment(dev)

	e.Config, err = common.NewConfig(configFile)
	if err != nil {
		fmt.Printf("Error parsing configuration: %s\n", err.Error())
		os.Exit(1)
	}

	e.Log = common.NewLogger(e.Config)
	e.Log.Debugf("Configuration loaded from %s", configFile)

	if dev {
		e.Log.Debug("Packet Guardian running in DEVELOPMENT mode")
	}

	e.Sessions, err = common.NewSessionStore(e.Config)
	if err != nil {
		e.Log.Fatalf("Error loading session store: %s", err.Error())
	}

	e.DB, err = common.NewDatabaseAccessor(e.Config)
	if err != nil {
		e.Log.Fatalf("Error loading database: %s", err.Error())
	}
	e.Log.Debugf("Using %s database at %s", e.Config.Database.Type, e.Config.Database.Address)

	e.Views, err = common.NewViews(e, "templates")
	if err != nil {
		e.Log.Fatalf("Error loading frontend templates: %s", err.Error())
	}

	// Let's begin!
	server.NewServer(e, server.LoadRoutes(e)).Run()
}
