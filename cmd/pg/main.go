// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/lfkeitel/verbose"
	"github.com/usi-lfkeitel/packet-guardian/src/common"
	"github.com/usi-lfkeitel/packet-guardian/src/server"
	"github.com/usi-lfkeitel/packet-guardian/src/tasks"
)

var (
	configFile string
	dev        bool
	verFlag    bool
	testConfig bool

	version   = ""
	buildTime = ""
	builder   = ""
	goversion = ""
)

func init() {
	flag.StringVar(&configFile, "c", "", "Configuration file path")
	flag.BoolVar(&dev, "d", false, "Run in development mode")
	flag.BoolVar(&verFlag, "version", false, "Display version information")
	flag.BoolVar(&verFlag, "v", verFlag, "Display version information")
	flag.BoolVar(&testConfig, "t", false, "Test main configuration")
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Parse CLI flags
	flag.Parse()

	if verFlag {
		displayVersionInfo()
		return
	}

	if configFile == "" || !common.FileExists(configFile) {
		configFile = common.FindConfigFile()
	}
	if configFile == "" {
		fmt.Println("No configuration file found")
		os.Exit(1)
	}

	if testConfig {
		testMainConfig()
		return
	}

	var err error
	e := common.NewEnvironment(common.EnvProd)
	if dev {
		e.Env = common.EnvDev
	}

	e.Config, err = common.NewConfig(configFile)
	if err != nil {
		fmt.Printf("Error loading configuration: %s\n", err)
		os.Exit(1)
	}

	e.Log = common.NewLogger(e.Config, "app")
	common.SystemLogger = e.Log
	e.Log.Debugf("Configuration loaded from %s", configFile)

	c := e.SubscribeShutdown()
	go func(e *common.Environment) {
		<-c
		e.Log.Notice("Shutting down...")
		time.Sleep(2)
	}(e)

	e.DB, err = common.NewDatabaseAccessor(e)
	if err != nil {
		e.Log.WithField("error", err).Fatal("Error loading database")
	}
	e.Log.WithFields(verbose.Fields{
		"type":    e.Config.Database.Type,
		"address": e.Config.Database.Address,
	}).Debug("Loaded database")

	e.Sessions, err = common.NewSessionStore(e)
	if err != nil {
		e.Log.WithField("error", err).Fatal("Error loading session store")
	}

	e.Views, err = common.NewViews(e, "templates")
	if err != nil {
		e.Log.WithField("error", err).Fatal("Error loading frontend templates")
	}

	go tasks.StartTaskScheduler(e)

	// Start web server
	server.NewServer(e, server.LoadRoutes(e)).Run()
}

func displayVersionInfo() {
	fmt.Printf(`Packet Guardian - (C) 2016 The Packet Guardian Authors

Component:   Web Server
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
