package middleware

import (
	"net/http"

	"github.com/onesimus-systems/packet-guardian/src/common"
)

// responseWriter is an http.ResponseWriter that keeps track of the length
// of its response as well as the request's status returned to the client
type responseWriter struct {
	http.ResponseWriter
	length int
	status int
}

func (w *responseWriter) Write(b []byte) (n int, err error) {
	n, err = w.ResponseWriter.Write(b)
	w.length += n
	return
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func Logging(e *common.Environment, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w = &responseWriter{w, 0, 200}
		next.ServeHTTP(w, r)
		if e.Config.Webserver.EnableLogging {
			resp := w.(*responseWriter)
			e.Log.GetLogger("server").Infof(
				"%s %s \"%s\" %d %d",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
				resp.status,
				resp.length,
			)
		}
	})
}
