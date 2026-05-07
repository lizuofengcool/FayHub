//go:build linux || darwin

package metrics

import (
	"syscall"
)

func getDiskUsage() map[string]interface{} {
	var total, free uint64
	var stat syscall.Statfs_t
	if err := syscall.Statfs(".", &stat); err == nil {
		total = stat.Blocks * uint64(stat.Bsize)
		free = stat.Bavail * uint64(stat.Bsize)
	}
	used := total - free
	usagePercent := float64(0)
	if total > 0 {
		usagePercent = float64(used) / float64(total) * 100
	}
	return map[string]interface{}{
		"total_gb":      float64(total) / 1024 / 1024 / 1024,
		"used_gb":       float64(used) / 1024 / 1024 / 1024,
		"free_gb":       float64(free) / 1024 / 1024 / 1024,
		"usage_percent": usagePercent,
	}
}
