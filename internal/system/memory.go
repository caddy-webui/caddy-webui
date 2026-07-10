package system

import (
	"os"
	"strconv"
	"strings"
)

type MemoryInfo struct {
	UsagePercent float64
	TotalMB      float64
	UsedMB       float64
}

func GetMemoryInfo() (*MemoryInfo, error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return nil, err
	}

	var memTotal, memAvailable uint64
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		value, _ := strconv.ParseUint(fields[1], 10, 64)

		switch fields[0] {
		case "MemTotal:":
			memTotal = value
		case "MemAvailable:":
			memAvailable = value
		}
	}

	if memTotal == 0 {
		return nil, err
	}

	totalMB := float64(memTotal) / 1024
	usedMB := float64(memTotal-memAvailable) / 1024
	usagePercent := usedMB / totalMB * 100

	return &MemoryInfo{
		UsagePercent: usagePercent,
		TotalMB:      totalMB,
		UsedMB:       usedMB,
	}, nil
}
