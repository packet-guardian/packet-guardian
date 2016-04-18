package common

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/naoina/toml"
)

// Config defines the configuration struct for the application
type Config struct {
	sourceFile string
	Core       struct {
		SiteTitle       string
		SiteCompanyName string
	}
	Database struct {
		Type     string
		Address  string
		Username string
		Password string
	}
	Registration struct {
		RegistrationPolicyFile   string
		AllowManualRegistrations bool
		DefaultDeviceLimit       int
		ManualRegPlatforms       []string
	}
	Webserver struct {
		EnableLogging       bool
		Address             string
		HttpPort            int
		HttpsPort           int
		TLSCertFile         string
		TLSKeyFile          string
		RedirectHttpToHttps bool
		SessionName         string
		SessionsDir         string
		SessionsAuthKey     string
		SessionsEncryptKey  string
	}
	Auth struct {
		AuthMethod    []string
		AdminUsers    []string
		HelpDeskUsers []string

		LDAP struct {
			UseAD   bool
			Servers []string
			UseTLS  bool
		}
	}
	DHCP struct {
		ConfigFile string
	}
}

func NewConfig(configFile string) (conf *Config, err error) {
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
	var con Config
	if err := toml.Unmarshal(buf, &con); err != nil {
		return nil, err
	}
	con.sourceFile = configFile
	return &con, nil
}

func (c *Config) Reload() error {
	con, err := NewConfig(c.sourceFile)
	if err != nil {
		return err
	}
	c = con
	return nil
}
