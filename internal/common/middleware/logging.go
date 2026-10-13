package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// wrappedResponseWriter captures the status code for logging.
type wrappedResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (w *wrappedResponseWriter) WriteHeader(statusCode int) {
	if !w.written {
		w.statusCode = statusCode
		w.written = true
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *wrappedResponseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.statusCode = http.StatusOK
		w.written = true
	}
	return w.ResponseWriter.Write(b)
}

// Logging returns a middleware that logs HTTP request details.
func Logging(log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapped := &wrappedResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)

			event := log.Info()
			if wrapped.statusCode >= 500 {
				event = log.Error()
			} else if wrapped.statusCode >= 400 {
				event = log.Warn()
			}

			event.
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("query", r.URL.RawQuery).
				Int("status", wrapped.statusCode).
				Dur("duration", duration).
				Str("remote_addr", r.RemoteAddr).
				Str("user_agent", r.UserAgent()).
				Msg("HTTP request")
		})
	}
}
