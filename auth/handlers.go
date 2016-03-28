package auth

import (
	"net/http"

	"github.com/onesimus-systems/packet-guardian/common"
)

// LoginHandler handles a login POST request
func LoginHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Assume invalid until convinced otherwise
		resp := common.NewAPIResponse(common.APIStatusInvalidAuth, "Invalid login", nil)
		if IsValidLogin(e.DB, r.FormValue("username"), r.FormValue("password")) {
			resp.Code = common.APIStatusOK
			resp.Message = ""
		}
		w.Write(resp.Encode())
	}
}
