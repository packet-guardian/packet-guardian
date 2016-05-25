// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/dhcp"
	"github.com/onesimus-systems/packet-guardian/src/server"
	"github.com/onesimus-systems/packet-guardian/src/tasks"
)

const (
	version = "0.6.2"
)

var (
	configFile   string
	dev          bool
	verFlag      bool
	testConfig   bool
	testDHCPConf bool
)

func init() {
	flag.StringVar(&configFile, "c", "", "Configuration file path")
	flag.BoolVar(&dev, "d", false, "Run in development mode")
	flag.BoolVar(&verFlag, "version", false, "Display version information")
	flag.BoolVar(&verFlag, "v", verFlag, "Display version information")
	flag.BoolVar(&testConfig, "test", false, "Test main configuration")
	flag.BoolVar(&testDHCPConf, "testd", false, "Test DHCP configuration only")
}

func main() {
	// Parse CLI flags
	flag.Parse()

	if verFlag {
		displayVersionInfo()
		return
	}

	configFile = findConfigFile()
	if configFile == "" {
		fmt.Println("No configuration file found")
		os.Exit(1)
	}

	if testConfig {
		testMainConfig()
		return
	}

	if testDHCPConf {
		testDHCPConfig()
		return
	}

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

	e.Log = common.NewLogger(e.Config, "app")
	e.Log.Debugf("Configuration loaded from %s", configFile)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func(e *common.Environment) {
		<-c
		e.Log.Notice("Shutting down...")
		time.Sleep(2)
	}(e)

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

	// Start DHCP server
	if e.Config.DHCP.Enabled {
		dhcpConfig, err := dhcp.ParseFile(e.Config.DHCP.ConfigFile)
		if err != nil {
			e.Log.WithField("ErrMsg", err).Fatal("Failed loading DHCP config")
		}

		handler := dhcp.NewDHCPServer(dhcpConfig, e)
		if err := handler.LoadLeases(); err != nil {
			e.Log.WithField("ErrMsg", err).Fatal("Couldn't load leases")
		}
		go func() {
			e.Log.Fatal(handler.ListenAndServe())
		}()
	}

	go tasks.StartTaskScheduler(e)

	// Start web server
	server.NewServer(e, server.LoadRoutes(e)).Run()
}

func displayVersionInfo() {
	fmt.Printf(`Packet Guardian - (C) 2016 Lee Keitel
Onesimus Systems

Version: %s
`, version)
}

func findConfigFile() string {
	// Find a configuration file if one wasn't given
	if configFile != "" && common.FileExists(configFile) {
		return configFile
	}

	if os.Getenv("PG_CONFIG") != "" && common.FileExists(os.Getenv("PG_CONFIG")) {
		return os.Getenv("PG_CONFIG")
	} else if common.FileExists("./config.toml") {
		return "./config.toml"
	} else if common.FileExists(os.ExpandEnv("$HOME/.pg/config.toml")) {
		return os.ExpandEnv("$HOME/.pg/config.toml")
	} else if common.FileExists("/etc/packet-guardian/config.toml") {
		return "/etc/packet-guardian/config.toml"
	}
	return ""
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
