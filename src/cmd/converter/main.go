package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/lfkeitel/verbose"
	"github.com/onesimus-systems/packet-guardian/src/common"
)

const (
	version = "0.1.0"
)

var (
	configFile string
)

func init() {
	flag.StringVar(&configFile, "c", "", "Configuration file path")
}

func main() {
	// Parse CLI flags
	flag.Parse()

	if configFile == "" {
		fmt.Println("No configuration file found")
		os.Exit(1)
	}

	var err error
	e := common.NewEnvironment(common.EnvProd)
	fmt.Println("/*")
	e.Config, err = common.NewConfig(configFile)
	if err != nil {
		fmt.Printf("Error loading configuration: %s\n", err.Error())
		os.Exit(1)
	}

	logger := verbose.New("converter")
	logger.AddHandler("stdout", verbose.NewStdoutHandler())
	e.Log = &common.Logger{Logger: logger}
	e.Log.Debugf("Configuration loaded from %s", configFile)

	e.DB, err = common.NewDatabaseAccessor(e.Config)
	if err != nil {
		e.Log.Fatalf("Error loading database: %s", err.Error())
	}
	e.Log.Debugf("Using %s database at %s", e.Config.Database.Type, e.Config.Database.Address)

	fmt.Println("\n*/")

	for _, file := range flag.Args() {
		if err := parseFile(file, e); err != nil {
			e.Log.Error(err)
		}
	}

	writeOutput()
}
