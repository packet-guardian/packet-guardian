package common

import (
	"io/ioutil"
	"os"

	"github.com/naoina/toml"
)

func init() {
	configFile := "config.toml"

	f, err := os.Open(configFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	if err := toml.Unmarshal(buf, &Config); err != nil {
		panic(err)
	}
}

type config struct {
	Webserver struct {
		Address            string
		Port               int
		SessionsDir        string
		SessionsAuthKey    string
		SessionsEncryptKey string
	}
}

// Config is the application-wide configuration object
var Config config
