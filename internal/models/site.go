package models

import "time"

type Admin struct {
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
}

type Site struct {
	ID            int64     `json:"id"`
	Domain        string    `json:"domain"`
	Enabled       bool      `json:"enabled"`
	ProxyTarget   string    `json:"proxy_target"`
	ProxyConfig   string    `json:"proxy_config"`
	CertMode      string    `json:"cert_mode"`
	CertStatus    string    `json:"cert_status"`
	CertExpiresAt *string   `json:"cert_expires_at"`
	CertFilePath  *string   `json:"cert_file_path"`
	KeyFilePath   *string   `json:"key_file_path"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ProxyConfig struct {
	Routes    []ProxyRoute `json:"routes"`
	WebSocket bool         `json:"websocket"`
}

type ProxyRoute struct {
	Path     string            `json:"path"`
	Backends []string          `json:"backends"`
	Headers  map[string]string `json:"headers"`
}

type SiteList struct {
	Items    []Site `json:"items"`
	Total    int64  `json:"total"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type CreateSiteRequest struct {
	Domain      string      `json:"domain"`
	Enabled     *bool       `json:"enabled"`
	ProxyTarget string      `json:"proxy_target"`
	CertMode    string      `json:"cert_mode"`
	ProxyConfig *ProxyConfig `json:"proxy_config"`
}

type UpdateSiteRequest struct {
	Domain      *string     `json:"domain"`
	Enabled     *bool       `json:"enabled"`
	ProxyTarget *string     `json:"proxy_target"`
	CertMode    *string     `json:"cert_mode"`
	ProxyConfig *ProxyConfig `json:"proxy_config"`
}

type ToggleSiteRequest struct {
	Enabled bool `json:"enabled"`
}
