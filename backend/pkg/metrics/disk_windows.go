//go:build windows

package metrics

import (
	"os"
	"syscall"
	"unsafe"
)

func getDiskUsage() map[string]interface{} {
	dir, err := os.Getwd()
	if err != nil {
		dir = "C:"
	}
	var free, total, avail uint64
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpaceEx := kernel32.NewProc("GetDiskFreeSpaceExW")
	ptr, _ := syscall.UTF16PtrFromString(dir)
	ret, _, _ := getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(ptr)),
		uintptr(unsafe.Pointer(&free)),
		uintptr(unsafe.Pointer(&total)),
		uintptr(unsafe.Pointer(&avail)),
	)
	_ = ret
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
