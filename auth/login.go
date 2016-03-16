package auth

import (
	"github.com/onesimus-systems/net-guardian/common"
)

var authFunctions = make([]authFunc, 0)

func init() {
	common.HTTPMux.HandleFunc("/login", common.NotImplementedHandler).Methods("GET")
	common.HTTPMux.HandleFunc("/login", common.NotImplementedHandler).Methods("POST")
}

type authFunc func(username, password string) bool
