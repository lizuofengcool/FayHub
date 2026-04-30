package metrics

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type APIMetrics struct {
	Path       string
	Method     string
	TotalCount int64
	ErrorCount int64
	TotalDur   int64
	MinDur     int64
	MaxDur     int64
}

type SystemMetrics struct {
	StartTime       time.Time
	TotalRequests   int64
	ActiveRequests  int64
	ErrorRequests   int64
	GoroutineCount  int
	MemoryAlloc     uint64
	MemorySys       uint64
	GCPauseTotal    time.Duration
	APIMap          map[string]*APIMetrics
	mu              sync.RWMutex
}

var globalMetrics = &SystemMetrics{
	StartTime: time.Now(),
	APIMap:    make(map[string]*APIMetrics),
}

func RecordRequest(method, path string, duration time.Duration, statusCode int) {
	atomic.AddInt64(&globalMetrics.TotalRequests, 1)

	if statusCode >= 400 {
		atomic.AddInt64(&globalMetrics.ErrorRequests, 1)
	}

	key := fmt.Sprintf("%s:%s", method, path)
	durMs := duration.Milliseconds()

	globalMetrics.mu.Lock()
	m, exists := globalMetrics.APIMap[key]
	if !exists {
		m = &APIMetrics{
			Path:   path,
			Method: method,
			MinDur: durMs,
			MaxDur: durMs,
		}
		globalMetrics.APIMap[key] = m
	}
	atomic.AddInt64(&m.TotalCount, 1)
	atomic.AddInt64(&m.TotalDur, durMs)
	if durMs < m.MinDur || m.MinDur == 0 {
		m.MinDur = durMs
	}
	if durMs > m.MaxDur {
		m.MaxDur = durMs
	}
	if statusCode >= 400 {
		atomic.AddInt64(&m.ErrorCount, 1)
	}
	globalMetrics.mu.Unlock()
}

func IncrementActiveRequests() {
	atomic.AddInt64(&globalMetrics.ActiveRequests, 1)
}

func DecrementActiveRequests() {
	atomic.AddInt64(&globalMetrics.ActiveRequests, -1)
}

func GetMetrics() map[string]interface{} {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	globalMetrics.mu.RLock()
	apiMetrics := make([]map[string]interface{}, 0, len(globalMetrics.APIMap))
	for _, m := range globalMetrics.APIMap {
		avgDur := int64(0)
		if m.TotalCount > 0 {
			avgDur = m.TotalDur / m.TotalCount
		}
		apiMetrics = append(apiMetrics, map[string]interface{}{
			"method":      m.Method,
			"path":        m.Path,
			"total_count": m.TotalCount,
			"error_count": m.ErrorCount,
			"avg_dur_ms":  avgDur,
			"min_dur_ms":  m.MinDur,
			"max_dur_ms":  m.MaxDur,
		})
	}
	globalMetrics.mu.RUnlock()

	uptime := time.Since(globalMetrics.StartTime)

	return map[string]interface{}{
		"uptime_seconds":    uptime.Seconds(),
		"total_requests":    atomic.LoadInt64(&globalMetrics.TotalRequests),
		"active_requests":   atomic.LoadInt64(&globalMetrics.ActiveRequests),
		"error_requests":    atomic.LoadInt64(&globalMetrics.ErrorRequests),
		"goroutine_count":   runtime.NumGoroutine(),
		"memory_alloc_mb":   float64(memStats.Alloc) / 1024 / 1024,
		"memory_sys_mb":     float64(memStats.Sys) / 1024 / 1024,
		"gc_pause_total_ms": float64(memStats.PauseTotalNs) / 1e6,
		"gc_count":          memStats.NumGC,
		"api_metrics":       apiMetrics,
	}
}

func GetPrometheusFormat() string {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	result := ""
	result += fmt.Sprintf("# HELP fayhub_requests_total Total number of requests\n")
	result += fmt.Sprintf("# TYPE fayhub_requests_total counter\n")
	result += fmt.Sprintf("fayhub_requests_total %d\n", atomic.LoadInt64(&globalMetrics.TotalRequests))

	result += fmt.Sprintf("# HELP fayhub_requests_active Currently active requests\n")
	result += fmt.Sprintf("# TYPE fayhub_requests_active gauge\n")
	result += fmt.Sprintf("fayhub_requests_active %d\n", atomic.LoadInt64(&globalMetrics.ActiveRequests))

	result += fmt.Sprintf("# HELP fayhub_requests_errors_total Total error requests\n")
	result += fmt.Sprintf("# TYPE fayhub_requests_errors_total counter\n")
	result += fmt.Sprintf("fayhub_requests_errors_total %d\n", atomic.LoadInt64(&globalMetrics.ErrorRequests))

	result += fmt.Sprintf("# HELP fayhub_goroutines Number of goroutines\n")
	result += fmt.Sprintf("# TYPE fayhub_goroutines gauge\n")
	result += fmt.Sprintf("fayhub_goroutines %d\n", runtime.NumGoroutine())

	result += fmt.Sprintf("# HELP fayhub_memory_alloc_bytes Allocated memory in bytes\n")
	result += fmt.Sprintf("# TYPE fayhub_memory_alloc_bytes gauge\n")
	result += fmt.Sprintf("fayhub_memory_alloc_bytes %d\n", memStats.Alloc)

	result += fmt.Sprintf("# HELP fayhub_memory_sys_bytes System memory in bytes\n")
	result += fmt.Sprintf("# TYPE fayhub_memory_sys_bytes gauge\n")
	result += fmt.Sprintf("fayhub_memory_sys_bytes %d\n", memStats.Sys)

	result += fmt.Sprintf("# HELP fayhub_uptime_seconds Service uptime in seconds\n")
	result += fmt.Sprintf("# TYPE fayhub_uptime_seconds gauge\n")
	result += fmt.Sprintf("fayhub_uptime_seconds %.0f\n", time.Since(globalMetrics.StartTime).Seconds())

	globalMetrics.mu.RLock()
	for _, m := range globalMetrics.APIMap {
		avgDur := int64(0)
		if m.TotalCount > 0 {
			avgDur = m.TotalDur / m.TotalCount
		}
		result += fmt.Sprintf("fayhub_api_duration_avg_ms{method=\"%s\",path=\"%s\"} %d\n", m.Method, m.Path, avgDur)
		result += fmt.Sprintf("fayhub_api_requests_total{method=\"%s\",path=\"%s\"} %d\n", m.Method, m.Path, m.TotalCount)
		result += fmt.Sprintf("fayhub_api_errors_total{method=\"%s\",path=\"%s\"} %d\n", m.Method, m.Path, m.ErrorCount)
	}
	globalMetrics.mu.RUnlock()

	return result
}
