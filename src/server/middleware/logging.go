package middleware

import (
	"net/http"

	"github.com/onesimus-systems/packet-guardian/src/common"
)

func Logging(e *common.Environment, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//e.Log.GetLogger("server").Infof("%s %s \"%s\"", r.RemoteAddr, r.Method, r.RequestURI)
		// TODO: Reenable server request logging
		next.ServeHTTP(w, r)
	})
}
