package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/caddy-webui/caddy-webui/internal/config"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				config.Error("panic recovered: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"code":    50001,
					"message": "服务器内部错误",
					"data":    nil,
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
