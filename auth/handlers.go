package auth

import (
	"net/http"

	"github.com/onesimus-systems/packet-guardian/common"
)

func LoginHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Assume invalid until convinced otherwise
		resp := common.NewApiResponse(common.ApiStatusInvalidAuth, "Invalid login", nil)
		if IsValidLogin(e.DB, r.FormValue("username"), r.FormValue("password")) {
			resp.Code = common.ApiStatusOK
			resp.Message = ""
		}
		w.Write(resp.Encode())
	}
}
