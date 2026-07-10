package models

type DashboardData struct {
	System       SystemStatus     `json:"system"`
	Caddy        CaddyStatus      `json:"caddy"`
	Sites        SitesStats       `json:"sites"`
	Certificates CertStats        `json:"certificates"`
}

type SystemStatus struct {
	CPUUsage     *float64 `json:"cpu_usage"`
	MemoryUsage  *float64 `json:"memory_usage"`
	DiskUsage    *float64 `json:"disk_usage"`
	MemoryTotalMB *float64 `json:"memory_total_mb"`
	MemoryUsedMB  *float64 `json:"memory_used_mb"`
	DiskTotalGB   *float64 `json:"disk_total_gb"`
	DiskUsedGB    *float64 `json:"disk_used_gb"`
}

type CaddyStatus struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

type SitesStats struct {
	Total    int64 `json:"total"`
	Enabled  int64 `json:"enabled"`
	Disabled int64 `json:"disabled"`
}

type CertStats struct {
	Valid        int64 `json:"valid"`
	ExpiringSoon int64 `json:"expiring_soon"`
	Expired      int64 `json:"expired"`
	Unknown      int64 `json:"unknown"`
}
