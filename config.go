package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/naoina/toml"
	"github.com/onesimus-systems/net-guardian/common"
)

func loadConfig(configFile string) *common.Config {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("There was an error loading the configuration file: ", r)
			os.Exit(1)
		}
	}()

	if configFile == "" {
		configFile = "config.toml"
	}

	f, err := os.Open(configFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	var conf common.Config
	if err := toml.Unmarshal(buf, &conf); err != nil {
		panic(err)
	}
	return &conf
}
