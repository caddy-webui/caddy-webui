package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/caddy-webui/caddy-webui/internal/caddy"
	"github.com/caddy-webui/caddy-webui/internal/config"
	"github.com/caddy-webui/caddy-webui/internal/database"
)

func HandleGetCaddyfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
		return
	}

	content, err := caddy.ReadCaddyfile()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "读取 Caddyfile 失败")
		return
	}

	SuccessResponse(w, "success", map[string]interface{}{
		"content": content,
	})
}

func HandleSaveCaddyfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, 40001, "请求格式不正确")
		return
	}

	if err := caddy.ValidateCaddyfile(req.Content); err != nil {
		ErrorResponse(w, http.StatusBadRequest, 40001, "Caddyfile 语法错误: "+err.Error())
		return
	}

	if err := caddy.WriteCaddyfile(req.Content); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "保存 Caddyfile 失败")
		return
	}

	if err := caddy.ReloadCaddy(req.Content); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "Caddy 重载失败，已自动恢复至上一个有效配置")
		return
	}

	SuccessResponse(w, "Caddyfile 保存成功，Caddy 已重载", nil)
}

func HandleFileUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
		return
	}

	if err := r.ParseMultipartForm(100 << 20); err != nil {
		ErrorResponse(w, http.StatusBadRequest, 40001, "解析表单数据失败")
		return
	}

	siteDomain := r.FormValue("site_domain")
	targetPath := r.FormValue("path")
	if targetPath == "" {
		targetPath = "/"
	}

	if siteDomain == "" {
		ErrorResponse(w, http.StatusBadRequest, 40001, "目标站点域名不能为空")
		return
	}

	site, err := database.GetSiteByDomain(siteDomain)
	if err != nil || site == nil {
		ErrorResponse(w, http.StatusNotFound, 40401, "目标站点不存在")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, 40001, "未找到上传文件")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	forbiddenExts := map[string]bool{".sh": true, ".bin": true, ".exe": true, ".bat": true, ".cmd": true, ".com": true}
	if forbiddenExts[ext] {
		ErrorResponse(w, http.StatusBadRequest, 40001, "不允许上传此类型的文件")
		return
	}

	destDir := filepath.Join("/opt/caddy-webui/www", siteDomain, targetPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "创建目标目录失败")
		return
	}

	destPath := filepath.Join(destDir, header.Filename)
	dst, err := os.Create(destPath)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "创建文件失败")
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "文件写入失败")
		return
	}

	config.Info("文件上传成功: %s (%d bytes)", destPath, written)

	SuccessResponse(w, "文件上传成功", map[string]interface{}{
		"path": destPath,
		"size": written,
	})
}
