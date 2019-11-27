package middleware

import (
	"net/http"

	"github.com/lfkeitel/verbose/v4"
	"github.com/packet-guardian/dhcp-lib"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
)

func CheckGuestReg(next http.Handler, e *common.Environment, leases stores.LeaseStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := common.GetIPFromContext(r)
		reg, err := dhcp.IsRegisteredByIP(leases, ip)
		if err != nil {
			e.Log.WithFields(verbose.Fields{
				"error":   err,
				"package": "middleware:guest",
				"ip":      ip.String(),
			}).Error("Error getting registration status")
		}

		if reg {
			data := map[string]interface{}{
				"msg":   "This device is already registered",
				"error": true,
			}
			e.Views.NewView("user-register-guest-msg", r).Render(w, data)
			return
		}

		next.ServeHTTP(w, r)
	})
}
