package system

import (
	"github.com/caddy-webui/caddy-webui/internal/models"
)

func GetSystemStatus() *models.SystemStatus {
	status := &models.SystemStatus{}

	cpu, err := GetCPUUsage()
	if err == nil {
		status.CPUUsage = &cpu
	}

	mem, err := GetMemoryInfo()
	if err == nil {
		status.MemoryUsage = &mem.UsagePercent
		status.MemoryTotalMB = &mem.TotalMB
		status.MemoryUsedMB = &mem.UsedMB
	}

	disk, err := GetDiskInfo("/")
	if err == nil {
		status.DiskUsage = &disk.UsagePercent
		status.DiskTotalGB = &disk.TotalGB
		status.DiskUsedGB = &disk.UsedGB
	}

	return status
}
