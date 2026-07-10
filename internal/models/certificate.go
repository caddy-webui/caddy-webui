package models

type CertInfo struct {
	SiteID    int64   `json:"site_id"`
	Domain    string  `json:"domain"`
	CertMode  string  `json:"cert_mode"`
	Status    string  `json:"status"`
	ExpiresAt *string `json:"expires_at"`
	Issuer    string  `json:"issuer"`
}

type CertModeRequest struct {
	CertMode string `json:"cert_mode"`
}
