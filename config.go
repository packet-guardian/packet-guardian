package main

import (
	"io/ioutil"
	"os"
	"errors"

	"github.com/naoina/toml"
	"github.com/onesimus-systems/packet-guardian/common"
)

func loadConfig(configFile string) (conf *common.Config, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()

	if configFile == "" {
		configFile = "config.toml"
	}

	f, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var con common.Config
	if err := toml.Unmarshal(buf, &con); err != nil {
		return nil, err
	}
	return &con, nil
}
