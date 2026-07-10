package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/caddy-webui/caddy-webui/internal/auth"
)

func LockCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/login" {
			next.ServeHTTP(w, r)
			return
		}

		locked, remaining, err := auth.CheckAccountLock()
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		if locked {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code":    40301,
				"message": "账号已锁定，请等待后再试",
				"data": map[string]interface{}{
					"remaining_minutes": remaining,
				},
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
