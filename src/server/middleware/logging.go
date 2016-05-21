package middleware

import (
	"net/http"
	"time"

	"github.com/onesimus-systems/packet-guardian/src/common"
)

// responseWriter is an http.ResponseWriter that keeps track of the length
// of its response as well as the request's status returned to the client
type responseWriter struct {
	http.ResponseWriter
	length    int
	status    int
	startTime time.Time
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

func (w *responseWriter) requestTime() time.Duration {
	return time.Since(w.startTime)
}

func Logging(e *common.Environment, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If HTTP logging is disabled, no need in wasting the space
		if !e.Config.Logging.EnableHTTP {
			next.ServeHTTP(w, r)
			return
		}

		newW := &responseWriter{
			ResponseWriter: w,
			status:         200,
			startTime:      time.Now(),
		}

		next.ServeHTTP(newW, r)

		if r.TLS != nil {
			// If the server is running on TLS, tell the client to prefer it
			w.Header().Add("Strict-Transport-Security", "max-age=63072000;")
		}

		e.Log.GetLogger("server").Infof(
			"%s %s \"%s\" %d %d %s",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			newW.status,
			newW.length,
			newW.requestTime().String(),
		)
	})
}
