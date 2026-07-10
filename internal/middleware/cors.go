package middleware

import (
	"net/http"
	"strings"
)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			host := r.Host
			if !isSameOrigin(origin, host) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"code":40301,"message":"跨域请求被拒绝","data":null}`))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func isSameOrigin(origin, host string) bool {
	origin = strings.TrimPrefix(origin, "http://")
	origin = strings.TrimPrefix(origin, "https://")
	return origin == host
}
