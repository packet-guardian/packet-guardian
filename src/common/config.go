// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"errors"
	"io/ioutil"
	"os"
	"time"

	"github.com/naoina/toml"
)

// Config defines the configuration struct for the application
type Config struct {
	sourceFile string
	Core       struct {
		SiteTitle          string
		SiteCompanyName    string
		SiteDomainName     string
		JobSchedulerWakeUp string
	}
	Logging struct {
		Enabled    bool
		EnableHTTP bool
		Level      string
		Path       string
	}
	Database struct {
		Type     string
		Address  string
		Username string
		Password string
	}
	Registration struct {
		RegistrationPolicyFile      string
		AllowManualRegistrations    bool
		DefaultDeviceLimit          int
		DefaultDeviceExpirationType string
		RollingExpirationLength     string
		DefaultDeviceExpiration     string
		ManualRegPlatforms          []string
	}
	Webserver struct {
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
			UseAD         bool
			Servers       []string
			VerifySSLCert bool
			DomainName    string

			BaseDN       string
			BindDN       string
			BindPassword string
			UserFilter   string
			GroupFilter  string
		}
		Radius struct {
			Servers []string
			Port    int
			Secret  string
		}
		CAS struct {
			Server string
		}
	}
	DHCP struct {
		Enabled    bool
		ConfigFile string
	}
}

func NewEmptyConfig() *Config {
	return &Config{}
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
	return setSensibleDefaults(&con)
}

func setSensibleDefaults(c *Config) (*Config, error) {
	// Anything not set here implies its zero value is the default

	// Core
	c.Core.SiteTitle = setStringOrDefault(c.Core.SiteTitle, "Packet Guardian")
	c.Core.JobSchedulerWakeUp = setStringOrDefault(c.Core.JobSchedulerWakeUp, "1h")
	if _, err := time.ParseDuration(c.Core.JobSchedulerWakeUp); err != nil {
		c.Core.JobSchedulerWakeUp = "1h"
	}

	// Logging
	c.Logging.Level = setStringOrDefault(c.Logging.Level, "notice")
	c.Logging.Path = setStringOrDefault(c.Logging.Path, "logs/pg.log")

	// Database
	c.Database.Type = setStringOrDefault(c.Database.Type, "sqlite")
	c.Database.Address = setStringOrDefault(c.Database.Address, "config/database.sqlite3")

	// Registration
	c.Registration.RegistrationPolicyFile = setStringOrDefault(c.Registration.RegistrationPolicyFile, "config/policy.txt")
	c.Registration.DefaultDeviceExpirationType = setStringOrDefault(c.Registration.DefaultDeviceExpirationType, "rolling")
	c.Registration.RollingExpirationLength = setStringOrDefault(c.Registration.RollingExpirationLength, "4380h")
	if _, err := time.ParseDuration(c.Registration.RollingExpirationLength); err != nil {
		c.Registration.RollingExpirationLength = "4380h"
	}

	// Webserver
	c.Webserver.HttpPort = setIntOrDefault(c.Webserver.HttpPort, 8080)
	c.Webserver.HttpsPort = setIntOrDefault(c.Webserver.HttpsPort, 1443)
	c.Webserver.SessionName = setStringOrDefault(c.Webserver.SessionName, "packet-guardian")
	c.Webserver.SessionsDir = setStringOrDefault(c.Webserver.SessionsDir, "sessions")

	// Authentication
	if len(c.Auth.AuthMethod) == 0 {
		c.Auth.AuthMethod = []string{"local"}
	}
	if len(c.Auth.AdminUsers) == 0 {
		c.Auth.AdminUsers = []string{"admin"}
	}
	if len(c.Auth.HelpDeskUsers) == 0 {
		c.Auth.HelpDeskUsers = []string{"helpdesk"}
	}

	// DHCP
	c.DHCP.ConfigFile = setStringOrDefault(c.DHCP.ConfigFile, "config/dhcp.conf")
	return c, nil
}

// Given string s, if it is empty, return v else return s.
func setStringOrDefault(s, v string) string {
	if s == "" {
		return v
	}
	return s
}

// Given integer s, if it is 0, return v else return s.
func setIntOrDefault(s, v int) int {
	if s == 0 {
		return v
	}
	return s
}

func (c *Config) Reload() error {
	con, err := NewConfig(c.sourceFile)
	if err != nil {
		return err
	}
	c = con
	return nil
}
