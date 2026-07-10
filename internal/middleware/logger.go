package middleware

import (
	"net/http"
	"time"

	"github.com/caddy-webui/caddy-webui/internal/config"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: 200}

		next.ServeHTTP(rw, r)

		elapsed := time.Since(start).Milliseconds()
		config.Info("%s %s %d %dms", r.Method, r.URL.Path, rw.statusCode, elapsed)
	})
}
