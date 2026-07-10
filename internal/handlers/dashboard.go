package handlers

import (
	"net/http"
	"strconv"

	"github.com/caddy-webui/caddy-webui/internal/caddy"
	"github.com/caddy-webui/caddy-webui/internal/database"
	"github.com/caddy-webui/caddy-webui/internal/models"
	"github.com/caddy-webui/caddy-webui/internal/system"
)

func HandleDashboard(w http.ResponseWriter, r *http.Request) {
	sysStatus := system.GetSystemStatus()

	caddyStatus, _ := caddy.GetCaddyStatus()
	caddyVersion := caddy.GetCaddyVersion()

	total, enabled, disabled, _ := database.CountSitesByStatus()

	valid, expiring, expired, unknown, _ := database.CountCertificatesByStatus()

	data := models.DashboardData{
		System: *sysStatus,
		Caddy: models.CaddyStatus{
			Status:  caddyStatus,
			Version: caddyVersion,
		},
		Sites: models.SitesStats{
			Total:    total,
			Enabled:  enabled,
			Disabled: disabled,
		},
		Certificates: models.CertStats{
			Valid:        valid,
			ExpiringSoon: expiring,
			Expired:      expired,
			Unknown:      unknown,
		},
	}

	SuccessResponse(w, "success", data)
}

func HandleGetSites(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	sites, total, err := database.ListSites(page, pageSize)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "获取站点列表失败")
		return
	}

	if sites == nil {
		sites = []models.Site{}
	}

	SuccessResponse(w, "success", models.SiteList{
		Items:    sites,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
