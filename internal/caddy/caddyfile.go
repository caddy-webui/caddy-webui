package caddy

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/caddy-webui/caddy-webui/internal/database"
	"github.com/caddy-webui/caddy-webui/internal/models"
)

func GenerateCaddyfile() (string, error) {
	sites, err := database.ListEnabledSites()
	if err != nil {
		return "", fmt.Errorf("获取站点列表失败: %w", err)
	}

	sort.Slice(sites, func(i, j int) bool {
		return sites[i].Domain < sites[j].Domain
	})

	var sb strings.Builder

	for _, site := range sites {
		sb.WriteString(fmt.Sprintf("%s {\n", site.Domain))
		sb.WriteString("    bind ::\n")

		if site.CertMode == "custom" && site.CertFilePath != nil && site.KeyFilePath != nil {
			sb.WriteString(fmt.Sprintf("    tls %s %s\n", *site.CertFilePath, *site.KeyFilePath))
		}

		var proxyConfig models.ProxyConfig
		if site.ProxyConfig != "" && site.ProxyConfig != "{}" {
			json.Unmarshal([]byte(site.ProxyConfig), &proxyConfig)
		}

		if site.ProxyTarget != "" {
			sb.WriteString(fmt.Sprintf("    reverse_proxy %s\n", site.ProxyTarget))
		}

		for _, route := range proxyConfig.Routes {
			sb.WriteString(fmt.Sprintf("    handle_path %s {\n", route.Path))
			if len(route.Backends) > 0 {
				sb.WriteString(fmt.Sprintf("        reverse_proxy %s\n", strings.Join(route.Backends, " ")))
			}
			for key, value := range route.Headers {
				sb.WriteString(fmt.Sprintf("        header_up %s %s\n", key, value))
			}
			sb.WriteString("    }\n")
		}

		sb.WriteString("}\n\n")
	}

	return sb.String(), nil
}

func GenerateSiteConfig(site *models.Site) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s {\n", site.Domain))
	sb.WriteString("    bind ::\n")

	if site.CertMode == "custom" && site.CertFilePath != nil && site.KeyFilePath != nil {
		sb.WriteString(fmt.Sprintf("    tls %s %s\n", *site.CertFilePath, *site.KeyFilePath))
	}

	var proxyConfig models.ProxyConfig
	if site.ProxyConfig != "" && site.ProxyConfig != "{}" {
		json.Unmarshal([]byte(site.ProxyConfig), &proxyConfig)
	}

	if site.ProxyTarget != "" {
		sb.WriteString(fmt.Sprintf("    reverse_proxy %s\n", site.ProxyTarget))
	}

	for _, route := range proxyConfig.Routes {
		sb.WriteString(fmt.Sprintf("    handle_path %s {\n", route.Path))
		if len(route.Backends) > 0 {
			sb.WriteString(fmt.Sprintf("        reverse_proxy %s\n", strings.Join(route.Backends, " ")))
		}
		for key, value := range route.Headers {
			sb.WriteString(fmt.Sprintf("        header_up %s %s\n", key, value))
		}
		sb.WriteString("    }\n")
	}

	sb.WriteString("}\n")

	return sb.String()
}
