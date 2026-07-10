package system

import (
	"syscall"
)

type DiskInfo struct {
	UsagePercent float64
	TotalGB      float64
	UsedGB       float64
}

func GetDiskInfo(path string) (*DiskInfo, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return nil, err
	}

	totalBytes := stat.Blocks * uint64(stat.Bsize)
	freeBytes := stat.Bavail * uint64(stat.Bsize)
	usedBytes := totalBytes - freeBytes

	totalGB := float64(totalBytes) / 1024 / 1024 / 1024
	usedGB := float64(usedBytes) / 1024 / 1024 / 1024

	var usagePercent float64
	if totalBytes > 0 {
		usagePercent = float64(usedBytes) / float64(totalBytes) * 100
	}

	return &DiskInfo{
		UsagePercent: usagePercent,
		TotalGB:      totalGB,
		UsedGB:       usedGB,
	}, nil
}
