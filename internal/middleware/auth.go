package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/caddy-webui/caddy-webui/internal/auth"
)

var publicPaths = []string{
	"/api/auth/login",
	"/api/auth/setup",
	"/api/auth/status",
	"/setup",
}

func isPublicPath(path string) bool {
	for _, p := range publicPaths {
		if path == p {
			return true
		}
	}
	if strings.HasPrefix(path, "/css/") ||
		strings.HasPrefix(path, "/js/") ||
		strings.HasPrefix(path, "/pages/") ||
		path == "/" ||
		path == "/index.html" ||
		strings.HasSuffix(path, ".css") ||
		strings.HasSuffix(path, ".js") ||
		strings.HasSuffix(path, ".html") ||
		strings.HasSuffix(path, ".ico") {
		return true
	}
	return false
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code":    40101,
				"message": "未提供认证令牌",
				"data":    nil,
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code":    40101,
				"message": "认证令牌格式不正确",
				"data":    nil,
			})
			return
		}

		claims, err := auth.ValidateToken(parts[1])
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code":    40101,
				"message": "认证令牌无效或已过期",
				"data":    nil,
			})
			return
		}

		r.Header.Set("X-User", claims.Subject)
		next.ServeHTTP(w, r)
	})
}
