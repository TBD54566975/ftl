package log

import (
	"net/http"

	"golang.org/x/exp/slog"
)

// Middleware to inject a logger into the context.
func Middleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := ContextWithLogger(r.Context(), logger)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
