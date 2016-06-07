// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/models"
)

func BlacklistCheck(e *common.Environment, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Public assets are available to everyone
		if strings.HasPrefix(r.URL.Path, "/public") || strings.HasPrefix(r.URL.Path, "/login") {
			next.ServeHTTP(w, r)
			return
		}

		// Admin user's bypass the blacklist
		sessionUser := models.GetUserFromContext(r)
		if sessionUser.Can(models.BypassBlacklist) {
			next.ServeHTTP(w, r)
			return
		}

		ip := net.ParseIP(strings.Split(r.RemoteAddr, ":")[0])
		lease, err := models.GetLeaseByIP(e, ip)
		if err == nil && lease.ID != 0 {
			device, err := models.GetDeviceByMAC(e, lease.MAC)
			if err != nil {
				e.Log.Errorf("Error getting device for blacklist check: %s", err.Error())
			} else if device.IsBlacklisted {
				e.Views.NewView("blacklisted", r).Render(w, nil)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
