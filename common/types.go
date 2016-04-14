package common

import (
	"encoding/json"
	"net/http"
)

// Config defines the configuration struct for the application
type Config struct {
	SourceFile string
	Core       struct {
		DatabaseFile             string
		SiteTitle                string
		SiteCompanyName          string
		RegistrationPolicyFile   string
		AllowManualRegistrations bool
		DefaultDeviceLimit       int
	}
	Webserver struct {
		Address            string
		Port               int
		TLSCertFile        string
		TLSKeyFile         string
		SessionName        string
		SessionsDir        string
		SessionsAuthKey    string
		SessionsEncryptKey string
	}
	Auth struct {
		AuthMethod []string
		AdminUsers []string

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

// APIStatus is an integer that states the success or failure of the request
type APIStatus int

const (
	// APIStatusOK everything went fine, no error
	APIStatusOK APIStatus = 0
	// APIStatusGenericError something went wrong but there's no specific error number for it
	APIStatusGenericError APIStatus = 1
	// APIStatusInvalidAuth failed login
	APIStatusInvalidAuth APIStatus = 10
	// APIStatusAuthNeeded no active login, but it's needed
	APIStatusAuthNeeded APIStatus = 11
	// APIStatusNotAdmin user is not an administrator
	APIStatusNotAdmin APIStatus = 12
)

// A APIResponse is returned as a JSON struct to the client
type APIResponse struct {
	Code    APIStatus
	Message string
	Data    interface{}
}

// NewAPIResponse creates an APIResponse object with status c, message m, and data d
func NewAPIResponse(c APIStatus, m string, d interface{}) *APIResponse {
	return &APIResponse{
		Code:    c,
		Message: m,
		Data:    d,
	}
}

func NewAPIOK(m string, d interface{}) *APIResponse {
	return &APIResponse{
		Code:    APIStatusOK,
		Message: m,
		Data:    d,
	}
}

// Encode the APIResponse into JSON
func (a *APIResponse) Encode() []byte {
	b, err := json.Marshal(a)
	if err != nil {
		// Do something
	}
	return b
}

func (a *APIResponse) WriteTo(w http.ResponseWriter) (int64, error) {
	r := a.Encode()
	l := len(r)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(r)
	return int64(l), nil
}
