package plugin

import (
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type PluginCallRecord struct {
	PluginID   string
	TenantID   int64
	DurationMs int64
	IsError    bool
	ErrorMsg   string
	MemoryKB   int64
}

type PluginRuntimeStats struct {
	PluginID        string
	TenantID        int64
	CallCount       int64
	ErrorCount      int64
	TotalDurationMs int64
	MaxDurationMs   int64
	AvgDurationMs   int64
	LastCallAt      time.Time
	LastErrorAt     *time.Time
	LastErrorMsg    string
	MemoryUsageKB   int64
	CPUPercent      float64
	Status          string
}

type ResourceAlert struct {
	PluginID     string
	TenantID     int64
	AlertType    string
	Severity     string
	Message      string
	CurrentValue int64
	Threshold    int64
	Timestamp    time.Time
}

type StatsPersister func(stats *PluginRuntimeStats)

type AlertHandler func(alert *ResourceAlert)

type ResourceThresholds struct {
	MaxDurationMs        int64
	MaxMemoryKB          int64
	MaxCPUPercent        float64
	MaxErrorRate         float64
	MaxConsecutiveErrors int64
}

func DefaultThresholds() *ResourceThresholds {
	return &ResourceThresholds{
		MaxDurationMs:        30000,
		MaxMemoryKB:          256 * 1024,
		MaxCPUPercent:        80.0,
		MaxErrorRate:         0.5,
		MaxConsecutiveErrors: 10,
	}
}

type PluginResourceMonitor struct {
	mu                sync.RWMutex
	stats             map[string]*PluginRuntimeStats
	buffer            chan PluginCallRecord
	stopCh            chan struct{}
	persister         StatsPersister
	alertHandler      AlertHandler
	thresholds        *ResourceThresholds
	consecutiveErrors map[string]int64
}

var globalMonitor *PluginResourceMonitor
var monitorOnce sync.Once

func GetResourceMonitor() *PluginResourceMonitor {
	monitorOnce.Do(func() {
		globalMonitor = &PluginResourceMonitor{
			stats:             make(map[string]*PluginRuntimeStats),
			buffer:            make(chan PluginCallRecord, 1000),
			stopCh:            make(chan struct{}),
			thresholds:        DefaultThresholds(),
			consecutiveErrors: make(map[string]int64),
		}
		go globalMonitor.processLoop()
	})
	return globalMonitor
}

func (m *PluginResourceMonitor) SetPersister(p StatsPersister) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.persister = p
}

func (m *PluginResourceMonitor) SetAlertHandler(h AlertHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alertHandler = h
}

func (m *PluginResourceMonitor) SetThresholds(t *ResourceThresholds) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t != nil {
		m.thresholds = t
	}
}

func (m *PluginResourceMonitor) RecordCall(record PluginCallRecord) {
	select {
	case m.buffer <- record:
	default:
	}
}

func (m *PluginResourceMonitor) processLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case record := <-m.buffer:
			m.updateStats(record)
		case <-ticker.C:
			m.flushToDB()
			m.checkAlerts()
		case <-m.stopCh:
			m.flushToDB()
			return
		}
	}
}

func (m *PluginResourceMonitor) updateStats(record PluginCallRecord) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := pluginKey(record.TenantID, record.PluginID)
	s, exists := m.stats[key]
	if !exists {
		s = &PluginRuntimeStats{
			PluginID: record.PluginID,
			TenantID: record.TenantID,
			Status:   "active",
		}
		m.stats[key] = s
	}

	atomic.AddInt64(&s.CallCount, 1)
	atomic.AddInt64(&s.TotalDurationMs, record.DurationMs)
	s.LastCallAt = time.Now()

	if record.DurationMs > s.MaxDurationMs {
		s.MaxDurationMs = record.DurationMs
	}

	count := atomic.LoadInt64(&s.CallCount)
	if count > 0 {
		total := atomic.LoadInt64(&s.TotalDurationMs)
		s.AvgDurationMs = total / count
	}

	if record.IsError {
		atomic.AddInt64(&s.ErrorCount, 1)
		now := time.Now()
		s.LastErrorAt = &now
		s.LastErrorMsg = record.ErrorMsg

		consecutive := m.consecutiveErrors[key] + 1
		m.consecutiveErrors[key] = consecutive
	} else {
		m.consecutiveErrors[key] = 0
	}

	if record.MemoryKB > 0 {
		s.MemoryUsageKB = record.MemoryKB
	}

	s.CPUPercent = float64(runtime.NumGoroutine())
}

func (m *PluginResourceMonitor) flushToDB() {
	m.mu.RLock()
	snapshot := make(map[string]*PluginRuntimeStats, len(m.stats))
	for k, v := range m.stats {
		snapshot[k] = &PluginRuntimeStats{
			PluginID:        v.PluginID,
			TenantID:        v.TenantID,
			CallCount:       atomic.LoadInt64(&v.CallCount),
			ErrorCount:      atomic.LoadInt64(&v.ErrorCount),
			TotalDurationMs: atomic.LoadInt64(&v.TotalDurationMs),
			MaxDurationMs:   v.MaxDurationMs,
			AvgDurationMs:   v.AvgDurationMs,
			LastCallAt:      v.LastCallAt,
			LastErrorAt:     v.LastErrorAt,
			LastErrorMsg:    v.LastErrorMsg,
			MemoryUsageKB:   v.MemoryUsageKB,
			CPUPercent:      v.CPUPercent,
			Status:          v.Status,
		}
	}
	m.mu.RUnlock()

	for _, s := range snapshot {
		if s.CallCount == 0 {
			continue
		}
		m.persistStats(s)
	}
}

func (m *PluginResourceMonitor) persistStats(s *PluginRuntimeStats) {
	if m.persister != nil {
		m.persister(s)
	}
}

func (m *PluginResourceMonitor) checkAlerts() {
	m.mu.RLock()
	thresholds := m.thresholds
	handler := m.alertHandler
	snapshot := make(map[string]*PluginRuntimeStats, len(m.stats))
	for k, v := range m.stats {
		snapshot[k] = &PluginRuntimeStats{
			PluginID:        v.PluginID,
			TenantID:        v.TenantID,
			CallCount:       atomic.LoadInt64(&v.CallCount),
			ErrorCount:      atomic.LoadInt64(&v.ErrorCount),
			TotalDurationMs: atomic.LoadInt64(&v.TotalDurationMs),
			MaxDurationMs:   v.MaxDurationMs,
			AvgDurationMs:   v.AvgDurationMs,
			LastCallAt:      v.LastCallAt,
			LastErrorAt:     v.LastErrorAt,
			LastErrorMsg:    v.LastErrorMsg,
			MemoryUsageKB:   v.MemoryUsageKB,
			CPUPercent:      v.CPUPercent,
			Status:          v.Status,
		}
	}
	consecutiveErrors := make(map[string]int64, len(m.consecutiveErrors))
	for k, v := range m.consecutiveErrors {
		consecutiveErrors[k] = v
	}
	m.mu.RUnlock()

	if handler == nil || thresholds == nil {
		return
	}

	for key, s := range snapshot {
		if s.MaxDurationMs > thresholds.MaxDurationMs {
			handler(&ResourceAlert{
				PluginID:     s.PluginID,
				TenantID:     s.TenantID,
				AlertType:    "high_latency",
				Severity:     "warning",
				Message:      "插件调用延迟过高",
				CurrentValue: s.MaxDurationMs,
				Threshold:    thresholds.MaxDurationMs,
				Timestamp:    time.Now(),
			})
		}

		if s.MemoryUsageKB > thresholds.MaxMemoryKB {
			handler(&ResourceAlert{
				PluginID:     s.PluginID,
				TenantID:     s.TenantID,
				AlertType:    "high_memory",
				Severity:     "critical",
				Message:      "插件内存使用超限",
				CurrentValue: s.MemoryUsageKB,
				Threshold:    thresholds.MaxMemoryKB,
				Timestamp:    time.Now(),
			})
		}

		if s.CPUPercent > thresholds.MaxCPUPercent {
			handler(&ResourceAlert{
				PluginID:     s.PluginID,
				TenantID:     s.TenantID,
				AlertType:    "high_cpu",
				Severity:     "warning",
				Message:      "插件CPU使用率过高",
				CurrentValue: int64(s.CPUPercent),
				Threshold:    int64(thresholds.MaxCPUPercent),
				Timestamp:    time.Now(),
			})
		}

		if s.CallCount > 0 {
			errorRate := float64(s.ErrorCount) / float64(s.CallCount)
			if errorRate > thresholds.MaxErrorRate {
				handler(&ResourceAlert{
					PluginID:     s.PluginID,
					TenantID:     s.TenantID,
					AlertType:    "high_error_rate",
					Severity:     "critical",
					Message:      "插件错误率过高",
					CurrentValue: int64(errorRate * 100),
					Threshold:    int64(thresholds.MaxErrorRate * 100),
					Timestamp:    time.Now(),
				})
			}
		}

		if consecutive, ok := consecutiveErrors[key]; ok && consecutive >= thresholds.MaxConsecutiveErrors {
			handler(&ResourceAlert{
				PluginID:     s.PluginID,
				TenantID:     s.TenantID,
				AlertType:    "consecutive_errors",
				Severity:     "critical",
				Message:      "插件连续错误次数过多",
				CurrentValue: consecutive,
				Threshold:    thresholds.MaxConsecutiveErrors,
				Timestamp:    time.Now(),
			})
		}
	}
}

func (m *PluginResourceMonitor) GetStats(tenantID int64, pluginID string) *PluginRuntimeStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := pluginKey(tenantID, pluginID)
	s, exists := m.stats[key]
	if !exists {
		return nil
	}

	return &PluginRuntimeStats{
		PluginID:        s.PluginID,
		TenantID:        s.TenantID,
		CallCount:       atomic.LoadInt64(&s.CallCount),
		ErrorCount:      atomic.LoadInt64(&s.ErrorCount),
		TotalDurationMs: atomic.LoadInt64(&s.TotalDurationMs),
		MaxDurationMs:   s.MaxDurationMs,
		AvgDurationMs:   s.AvgDurationMs,
		LastCallAt:      s.LastCallAt,
		LastErrorAt:     s.LastErrorAt,
		LastErrorMsg:    s.LastErrorMsg,
		MemoryUsageKB:   s.MemoryUsageKB,
		CPUPercent:      s.CPUPercent,
		Status:          s.Status,
	}
}

func (m *PluginResourceMonitor) GetAllStats(tenantID int64) []*PluginRuntimeStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	prefix := pluginKey(tenantID, "")
	result := make([]*PluginRuntimeStats, 0)

	for key, s := range m.stats {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			result = append(result, &PluginRuntimeStats{
				PluginID:        s.PluginID,
				TenantID:        s.TenantID,
				CallCount:       atomic.LoadInt64(&s.CallCount),
				ErrorCount:      atomic.LoadInt64(&s.ErrorCount),
				TotalDurationMs: atomic.LoadInt64(&s.TotalDurationMs),
				MaxDurationMs:   s.MaxDurationMs,
				AvgDurationMs:   s.AvgDurationMs,
				LastCallAt:      s.LastCallAt,
				LastErrorAt:     s.LastErrorAt,
				LastErrorMsg:    s.LastErrorMsg,
				MemoryUsageKB:   s.MemoryUsageKB,
				CPUPercent:      s.CPUPercent,
				Status:          s.Status,
			})
		}
	}
	return result
}

func (m *PluginResourceMonitor) ResetStats(tenantID int64, pluginID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := pluginKey(tenantID, pluginID)
	if s, exists := m.stats[key]; exists {
		atomic.StoreInt64(&s.CallCount, 0)
		atomic.StoreInt64(&s.ErrorCount, 0)
		atomic.StoreInt64(&s.TotalDurationMs, 0)
		s.MaxDurationMs = 0
		s.AvgDurationMs = 0
		s.MemoryUsageKB = 0
		s.CPUPercent = 0
	}
	delete(m.consecutiveErrors, key)
}

func (m *PluginResourceMonitor) Stop() {
	close(m.stopCh)
}

func init() {
	log.Println("[PluginResourceMonitor] 资源监控模块已初始化")
}
