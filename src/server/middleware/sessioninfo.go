// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package middleware

import (
	"net/http"

	"github.com/lfkeitel/verbose"
	"github.com/usi-lfkeitel/packet-guardian/src/common"
	"github.com/usi-lfkeitel/packet-guardian/src/models"
)

func SetSessionInfo(e *common.Environment, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := e.Sessions.GetSession(r)
		sessionUser, err := models.GetUserByUsername(e, session.GetString("username"))
		if err != nil {
			e.Log.WithFields(verbose.Fields{
				"error":    err,
				"package":  "middleware:session",
				"username": session.GetString("username"),
			}).Error("Error getting session user")
		}
		r = common.SetSessionToContext(r, session)
		r = common.SetEnvironmentToContext(r, e)
		r = models.SetUserToContext(r, sessionUser)

		// If running behind a proxy, set the RemoteAddr to the real address
		if r.Header.Get("X-Real-IP") != "" {
			r.RemoteAddr = r.Header.Get("X-Real-IP")
		}
		r = common.SetIPToContext(r)

		next.ServeHTTP(w, r)

		sessionUser.Release()
	})
}
