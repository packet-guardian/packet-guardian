// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// PageSize is the number of items per page
var PageSize = 30

// Config defines the configuration struct for the application
type Config struct {
	sourceFile string
	Core       struct {
		SiteTitle          string
		SiteCompanyName    string
		SiteDomainName     string
		SiteFooterText     string
		JobSchedulerWakeUp string
		PageSize           int
		Debug              bool
	}
	Logging struct {
		Enabled    bool
		EnableHTTP bool
		Level      string
		Path       string
	}
	Database struct {
		Type         string
		Address      string
		Port         int
		Username     string
		Password     string
		Name         string
		Retry        int
		RetryTimeout string
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
	Guest struct {
		Enabled              bool
		GuestOnly            bool
		DeviceLimit          int
		DeviceExpirationType string
		DeviceExpiration     string
		Checker              string
		VerifyCodeExpiration int
		DisableCaptcha       bool
		RegPageHeader        string

		Email struct {
		}

		Twilio struct {
			AccountSID  string
			AuthToken   string
			PhoneNumber string
		}

		Smseagle struct {
			Address      string
			Username     string
			Password     string
			HighPriority int
			FlashMsg     int
		}
	}
	Webserver struct {
		Address             string
		HTTPPort            int
		HTTPSPort           int
		TLSCertFile         string
		TLSKeyFile          string
		RedirectHTTPToHTTPS bool
		SessionStore        string
		SessionName         string
		SessionsDir         string
		SessionsAuthKey     string
		SessionsEncryptKey  string
		CustomDataDir       string
	}
	Auth struct {
		AuthMethod        []string
		AdminUsers        []string
		HelpDeskUsers     []string
		ReadOnlyUsers     []string
		APIReadOnlyUsers  []string
		APIReadWriteUsers []string
		APIStatusUsers    []string

		LDAP struct {
			Server             string
			Port               int
			UseSSL             bool
			InsecureSkipVerify bool
			SkipTLS            bool
			DomainName         string
		}
		Radius struct {
			Servers []string
			Port    int
			Secret  string
		}
		CAS struct {
			Server string
		}
		Openid struct {
			Server           string
			ClientID         string
			ClientSecret     string
			StripDomain      bool
			AuthorizeEndoint string `toml:"-"`
			TokenEndoint     string `toml:"-"`
			UserinfoEndpoint string `toml:"-"`
		}
	}
	DHCP struct {
		ConfigFile string
	}
	Email struct {
		Address     string
		Port        int
		Username    string
		Password    string
		FromAddress string
		ToAddresses []string
	}
}

// FindConfigFile searches for a configuration file. The order of search is
// environment, current dir, home dir, and /etc.
func FindConfigFile() string {
	filename := ""

	if os.Getenv("PG_CONFIG") != "" && FileExists(os.Getenv("PG_CONFIG")) {
		filename = os.Getenv("PG_CONFIG")
	} else if FileExists("./config.toml") {
		filename = "./config.toml"
	} else if FileExists("./config/config.toml") {
		filename = "./config/config.toml"
	} else if FileExists(os.ExpandEnv("$HOME/.pg/config.toml")) {
		filename = os.ExpandEnv("$HOME/.pg/config.toml")
	} else if FileExists("/etc/packet-guardian/config.toml") {
		filename = "/etc/packet-guardian/config.toml"
	}

	return filename
}

// NewEmptyConfig returns an empty config with type defaults only.
func NewEmptyConfig() *Config {
	return &Config{}
}

// NewConfig reads the given filename into a Config. If filename is empty,
// the config is looked for in the documented order.
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

	var con Config
	if _, err := toml.DecodeFile(configFile, &con); err != nil {
		return nil, err
	}
	con.sourceFile = configFile

	c, err := setSensibleDefaults(&con)
	if err != nil {
		return nil, err
	}

	PageSize = c.Core.PageSize
	return c, nil
}

func setSensibleDefaults(c *Config) (*Config, error) {
	var err error
	// Anything not set here implies its zero value is the default

	// Core
	c.Core.SiteTitle = setStringOrDefault(c.Core.SiteTitle, "Packet Guardian")
	c.Core.SiteFooterText = setStringOrDefault(c.Core.SiteFooterText, "The Guardian of Packets")
	c.Core.JobSchedulerWakeUp = setStringOrDefault(c.Core.JobSchedulerWakeUp, "1h")
	if _, err := time.ParseDuration(c.Core.JobSchedulerWakeUp); err != nil {
		c.Core.JobSchedulerWakeUp = "1h"
	}
	c.Core.PageSize = setIntOrDefault(c.Core.PageSize, 30)
	if c.Core.SiteDomainName != "" {
		c.Core.SiteDomainName, err = validateURL(c.Core.SiteDomainName, "CAS server")
		if err != nil {
			return nil, err
		}
	}

	// Logging
	c.Logging.Level = setStringOrDefault(c.Logging.Level, "notice")
	c.Logging.Path = setStringOrDefault(c.Logging.Path, "logs/pg.log")

	// Database
	c.Database.Type = setStringOrDefault(c.Database.Type, "mysql")
	c.Database.Address = setStringOrDefault(c.Database.Address, "localhost")
	c.Database.RetryTimeout = setStringOrDefault(c.Database.RetryTimeout, "10s")

	// Registration
	c.Registration.RegistrationPolicyFile = setStringOrDefault(c.Registration.RegistrationPolicyFile, "config/policy.txt")
	c.Registration.DefaultDeviceExpirationType = setStringOrDefault(c.Registration.DefaultDeviceExpirationType, "rolling")
	c.Registration.RollingExpirationLength = setStringOrDefault(c.Registration.RollingExpirationLength, "4380h")
	if _, err := time.ParseDuration(c.Registration.RollingExpirationLength); err != nil {
		c.Registration.RollingExpirationLength = "4380h"
	}

	// Guest registrations
	c.Guest.DeviceExpirationType = setStringOrDefault(c.Guest.DeviceExpirationType, "daily")
	c.Guest.DeviceExpiration = setStringOrDefault(c.Guest.DeviceExpiration, "24:00")
	c.Guest.Checker = setStringOrDefault(c.Guest.Checker, "email")
	c.Guest.VerifyCodeExpiration = setIntOrDefault(c.Guest.VerifyCodeExpiration, 3)

	// Webserver
	c.Webserver.HTTPPort = setIntOrDefault(c.Webserver.HTTPPort, 80)
	c.Webserver.HTTPSPort = setIntOrDefault(c.Webserver.HTTPSPort, 443)
	c.Webserver.SessionName = setStringOrDefault(c.Webserver.SessionName, "packet-guardian")
	c.Webserver.SessionsDir = setStringOrDefault(c.Webserver.SessionsDir, "sessions")
	c.Webserver.SessionStore = setStringOrDefault(c.Webserver.SessionStore, "filesystem")
	c.Webserver.CustomDataDir = setStringOrDefault(c.Webserver.CustomDataDir, "")

	// Authentication
	if len(c.Auth.AuthMethod) == 0 {
		c.Auth.AuthMethod = []string{"local"}
	}

	if len(c.Auth.AdminUsers) > 0 {
		fmt.Println("Setting Auth.AdminUsers is deprecated and no longer used")
	}
	if len(c.Auth.HelpDeskUsers) > 0 {
		fmt.Println("Setting Auth.HelpDeskUsers is deprecated and no longer used")
	}
	if len(c.Auth.ReadOnlyUsers) > 0 {
		fmt.Println("Setting Auth.ReadOnlyUsers is deprecated and no longer used")
	}
	if len(c.Auth.APIReadOnlyUsers) > 0 {
		fmt.Println("Setting Auth.APIReadOnlyUsers is deprecated and no longer used")
	}
	if len(c.Auth.APIReadWriteUsers) > 0 {
		fmt.Println("Setting Auth.APIReadWriteUsers is deprecated and no longer used")
	}
	if len(c.Auth.APIStatusUsers) > 0 {
		fmt.Println("Setting Auth.APIStatusUsers is deprecated and no longer used")
	}

	c.Auth.LDAP.Server = setStringOrDefault(c.Auth.LDAP.Server, "127.0.0.1")
	c.Auth.LDAP.Port = setIntOrDefault(c.Auth.LDAP.Port, 389)
	if c.Auth.LDAP.Port == 636 {
		c.Auth.LDAP.UseSSL = true
	}

	c.Auth.CAS.Server, err = validateURL(c.Auth.CAS.Server, "CAS server")
	if err != nil {
		return nil, err
	}

	if c.Auth.CAS.Server != "" {
		if c.Core.SiteDomainName == "" {
			fmt.Println("CAS is not available without SiteDomainName set")
			c.Auth.CAS.Server = ""
		}
	}

	if c.Auth.Openid.Server != "" {
		if c.Core.SiteDomainName == "" {
			fmt.Println("OpenID Connect is not available without SiteDomainName set")
			c.Auth.Openid.Server = ""
		} else {
			c.Auth.Openid.Server, err = validateURL(c.Auth.Openid.Server, "OpenID server")
			if err != nil {
				return nil, err
			}

			if c.Auth.Openid.ClientID == "" {
				return nil, errors.New("OpenID server defined but no client ID configured")
			}

			if c.Auth.Openid.ClientSecret == "" {
				return nil, errors.New("OpenID server defined but no client secret configured")
			}

			if err := getOpenIDPaths(c); err != nil {
				return nil, err
			}
		}
	}

	// DHCP
	c.DHCP.ConfigFile = setStringOrDefault(c.DHCP.ConfigFile, "config/dhcp.conf")

	// Email
	if c.Email.Address != "" {
		if c.Email.FromAddress == "" {
			return nil, errors.New("Email.FromAddress cannot be empty")
		}
		if len(c.Email.ToAddresses) == 0 {
			return nil, errors.New("Email.ToAddresses cannot be empty")
		}
		c.Email.Port = setIntOrDefault(c.Email.Port, 25)
	}
	return c, nil
}

type openIDDiscoveryConfResp struct {
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	UserinfoEndpoint      string   `json:"userinfo_endpoint"`
	TokenEndpoint         string   `json:"token_endpoint"`
	ScopesSupported       []string `json:"scopes_supported"`
}

func getOpenIDPaths(c *Config) error {
	fmt.Println("Discovering OpenID Configuration")

	configPath := fmt.Sprintf("%s/.well-known/openid-configuration", c.Auth.Openid.Server)
	fmt.Println(configPath)
	req, err := http.NewRequest("GET", configPath, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("Error getting OpenID server configuration: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Non 200 response while getting OpenID server configuration")
	}

	var discoResp openIDDiscoveryConfResp
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&discoResp); err != nil {
		return fmt.Errorf("Error decoding OpenID server configuration: %s", err.Error())
	}

	if !StringInSlice("profile", discoResp.ScopesSupported) {
		return errors.New("OpenID server doesn't support the profile scope")
	}

	if !StringInSlice("email", discoResp.ScopesSupported) {
		return errors.New("OpenID server doesn't support the email scope")
	}

	fmt.Printf("OpenID discovered auth endpoint: %s\n", discoResp.AuthorizationEndpoint)
	fmt.Printf("OpenID discovered token endpoint: %s\n", discoResp.TokenEndpoint)
	fmt.Printf("OpenID discovered userinfo endpoint: %s\n", discoResp.UserinfoEndpoint)

	c.Auth.Openid.AuthorizeEndoint = discoResp.AuthorizationEndpoint
	c.Auth.Openid.TokenEndoint = discoResp.TokenEndpoint
	c.Auth.Openid.UserinfoEndpoint = discoResp.UserinfoEndpoint
	return nil
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

func validateURL(path, description string) (string, error) {
	path = strings.TrimRight(path, "/")
	if _, err := url.Parse(path); err != nil {
		return "", fmt.Errorf("Invalid %s URL", description)
	}
	return path, nil
}
