package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog"
)

// Recovery returns a middleware that recovers from panics and logs them.
func Recovery(log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error().
						Interface("panic", rec).
						Str("stack", string(debug.Stack())).
						Str("method", r.Method).
						Str("path", r.URL.Path).
						Msg("panic recovered")

					http.Error(w, `{"success":false,"error":{"code":"INTERNAL_ERROR","message":"an unexpected error occurred"}}`, http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
