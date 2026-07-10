package system

import (
	"os"
	"strconv"
	"strings"
	"time"
)

func GetCPUUsage() (float64, error) {
	stat1, err := readCPUStats()
	if err != nil {
		return 0, err
	}

	time.Sleep(1 * time.Second)

	stat2, err := readCPUStats()
	if err != nil {
		return 0, err
	}

	totalDiff := stat2.total - stat1.total
	idleDiff := stat2.idle - stat1.idle

	if totalDiff == 0 {
		return 0, nil
	}

	usage := float64(totalDiff-idleDiff) / float64(totalDiff) * 100
	return usage, nil
}

type cpuStats struct {
	total uint64
	idle  uint64
}

func readCPUStats() (*cpuStats, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return nil, err
	}

	line := strings.Split(string(data), "\n")[0]
	fields := strings.Fields(line)

	if len(fields) < 5 {
		return nil, err
	}

	var total, idle uint64
	for i := 1; i < len(fields); i++ {
		v, _ := strconv.ParseUint(fields[i], 10, 64)
		total += v
		if i == 4 {
			idle = v
		}
	}

	return &cpuStats{total: total, idle: idle}, nil
}
