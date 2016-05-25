// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package middleware

import (
	"net/http"

	"github.com/onesimus-systems/packet-guardian/src/common"
)

func Cache(e *common.Environment, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != e.Config.Core.SiteDomainName {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
			w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
			w.Header().Set("Expires", "0")                                         // Proxies.
		}
		next.ServeHTTP(w, r)
	})
}
