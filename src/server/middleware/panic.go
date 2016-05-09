package middleware

import (
	"net/http"

	"github.com/onesimus-systems/packet-guardian/src/common"
)

func Panic(e *common.Environment, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				e.Log.Alert(r)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
