// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package middleware

import (
	"net/http"
	"runtime"

	"github.com/onesimus-systems/packet-guardian/src/common"
)

func Panic(e *common.Environment, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(runtime.Error); ok {
					buf := make([]byte, 1024)
					runtime.Stack(buf, false)
					e.Log.WithField("Stack", string(buf)).Alert()
				}
				e.Log.Alert(r)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
