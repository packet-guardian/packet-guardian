// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package middleware

import (
	"net/http"
	"strings"

	"github.com/lfkeitel/verbose/v4"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
)

func BlacklistCheck(e *common.Environment, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Public assets are available to everyone
		if strings.HasPrefix(r.URL.Path, "/public") ||
			strings.HasPrefix(r.URL.Path, "/login") ||
			strings.HasPrefix(r.URL.Path, "/logout") {
			next.ServeHTTP(w, r)
			return
		}

		// Admin user's bypass the blacklist
		sessionUser := models.GetUserFromContext(r)
		if sessionUser.Can(models.BypassBlacklist) {
			next.ServeHTTP(w, r)
			return
		}

		ip := common.GetIPFromContext(r)
		lease, err := stores.GetLeaseStore(e).GetLeaseByIP(ip)
		if err == nil && lease.ID != 0 {
			device, err := stores.GetDeviceStore(e).GetDeviceByMAC(lease.MAC)
			if err != nil {
				e.Log.WithFields(verbose.Fields{
					"error":   err,
					"package": "middlware:blacklist",
					"mac":     lease.MAC.String(),
				}).Critical("Error getting device")
			} else if device.IsBlacklisted() {
				e.Views.NewView("blacklisted", r).Render(w, nil)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
