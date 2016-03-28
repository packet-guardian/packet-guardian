package common

import "encoding/json"

// Config defines the configuration struct for the application
type Config struct {
	Core struct {
		DatabaseFile    string
		SiteTitle       string
		SiteCompanyName string
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

// APIStatus is an integer that states the success or failure of the request
type APIStatus int

const (
	// APIStatusOK everything went fine, no error
	APIStatusOK APIStatus = iota
	// APIStatusGenericError something went wrong but there's no specific error number for it
	APIStatusGenericError
	// APIStatusInvalidAuth failed login
	APIStatusInvalidAuth
	// APIStatusAuthNeeded no active login, but it's needed
	APIStatusAuthNeeded
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

// Encode the APIResponse into JSON
func (a *APIResponse) Encode() []byte {
	b, err := json.Marshal(a)
	if err != nil {
		// Do something
	}
	return b
}
