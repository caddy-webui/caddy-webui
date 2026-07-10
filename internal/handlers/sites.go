package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/caddy-webui/caddy-webui/internal/caddy"
	"github.com/caddy-webui/caddy-webui/internal/config"
	"github.com/caddy-webui/caddy-webui/internal/database"
	"github.com/caddy-webui/caddy-webui/internal/models"
)

var domainRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

func HandleCreateSite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
		return
	}

	var req models.CreateSiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, 40001, "请求格式不正确")
		return
	}

	if !domainRegex.MatchString(req.Domain) {
		ErrorResponse(w, http.StatusBadRequest, 40001, "域名格式不正确")
		return
	}

	existing, _ := database.GetSiteByDomain(req.Domain)
	if existing != nil {
		ErrorResponse(w, http.StatusConflict, 40901, "该域名已存在")
		return
	}

	if req.ProxyTarget != "" && !strings.HasPrefix(req.ProxyTarget, "http://") && !strings.HasPrefix(req.ProxyTarget, "https://") {
		ErrorResponse(w, http.StatusBadRequest, 40001, "代理目标 URL 格式不正确")
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	certMode := "auto"
	if req.CertMode != "" {
		certMode = req.CertMode
	}

	var proxyConfigJSON string
	if req.ProxyConfig != nil {
		data, _ := json.Marshal(req.ProxyConfig)
		proxyConfigJSON = string(data)
	} else {
		proxyConfigJSON = "{}"
	}

	site := &models.Site{
		Domain:      req.Domain,
		Enabled:     enabled,
		ProxyTarget: req.ProxyTarget,
		ProxyConfig: proxyConfigJSON,
		CertMode:    certMode,
		CertStatus:  "none",
	}

	if err := database.CreateSite(site); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "创建站点失败")
		return
	}

	if site.Enabled {
		if err := regenerateAndReload(); err != nil {
			config.Warn("Caddyfile 重载失败: %v", err)
		}
	}

	SuccessResponse(w, "站点创建成功", map[string]interface{}{
		"id":     site.ID,
		"domain": site.Domain,
	})
}

func HandleGetSite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
		return
	}

	id, err := strconv.ParseInt(getPathParam(r.URL.Path, 3), 10, 64)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, 40001, "无效的站点 ID")
		return
	}

	site, err := database.GetSiteByID(id)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "查询站点失败")
		return
	}
	if site == nil {
		ErrorResponse(w, http.StatusNotFound, 40401, "站点不存在")
		return
	}

	SuccessResponse(w, "success", site)
}

func HandleUpdateSite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
		return
	}

	id, err := strconv.ParseInt(getPathParam(r.URL.Path, 3), 10, 64)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, 40001, "无效的站点 ID")
		return
	}

	site, err := database.GetSiteByID(id)
	if err != nil || site == nil {
		ErrorResponse(w, http.StatusNotFound, 40401, "站点不存在")
		return
	}

	var req models.UpdateSiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, 40001, "请求格式不正确")
		return
	}

	if req.Domain != nil {
		if !domainRegex.MatchString(*req.Domain) {
			ErrorResponse(w, http.StatusBadRequest, 40001, "域名格式不正确")
			return
		}
		existing, _ := database.GetSiteByDomain(*req.Domain)
		if existing != nil && existing.ID != id {
			ErrorResponse(w, http.StatusConflict, 40901, "该域名已存在")
			return
		}
		site.Domain = *req.Domain
	}

	if req.ProxyTarget != nil {
		if *req.ProxyTarget != "" && !strings.HasPrefix(*req.ProxyTarget, "http://") && !strings.HasPrefix(*req.ProxyTarget, "https://") {
			ErrorResponse(w, http.StatusBadRequest, 40001, "代理目标 URL 格式不正确")
			return
		}
		site.ProxyTarget = *req.ProxyTarget
	}

	if req.Enabled != nil {
		site.Enabled = *req.Enabled
	}

	if req.CertMode != nil {
		site.CertMode = *req.CertMode
	}

	if req.ProxyConfig != nil {
		data, _ := json.Marshal(req.ProxyConfig)
		site.ProxyConfig = string(data)
	}

	if err := database.UpdateSite(site); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "更新站点失败")
		return
	}

	if err := regenerateAndReload(); err != nil {
		config.Warn("Caddyfile 重载失败: %v", err)
	}

	SuccessResponse(w, "站点更新成功", nil)
}

func HandleDeleteSite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
		return
	}

	id, err := strconv.ParseInt(getPathParam(r.URL.Path, 3), 10, 64)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, 40001, "无效的站点 ID")
		return
	}

	site, err := database.GetSiteByID(id)
	if err != nil || site == nil {
		ErrorResponse(w, http.StatusNotFound, 40401, "站点不存在")
		return
	}

	if err := database.DeleteSite(id); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "删除站点失败")
		return
	}

	if err := regenerateAndReload(); err != nil {
		config.Warn("Caddyfile 重载失败: %v", err)
	}

	SuccessResponse(w, "站点删除成功", nil)
}

func HandleToggleSite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
		return
	}

	id, err := strconv.ParseInt(getPathParam(r.URL.Path, 4), 10, 64)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, 40001, "无效的站点 ID")
		return
	}

	site, err := database.GetSiteByID(id)
	if err != nil || site == nil {
		ErrorResponse(w, http.StatusNotFound, 40401, "站点不存在")
		return
	}

	var req models.ToggleSiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, 40001, "请求格式不正确")
		return
	}

	if err := database.ToggleSiteEnabled(id, req.Enabled); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "切换站点状态失败")
		return
	}

	if err := regenerateAndReload(); err != nil {
		config.Warn("Caddyfile 重载失败: %v", err)
	}

	SuccessResponse(w, "站点状态已更新", nil)
}

func regenerateAndReload() error {
	content, err := caddy.GenerateCaddyfile()
	if err != nil {
		return err
	}

	if err := caddy.WriteCaddyfile(content); err != nil {
		return err
	}

	return caddy.ReloadCaddy(content)
}

func getPathParam(path string, pos int) string {
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	if len(parts) > pos {
		return parts[pos]
	}
	return ""
}
