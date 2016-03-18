package common

// Config defines the configuration struct for the application
type Config struct {
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
	DHCP struct {
		LeasesFile string
		HostsFile  string
	}
}
