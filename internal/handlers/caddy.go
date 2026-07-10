package handlers

import (
	"net/http"

	"github.com/caddy-webui/caddy-webui/internal/caddy"
)

func HandleCaddyStatus(w http.ResponseWriter, r *http.Request) {
	status, _ := caddy.GetCaddyStatus()
	version := caddy.GetCaddyVersion()

	SuccessResponse(w, "success", map[string]interface{}{
		"status":  status,
		"version": version,
	})
}

func HandleCaddyStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
		return
	}

	status, _ := caddy.GetCaddyStatus()
	if status == "running" {
		ErrorResponse(w, http.StatusConflict, 40901, "Caddy 正在运行中")
		return
	}

	if err := caddy.StartCaddy(); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "Caddy 启动失败: "+err.Error())
		return
	}

	SuccessResponse(w, "Caddy 启动成功", nil)
}

func HandleCaddyStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
		return
	}

	status, _ := caddy.GetCaddyStatus()
	if status == "stopped" {
		ErrorResponse(w, http.StatusConflict, 40901, "Caddy 已处于停止状态")
		return
	}

	if err := caddy.StopCaddy(); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "Caddy 停止失败: "+err.Error())
		return
	}

	SuccessResponse(w, "Caddy 停止成功", nil)
}

func HandleCaddyRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
		return
	}

	if err := caddy.RestartCaddy(); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "Caddy 重启失败: "+err.Error())
		return
	}

	SuccessResponse(w, "Caddy 重启成功", nil)
}

func HandleCaddyReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
		return
	}

	content, err := caddy.GenerateCaddyfile()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "生成 Caddyfile 失败")
		return
	}

	if err := caddy.WriteCaddyfile(content); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "写入 Caddyfile 失败")
		return
	}

	if err := caddy.ReloadCaddy(content); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "Caddy 重载失败，已恢复原配置")
		return
	}

	SuccessResponse(w, "Caddy 配置重载成功", nil)
}
