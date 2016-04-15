package middleware

import (
	"net/http"
	"time"

	"github.com/onesimus-systems/packet-guardian/src/common"
)

func Logging(e *common.Environment, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now().Format("02/Jan/2006 15:04:05 -0700")
		e.Log.Infof("%s %s \"%s\"", r.RemoteAddr, now, r.RequestURI)
		next(w, r)
	}
}
