package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/caddy-webui/caddy-webui/internal/config"
	"github.com/caddy-webui/caddy-webui/internal/database"
)

func HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
		return
	}

	username, _ := database.GetAdminUsername()
	logLevel, _ := database.GetLogLevel()
	port, _ := database.GetServerPort()

	SuccessResponse(w, "success", map[string]interface{}{
		"port":      port,
		"username":  username,
		"log_level": logLevel,
	})
}

func HandleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
		return
	}

	var req struct {
		Port     *int    `json:"port"`
		LogLevel *string `json:"log_level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, 40001, "请求格式不正确")
		return
	}

	needRestart := false

	if req.Port != nil {
		if *req.Port < 1024 || *req.Port > 65535 {
			ErrorResponse(w, http.StatusBadRequest, 40001, "端口号必须在 1024-65535 范围内")
			return
		}
		if err := database.SetServerPort(*req.Port); err != nil {
			ErrorResponse(w, http.StatusInternalServerError, 50001, "保存端口设置失败")
			return
		}
		needRestart = true
	}

	if req.LogLevel != nil {
		level := *req.LogLevel
		if level != "DEBUG" && level != "INFO" && level != "WARN" && level != "ERROR" {
			ErrorResponse(w, http.StatusBadRequest, 40001, "日志级别值不合法")
			return
		}
		if err := database.SetLogLevel(level); err != nil {
			ErrorResponse(w, http.StatusInternalServerError, 50001, "保存日志级别失败")
			return
		}
		config.SetLogLevel(level)
	}

	data := map[string]interface{}{
		"need_restart": needRestart,
	}
	if needRestart {
		data["restart_hint"] = "端口已更新，请执行 systemctl restart caddy-webui 使新端口生效"
	}

	SuccessResponse(w, "设置更新成功", data)
}
