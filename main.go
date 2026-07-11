package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/caddy-webui/caddy-webui/internal/auth"
	"github.com/caddy-webui/caddy-webui/internal/config"
	"github.com/caddy-webui/caddy-webui/internal/database"
	"github.com/caddy-webui/caddy-webui/internal/handlers"
	"github.com/caddy-webui/caddy-webui/internal/middleware"
)

//go:embed static
var staticFiles embed.FS

func main() {
	if err := config.Load(); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		return
	}

	if err := config.InitLogger(); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		return
	}
	defer config.CloseLogger()

	config.Info("Caddy WebUI 管理面板启动中...")

	if err := database.Init(); err != nil {
		config.Error("数据库初始化失败: %v", err)
		return
	}
	defer database.Close()

	if err := auth.InitJWT(); err != nil {
		config.Error("JWT 初始化失败: %v", err)
		return
	}

	mux := http.NewServeMux()
	registerAPIRoutes(mux)
	registerStaticRoutes(mux)

	var handler http.Handler = mux
	handler = middleware.Recovery(handler)
	handler = middleware.Logger(handler)
	handler = middleware.CORS(handler)
	handler = middleware.Auth(handler)
	handler = middleware.LockCheck(handler)

	addr := config.Addr()
	config.Info("面板监听地址: %s", addr)
	config.Info("访问 http://%s 开始使用", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		config.Error("服务启动失败: %v", err)
	}
}

func registerAPIRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/auth/setup", handlers.HandleSetup)
	mux.HandleFunc("/api/auth/login", handlers.HandleLogin)
	mux.HandleFunc("/api/auth/password", handlers.HandleChangePassword)
	mux.HandleFunc("/api/auth/status", handlers.HandleAuthStatus)

	mux.HandleFunc("/api/dashboard", handlers.HandleDashboard)

	mux.HandleFunc("/api/sites", handleSitesRouter)
	mux.HandleFunc("/api/sites/", handleSiteDetailRouter)

	mux.HandleFunc("/api/caddy/status", handlers.HandleCaddyStatus)
	mux.HandleFunc("/api/caddy/start", handlers.HandleCaddyStart)
	mux.HandleFunc("/api/caddy/stop", handlers.HandleCaddyStop)
	mux.HandleFunc("/api/caddy/restart", handlers.HandleCaddyRestart)
	mux.HandleFunc("/api/caddy/reload", handlers.HandleCaddyReload)

	mux.HandleFunc("/api/certificates", handlers.HandleGetCertificates)
	mux.HandleFunc("/api/certificates/", handleCertificateRouter)

	mux.HandleFunc("/api/settings", handleSettingsRouter)

	mux.HandleFunc("/api/files/caddyfile", handleCaddyfileRouter)
	mux.HandleFunc("/api/files/upload", handlers.HandleFileUpload)
}

func registerStaticRoutes(mux *http.ServeMux) {
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		config.Error("静态资源加载失败: %v", err)
		return
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		// fs.ReadFile 要求路径不带前导 /
		fsPath := strings.TrimPrefix(path, "/")

		data, err := fs.ReadFile(sub, fsPath)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", mimeByExt(path))
		w.Write(data)
	})
}

func mimeByExt(path string) string {
	switch filepath.Ext(path) {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".eot":
		return "application/vnd.ms-fontobject"
	default:
		return "application/octet-stream"
	}
}

func handleSitesRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handlers.HandleGetSites(w, r)
	case http.MethodPost:
		handlers.HandleCreateSite(w, r)
	default:
		handlers.ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
	}
}

func handleSiteDetailRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasSuffix(path, "/toggle") {
		handlers.HandleToggleSite(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		handlers.HandleGetSite(w, r)
	case http.MethodPut:
		handlers.HandleUpdateSite(w, r)
	case http.MethodDelete:
		handlers.HandleDeleteSite(w, r)
	default:
		handlers.ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
	}
}

func handleCertificateRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasSuffix(path, "/renew") {
		handlers.HandleRenewCertificate(w, r)
		return
	}
	if strings.HasSuffix(path, "/upload") {
		handlers.HandleUploadCertificate(w, r)
		return
	}
	if strings.HasSuffix(path, "/update") {
		handlers.HandleUpdateCertificate(w, r)
		return
	}
	if strings.HasSuffix(path, "/mode") {
		handlers.HandleCertMode(w, r)
		return
	}
	handlers.HandleGetCertificates(w, r)
}

func handleSettingsRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handlers.HandleGetSettings(w, r)
	case http.MethodPut:
		handlers.HandleUpdateSettings(w, r)
	default:
		handlers.ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
	}
}

func handleCaddyfileRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handlers.HandleGetCaddyfile(w, r)
	case http.MethodPut:
		handlers.HandleSaveCaddyfile(w, r)
	default:
		handlers.ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
	}
}
