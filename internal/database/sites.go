package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/caddy-webui/caddy-webui/internal/models"
)

func CreateSite(site *models.Site) error {
	proxyConfigJSON := "{}"
	if site.ProxyConfig != "" {
		proxyConfigJSON = site.ProxyConfig
	}

	result, err := DB.Exec(
		`INSERT INTO sites (domain, enabled, proxy_target, proxy_config, cert_mode, cert_status, cert_file_path, key_file_path)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		site.Domain, boolToInt(site.Enabled), site.ProxyTarget, proxyConfigJSON,
		site.CertMode, site.CertStatus, site.CertFilePath, site.KeyFilePath,
	)
	if err != nil {
		return fmt.Errorf("创建站点失败: %w", err)
	}

	id, _ := result.LastInsertId()
	site.ID = id
	return nil
}

func GetSiteByID(id int64) (*models.Site, error) {
	site := &models.Site{}
	var enabled int
	var certExpiresAt, certFilePath, keyFilePath sql.NullString

	err := DB.QueryRow(
		`SELECT id, domain, enabled, proxy_target, proxy_config, cert_mode, cert_status,
		        cert_expires_at, cert_file_path, key_file_path, created_at, updated_at
		 FROM sites WHERE id = ?`, id,
	).Scan(&site.ID, &site.Domain, &enabled, &site.ProxyTarget, &site.ProxyConfig,
		&site.CertMode, &site.CertStatus, &certExpiresAt, &certFilePath, &keyFilePath,
		&site.CreatedAt, &site.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	site.Enabled = intToBool(enabled)
	site.CertExpiresAt = nullStringToPtr(certExpiresAt)
	site.CertFilePath = nullStringToPtr(certFilePath)
	site.KeyFilePath = nullStringToPtr(keyFilePath)
	return site, nil
}

func GetSiteByDomain(domain string) (*models.Site, error) {
	site := &models.Site{}
	var enabled int
	var certExpiresAt, certFilePath, keyFilePath sql.NullString

	err := DB.QueryRow(
		`SELECT id, domain, enabled, proxy_target, proxy_config, cert_mode, cert_status,
		        cert_expires_at, cert_file_path, key_file_path, created_at, updated_at
		 FROM sites WHERE domain = ?`, domain,
	).Scan(&site.ID, &site.Domain, &enabled, &site.ProxyTarget, &site.ProxyConfig,
		&site.CertMode, &site.CertStatus, &certExpiresAt, &certFilePath, &keyFilePath,
		&site.CreatedAt, &site.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	site.Enabled = intToBool(enabled)
	site.CertExpiresAt = nullStringToPtr(certExpiresAt)
	site.CertFilePath = nullStringToPtr(certFilePath)
	site.KeyFilePath = nullStringToPtr(keyFilePath)
	return site, nil
}

func ListSites(page, pageSize int) ([]models.Site, int64, error) {
	var total int64
	DB.QueryRow("SELECT COUNT(*) FROM sites").Scan(&total)

	offset := (page - 1) * pageSize
	rows, err := DB.Query(
		`SELECT id, domain, enabled, proxy_target, proxy_config, cert_mode, cert_status,
		        cert_expires_at, cert_file_path, key_file_path, created_at, updated_at
		 FROM sites ORDER BY id DESC LIMIT ? OFFSET ?`, pageSize, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var sites []models.Site
	for rows.Next() {
		var site models.Site
		var enabled int
		var certExpiresAt, certFilePath, keyFilePath sql.NullString

		err := rows.Scan(&site.ID, &site.Domain, &enabled, &site.ProxyTarget, &site.ProxyConfig,
			&site.CertMode, &site.CertStatus, &certExpiresAt, &certFilePath, &keyFilePath,
			&site.CreatedAt, &site.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		site.Enabled = intToBool(enabled)
		site.CertExpiresAt = nullStringToPtr(certExpiresAt)
		site.CertFilePath = nullStringToPtr(certFilePath)
		site.KeyFilePath = nullStringToPtr(keyFilePath)
		sites = append(sites, site)
	}

	return sites, total, nil
}

func UpdateSite(site *models.Site) error {
	proxyConfigJSON := "{}"
	if site.ProxyConfig != "" {
		proxyConfigJSON = site.ProxyConfig
	}

	now := time.Now().Format(time.RFC3339)
	_, err := DB.Exec(
		`UPDATE sites SET domain = ?, enabled = ?, proxy_target = ?, proxy_config = ?,
		 cert_mode = ?, cert_status = ?, cert_expires_at = ?, cert_file_path = ?,
		 key_file_path = ?, updated_at = ? WHERE id = ?`,
		site.Domain, boolToInt(site.Enabled), site.ProxyTarget, proxyConfigJSON,
		site.CertMode, site.CertStatus, ptrToNullString(site.CertExpiresAt),
		ptrToNullString(site.CertFilePath), ptrToNullString(site.KeyFilePath),
		now, site.ID,
	)
	return err
}

func DeleteSite(id int64) error {
	_, err := DB.Exec("DELETE FROM sites WHERE id = ?", id)
	return err
}

func ToggleSiteEnabled(id int64, enabled bool) error {
	now := time.Now().Format(time.RFC3339)
	_, err := DB.Exec("UPDATE sites SET enabled = ?, updated_at = ? WHERE id = ?",
		boolToInt(enabled), now, id)
	return err
}

func CountSitesByStatus() (total, enabled, disabled int64, err error) {
	DB.QueryRow("SELECT COUNT(*) FROM sites").Scan(&total)
	DB.QueryRow("SELECT COUNT(*) FROM sites WHERE enabled = 1").Scan(&enabled)
	disabled = total - enabled
	return
}

func CountCertificatesByStatus() (valid, expiring, expired, unknown int64, err error) {
	DB.QueryRow("SELECT COUNT(*) FROM sites WHERE cert_status = 'valid'").Scan(&valid)

	now := time.Now().Format("2006-01-02T15:04:05Z")
	thirtyDaysLater := time.Now().Add(30 * 24 * time.Hour).Format("2006-01-02T15:04:05Z")

	DB.QueryRow(
		"SELECT COUNT(*) FROM sites WHERE cert_status = 'valid' AND cert_expires_at IS NOT NULL AND cert_expires_at <= ? AND cert_expires_at > ?",
		thirtyDaysLater, now,
	).Scan(&expiring)

	DB.QueryRow(
		"SELECT COUNT(*) FROM sites WHERE cert_status = 'expired' OR (cert_expires_at IS NOT NULL AND cert_expires_at <= ?)",
		now,
	).Scan(&expired)

	DB.QueryRow("SELECT COUNT(*) FROM sites WHERE cert_status = 'unknown'").Scan(&unknown)
	return
}

func UpdateSiteCertStatus(id int64, status, expiresAt string) error {
	now := time.Now().Format(time.RFC3339)
	_, err := DB.Exec(
		"UPDATE sites SET cert_status = ?, cert_expires_at = ?, updated_at = ? WHERE id = ?",
		status, expiresAt, now, id,
	)
	return err
}

func UpdateSiteCertMode(id int64, mode string) error {
	now := time.Now().Format(time.RFC3339)
	_, err := DB.Exec(
		"UPDATE sites SET cert_mode = ?, updated_at = ? WHERE id = ?",
		mode, now, id,
	)
	return err
}

func UpdateSiteCertFiles(id int64, certPath, keyPath string) error {
	now := time.Now().Format(time.RFC3339)
	_, err := DB.Exec(
		"UPDATE sites SET cert_file_path = ?, key_file_path = ?, updated_at = ? WHERE id = ?",
		certPath, keyPath, now, id,
	)
	return err
}

func ListEnabledSites() ([]models.Site, error) {
	rows, err := DB.Query(
		`SELECT id, domain, enabled, proxy_target, proxy_config, cert_mode, cert_status,
		        cert_expires_at, cert_file_path, key_file_path, created_at, updated_at
		 FROM sites WHERE enabled = 1 ORDER BY domain`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sites []models.Site
	for rows.Next() {
		var site models.Site
		var enabled int
		var certExpiresAt, certFilePath, keyFilePath sql.NullString

		err := rows.Scan(&site.ID, &site.Domain, &enabled, &site.ProxyTarget, &site.ProxyConfig,
			&site.CertMode, &site.CertStatus, &certExpiresAt, &certFilePath, &keyFilePath,
			&site.CreatedAt, &site.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		site.Enabled = intToBool(enabled)
		site.CertExpiresAt = nullStringToPtr(certExpiresAt)
		site.CertFilePath = nullStringToPtr(certFilePath)
		site.KeyFilePath = nullStringToPtr(keyFilePath)
		sites = append(sites, site)
	}

	return sites, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func intToBool(i int) bool {
	return i == 1
}

func nullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func ptrToNullString(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}
