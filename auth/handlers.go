package auth

import (
	"net/http"

	"github.com/onesimus-systems/packet-guardian/common"
)

func LoginHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsValidLogin(e.DB, r.FormValue("username"), r.FormValue("password")) {
			http.Error(w, "", http.StatusForbidden)
		}
	}
}
