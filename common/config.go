package common

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/naoina/toml"
)

// LoadConfig loads a toml configuration from the given configFile. If configFile is
// an empty string then it defaults to "config.toml"
func LoadConfig(configFile string) {
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
	if err := toml.Unmarshal(buf, &Config); err != nil {
		panic(err)
	}
}

type config struct {
	Core struct {
		DatabaseFile string
	}
	Webserver struct {
		Address            string
		Port               int
		SessionName        string
		SessionsDir        string
		SessionsAuthKey    string
		SessionsEncryptKey string
	}
	Auth struct {
		AuthMethod []string

		LDAP struct {
			UseAD   bool
			Servers []string
			UseTLS  bool
		}
	}
}

// Config is the application-wide configuration object
var Config config
