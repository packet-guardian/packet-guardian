// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/lfkeitel/verbose/v4"
	"github.com/packet-guardian/packet-guardian/src/bindata"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/db"
	"github.com/packet-guardian/packet-guardian/src/models"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
	"github.com/packet-guardian/packet-guardian/src/server"
	"github.com/packet-guardian/packet-guardian/src/tasks"
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
	flag.BoolVar(&testConfig, "t", false, "Test main configuration file")
	flag.BoolVar(&verFlag, "version", false, "Display version information")
	flag.BoolVar(&verFlag, "v", verFlag, "Display version information")
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	common.SystemVersion = version

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

	e := setupEnvironment()
	startShutdownWatcher(e)

	if err := bindata.SetCustomDir(e.Config.Webserver.CustomDataDir); err != nil {
		e.Log.WithField("error", err).Fatal("Error loading frontend templates")
	}

	if err := common.RunSystemInits(e); err != nil {
		e.Log.WithField("error", err).Fatal("System initialization failed")
	}

	appStores := stores.StoreCollection{
		Blacklist: stores.GetBlacklistStore(e),
		Devices:   stores.GetDeviceStore(e),
		Leases:    stores.GetLeaseStore(e),
		Users:     stores.GetUserStore(e),
	}

	go tasks.StartTaskScheduler(e, appStores)

	// Start web server
	server.NewServer(e, server.LoadRoutes(e, appStores)).Run()
}

func setupEnvironment() *common.Environment {
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

	e.DB, err = db.NewDatabaseAccessor(e)
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

	e.Views.InjectData("config", e.Config)
	e.Views.InjectData("systemVersion", version)
	e.Views.InjectData("buildTime", buildTime)
	e.Views.InjectDataFunc("sessionUser", func(r *http.Request) interface{} {
		return models.GetUserFromContext(r)
	})
	e.Views.InjectDataFunc("canViewUsers", func(r *http.Request) interface{} {
		return models.GetUserFromContext(r).Can(models.ViewUsers)
	})

	return e
}

func startShutdownWatcher(e *common.Environment) {
	c := e.SubscribeShutdown()
	go func(e *common.Environment) {
		<-c
		if err := e.DB.Close(); err != nil {
			e.Log.Warningf("Error closing database: %s", err)
		}
		e.Log.Notice("Shutting down...")
		time.Sleep(2)
	}(e)
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
