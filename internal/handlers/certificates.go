package handlers

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/caddy-webui/caddy-webui/internal/caddy"
	"github.com/caddy-webui/caddy-webui/internal/config"
	"github.com/caddy-webui/caddy-webui/internal/database"
	"github.com/caddy-webui/caddy-webui/internal/models"
)

func HandleGetCertificates(w http.ResponseWriter, r *http.Request) {
	sites, _, err := database.ListSites(1, 1000)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "获取证书列表失败")
		return
	}

	var certs []models.CertInfo
	for _, site := range sites {
		issuer := "Let's Encrypt"
		if site.CertMode == "custom" {
			issuer = "Custom"
		}

		certs = append(certs, models.CertInfo{
			SiteID:    site.ID,
			Domain:    site.Domain,
			CertMode:  site.CertMode,
			Status:    site.CertStatus,
			ExpiresAt: site.CertExpiresAt,
			Issuer:    issuer,
		})
	}

	if certs == nil {
		certs = []models.CertInfo{}
	}

	SuccessResponse(w, "success", certs)
}

func HandleRenewCertificate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
		return
	}

	siteID, err := strconv.ParseInt(getPathParam(r.URL.Path, 4), 10, 64)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, 40001, "无效的站点 ID")
		return
	}

	site, err := database.GetSiteByID(siteID)
	if err != nil || site == nil {
		ErrorResponse(w, http.StatusNotFound, 40401, "站点不存在")
		return
	}

	if site.CertMode != "auto" {
		ErrorResponse(w, http.StatusBadRequest, 40001, "自定义上传模式的证书不支持自动续期，请上传新证书")
		return
	}

	if err := caddy.RestartCaddy(); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "证书续期失败: "+err.Error())
		return
	}

	SuccessResponse(w, "证书续期请求已发送", nil)
}

func HandleUploadCertificate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
		return
	}

	siteID, err := strconv.ParseInt(getPathParam(r.URL.Path, 4), 10, 64)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, 40001, "无效的站点 ID")
		return
	}

	site, err := database.GetSiteByID(siteID)
	if err != nil || site == nil {
		ErrorResponse(w, http.StatusNotFound, 40401, "站点不存在")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		ErrorResponse(w, http.StatusBadRequest, 40001, "解析表单数据失败")
		return
	}

	certFile, certHeader, err := r.FormFile("cert_file")
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, 40001, "证书文件和私钥文件必须同时上传")
		return
	}
	defer certFile.Close()

	keyFile, keyHeader, err := r.FormFile("key_file")
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, 40001, "证书文件和私钥文件必须同时上传")
		return
	}
	defer keyFile.Close()

	if !isValidCertExt(certHeader.Filename) {
		ErrorResponse(w, http.StatusBadRequest, 40001, "证书文件格式不正确，仅支持 .pem/.crt 格式")
		return
	}
	if !isValidKeyExt(keyHeader.Filename) {
		ErrorResponse(w, http.StatusBadRequest, 40001, "私钥文件格式不正确，仅支持 .key 格式")
		return
	}

	certBytes, err := io.ReadAll(certFile)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "读取证书文件失败")
		return
	}

	keyBytes, err := io.ReadAll(keyFile)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "读取私钥文件失败")
		return
	}

	if !isValidPEM(certBytes, "CERTIFICATE") {
		ErrorResponse(w, http.StatusBadRequest, 40001, "证书文件格式不正确，必须为 PEM 格式")
		return
	}
	if !isValidPEM(keyBytes, "PRIVATE KEY") && !isValidPEM(keyBytes, "RSA PRIVATE KEY") && !isValidPEM(keyBytes, "EC PRIVATE KEY") {
		ErrorResponse(w, http.StatusBadRequest, 40001, "私钥文件格式不正确，必须为 PEM 格式")
		return
	}

	if !validateKeyPair(certBytes, keyBytes) {
		ErrorResponse(w, http.StatusBadRequest, 40001, "证书与私钥不匹配，请确认上传的证书和私钥为同一对")
		return
	}

	expiresAt, err := extractCertExpiry(certBytes)
	if err != nil {
		config.Warn("提取证书有效期失败: %v", err)
	}

	sslDir := filepath.Join("/opt/caddy-webui/ssl", site.Domain)
	if err := os.MkdirAll(sslDir, 0700); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "创建证书目录失败")
		return
	}

	certPath := filepath.Join(sslDir, "cert.pem")
	keyPath := filepath.Join(sslDir, "key.key")

	if err := os.WriteFile(certPath, certBytes, 0600); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "保存证书文件失败")
		return
	}
	if err := os.WriteFile(keyPath, keyBytes, 0600); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "保存私钥文件失败")
		return
	}

	site.CertMode = "custom"
	site.CertFilePath = &certPath
	site.KeyFilePath = &keyPath
	site.CertStatus = "valid"
	site.CertExpiresAt = &expiresAt
	database.UpdateSite(site)

	if err := regenerateAndReload(); err != nil {
		config.Warn("Caddyfile 重载失败: %v", err)
	}

	SuccessResponse(w, "证书上传成功", map[string]interface{}{
		"cert_file_path": certPath,
		"key_file_path":  keyPath,
		"expires_at":     expiresAt,
	})
}

func HandleUpdateCertificate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
		return
	}

	siteID, err := strconv.ParseInt(getPathParam(r.URL.Path, 4), 10, 64)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, 40001, "无效的站点 ID")
		return
	}

	site, err := database.GetSiteByID(siteID)
	if err != nil || site == nil {
		ErrorResponse(w, http.StatusNotFound, 40401, "站点不存在")
		return
	}

	if site.CertMode != "custom" {
		ErrorResponse(w, http.StatusBadRequest, 40001, "站点证书模式为自动申请，无法更新证书文件")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		ErrorResponse(w, http.StatusBadRequest, 40001, "解析表单数据失败")
		return
	}

	sslDir := filepath.Join("/opt/caddy-webui/ssl", site.Domain)
	certPath := filepath.Join(sslDir, "cert.pem")
	keyPath := filepath.Join(sslDir, "key.key")

	var certBytes, keyBytes []byte

	certFile, certHeader, errCert := r.FormFile("cert_file")
	if errCert == nil {
		defer certFile.Close()
		if !isValidCertExt(certHeader.Filename) {
			ErrorResponse(w, http.StatusBadRequest, 40001, "证书文件格式不正确")
			return
		}
		certBytes, _ = io.ReadAll(certFile)
		if !isValidPEM(certBytes, "CERTIFICATE") {
			ErrorResponse(w, http.StatusBadRequest, 40001, "证书文件格式不正确")
			return
		}
	} else {
		existing, err := os.ReadFile(certPath)
		if err != nil {
			ErrorResponse(w, http.StatusBadRequest, 40001, "缺少证书文件")
			return
		}
		certBytes = existing
	}

	keyFile, keyHeader, errKey := r.FormFile("key_file")
	if errKey == nil {
		defer keyFile.Close()
		if !isValidKeyExt(keyHeader.Filename) {
			ErrorResponse(w, http.StatusBadRequest, 40001, "私钥文件格式不正确")
			return
		}
		keyBytes, _ = io.ReadAll(keyFile)
		if !isValidPEM(keyBytes, "PRIVATE KEY") && !isValidPEM(keyBytes, "RSA PRIVATE KEY") && !isValidPEM(keyBytes, "EC PRIVATE KEY") {
			ErrorResponse(w, http.StatusBadRequest, 40001, "私钥文件格式不正确")
			return
		}
	} else {
		existing, err := os.ReadFile(keyPath)
		if err != nil {
			ErrorResponse(w, http.StatusBadRequest, 40001, "缺少私钥文件")
			return
		}
		keyBytes = existing
	}

	if !validateKeyPair(certBytes, keyBytes) {
		ErrorResponse(w, http.StatusBadRequest, 40001, "证书与私钥不匹配")
		return
	}

	expiresAt, _ := extractCertExpiry(certBytes)

	if errCert == nil {
		os.WriteFile(certPath, certBytes, 0600)
	}
	if errKey == nil {
		os.WriteFile(keyPath, keyBytes, 0600)
	}

	site.CertStatus = "valid"
	site.CertExpiresAt = &expiresAt
	database.UpdateSite(site)

	if err := regenerateAndReload(); err != nil {
		config.Warn("Caddyfile 重载失败: %v", err)
	}

	SuccessResponse(w, "证书更新成功", map[string]interface{}{
		"cert_file_path": certPath,
		"key_file_path":  keyPath,
		"expires_at":     expiresAt,
	})
}

func HandleCertMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
		return
	}

	siteID, err := strconv.ParseInt(getPathParam(r.URL.Path, 4), 10, 64)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, 40001, "无效的站点 ID")
		return
	}

	site, err := database.GetSiteByID(siteID)
	if err != nil || site == nil {
		ErrorResponse(w, http.StatusNotFound, 40401, "站点不存在")
		return
	}

	var req models.CertModeRequest
	if err := jsonDecode(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, 40001, "请求格式不正确")
		return
	}

	if req.CertMode != "auto" && req.CertMode != "custom" {
		ErrorResponse(w, http.StatusBadRequest, 40001, "证书模式值不合法，仅支持 auto 或 custom")
		return
	}

	if req.CertMode == "custom" {
		if site.CertFilePath == nil || site.KeyFilePath == nil {
			ErrorResponse(w, http.StatusBadRequest, 40001, "切换到自定义上传模式需要先上传证书文件和私钥文件")
			return
		}
	}

	if req.CertMode == "auto" {
		site.CertFilePath = nil
		site.KeyFilePath = nil
	}

	site.CertMode = req.CertMode
	database.UpdateSite(site)

	if err := regenerateAndReload(); err != nil {
		config.Warn("Caddyfile 重载失败: %v", err)
	}

	modeName := "自动申请"
	if req.CertMode == "custom" {
		modeName = "自定义上传"
	}

	SuccessResponse(w, "证书模式已切换为"+modeName, map[string]interface{}{
		"cert_mode": req.CertMode,
	})
}

func isValidCertExt(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".pem" || ext == ".crt"
}

func isValidKeyExt(filename string) bool {
	return strings.ToLower(filepath.Ext(filename)) == ".key"
}

func isValidPEM(data []byte, blockType string) bool {
	block, _ := pem.Decode(data)
	return block != nil && strings.Contains(block.Type, blockType)
}

func validateKeyPair(certPEM, keyPEM []byte) bool {
	_, err := tls.X509KeyPair(certPEM, keyPEM)
	return err == nil
}

func extractCertExpiry(certPEM []byte) (string, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return "", fmt.Errorf("无法解析证书")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}
	return cert.NotAfter.Format("2006-01-02T15:04:05Z"), nil
}

func jsonDecode(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
