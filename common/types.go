package common

import "encoding/json"

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

type ApiStatus int

const (
	ApiStatusOK ApiStatus = iota
	ApiStatusGenericError
	ApiStatusInvalidAuth
	ApiStatusAuthNeeded
)

type ApiResponse struct {
	Code    ApiStatus
	Message string
	Data    interface{}
}

func NewApiResponse(c ApiStatus, m string, d interface{}) *ApiResponse {
	return &ApiResponse{
		Code:    c,
		Message: m,
		Data:    d,
	}
}

func (a *ApiResponse) Encode() []byte {
	b, err := json.Marshal(a)
	if err != nil {
		// Do something
	}
	return b
}
